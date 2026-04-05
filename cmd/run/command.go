package run

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/matthiasharzer/livestream-snapshot-tool/api/latestclip"
	"github.com/matthiasharzer/livestream-snapshot-tool/logging"
	"github.com/matthiasharzer/livestream-snapshot-tool/showmaster"
	"github.com/matthiasharzer/livestream-snapshot-tool/stream"
	"github.com/matthiasharzer/livestream-snapshot-tool/util/fsutil"
	"github.com/spf13/cobra"
)

var streamURLString string
var intervalMinutes int
var httpPort int
var httpHost string

func init() {
	Command.Flags().StringVarP(&streamURLString, "url", "u", "", "URL of the livestream to snapshot (required)")
	err := Command.MarkFlagRequired("url")
	if err != nil {
		panic(err)
	}

	Command.Flags().IntVarP(&intervalMinutes, "interval", "i", 10, "Interval in minutes between snapshots")
	Command.Flags().IntVarP(&httpPort, "port", "p", 8080, "HTTP server port")
	Command.Flags().StringVarP(&httpHost, "host", "", "", "HTTP server host (default: all interfaces)")
}

var Command = &cobra.Command{
	Use:   "run",
	Short: "Run the livestream snapshot server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if intervalMinutes <= 0 {
			return errors.New("interval must be a positive integer")
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

		master := showmaster.New()
		master.LatestClip.SetPath("daniel.mp4")

		onSegment := func(filePath string, err error) {
			if err != nil {
				logging.Error("error processing segment: %v", err)
				return
			}
			master.LatestClip.SetPath(filePath)
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
		mux.HandleFunc("GET /api/v1/latest", latestclip.Handler(master))

		err = http.ListenAndServe(addr, mux)
		if err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}

		return nil
	},
}
