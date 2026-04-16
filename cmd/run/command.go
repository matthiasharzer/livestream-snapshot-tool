package run

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/matthiasharzer/livestream-snapshotting-tool/api/clip"
	"github.com/matthiasharzer/livestream-snapshotting-tool/logging"
	"github.com/matthiasharzer/livestream-snapshotting-tool/showmaster"
	"github.com/matthiasharzer/livestream-snapshotting-tool/stream"
	"github.com/matthiasharzer/livestream-snapshotting-tool/util/fsutil"
	"github.com/spf13/cobra"
)

var streamURLString string
var intervalMinutes int
var httpPort int
var httpHost string
var historySize int

func init() {
	Command.Flags().StringVarP(&streamURLString, "url", "u", "", "URL of the livestream to snapshot (required)")
	err := Command.MarkFlagRequired("url")
	if err != nil {
		panic(err)
	}

	Command.Flags().IntVarP(&intervalMinutes, "interval", "i", 10, "Interval in minutes between snapshots")
	Command.Flags().IntVarP(&httpPort, "port", "p", 8080, "HTTP server port")
	Command.Flags().StringVarP(&httpHost, "host", "", "", "HTTP server host (default: all interfaces)")
	Command.Flags().IntVarP(&historySize, "history-size", "", 1, "Number of historical clips to keep")
}

var Command = &cobra.Command{
	Use:   "run",
	Short: "Run the livestream snapshotting server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if intervalMinutes <= 0 {
			return errors.New("interval must be a positive integer")
		}
		if httpPort <= 0 || httpPort > 65535 {
			return errors.New("port must be a valid TCP port number")
		}
		if historySize < 1 {
			return errors.New("history size must be at least 1")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		interval := time.Minute * time.Duration(intervalMinutes)
		streamURL, err := url.Parse(streamURLString)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}

		outDir, cleanup, err := fsutil.TemporaryDirectory()
		if err != nil {
			return err
		}
		defer cleanup()

		master := showmaster.New(historySize)

		onSegment := func(filePath string, err error) {
			if err != nil {
				logging.Error("error processing segment", "err", err)
				return
			}
			err = master.AddClip(filePath)
			if err != nil {
				logging.Error("failed to add clip to master", "err", err)
			}
			logging.Info("clip added to master", "filePath", filePath)
		}

		ripper := stream.NewRipper(*streamURL, interval, outDir, onSegment)
		err = ripper.Start()
		if err != nil {
			return fmt.Errorf("failed to start ripper: %w", err)
		}
		defer ripper.Stop()

		addr := fmt.Sprintf("%s:%d", httpHost, httpPort)

		logging.Info("starting livestream snapshot server", "host", httpHost, "port", httpPort)
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/clip/{clip}", clip.Handler(master))

		err = http.ListenAndServe(addr, mux)
		if err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}

		return nil
	},
}
