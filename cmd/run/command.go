package run

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/matthiasharzer/livestream-snapshotting-tool/api/clip"
	"github.com/matthiasharzer/livestream-snapshotting-tool/logging"
	"github.com/matthiasharzer/livestream-snapshotting-tool/showmaster"
	"github.com/matthiasharzer/livestream-snapshotting-tool/stream"
	"github.com/matthiasharzer/livestream-snapshotting-tool/util/fsutil"
	"github.com/spf13/cobra"
)

var streamURLString string
var bufferMinutes int
var httpPort int
var httpHost string
var historySize int
var cookiesFile string

func init() {
	Command.Flags().StringVarP(&streamURLString, "url", "u", "", "URL of the livestream to snapshot (required)")
	err := Command.MarkFlagRequired("url")
	if err != nil {
		panic(err)
	}

	Command.Flags().IntVarP(&bufferMinutes, "buffer", "b", 10, "Duration of the live buffer in minutes")
	Command.Flags().IntVarP(&httpPort, "port", "p", 4000, "HTTP server port")
	Command.Flags().StringVarP(&httpHost, "host", "", "", "HTTP server host (default: all interfaces)")
	Command.Flags().IntVarP(&historySize, "history-size", "", 1, "Number of historical clips to keep")
	Command.Flags().StringVarP(&cookiesFile, "cookies-file", "", "", "Path to a file containing cookies for yt-dlp")
}

var Command = &cobra.Command{
	Use:   "run",
	Short: "Run the livestream snapshotting server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if bufferMinutes <= 0 {
			return errors.New("buffer duration must be a positive integer")
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
		bufferDuration := time.Minute * time.Duration(bufferMinutes)
		streamURL, err := url.Parse(streamURLString)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}

		outDir, cleanup, err := fsutil.TemporaryDirectory()
		if err != nil {
			return err
		}
		defer cleanup()

		outDir = ".tmp"
		os.MkdirAll(outDir, 0755)

		_, err = showmaster.New(historySize)
		if err != nil {
			return fmt.Errorf("failed to create show master: %w", err)
		}

		logging.Info("starting live buffer", "url", streamURLString, "bufferDuration", bufferDuration.String(), "outputDir", outDir)
		liveBuffer := stream.NewLiveBuffer(streamURL.String(), bufferDuration, outDir, true, cookiesFile)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = liveBuffer.Start(ctx)
		if err != nil {
			return fmt.Errorf("failed to start live buffer: %w", err)
		}
		defer liveBuffer.Stop()

		//onSegment := func(filePath string, err error) {
		//	if err != nil {
		//		logging.Error("error processing segment", "err", err)
		//		return
		//	}
		//	err = master.AddClip(filePath)
		//	if err != nil {
		//		logging.Error("failed to add clip to master", "err", err)
		//	}
		//	logging.Info("clip added to master", "filePath", filePath)
		//}

		//ripper := stream.NewRipper(*streamURL, bufferDuration, outDir, onSegment, cookiesFile)
		//err = ripper.Start()
		//if err != nil {
		//	return fmt.Errorf("failed to start ripper: %w", err)
		//}
		//defer ripper.Stop()

		addr := fmt.Sprintf("%s:%d", httpHost, httpPort)

		logging.Info("starting livestream snapshot server", "host", httpHost, "port", httpPort)
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/clip", clip.Handler(liveBuffer))

		err = http.ListenAndServe(addr, mux)
		if err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}

		return nil
	},
}
