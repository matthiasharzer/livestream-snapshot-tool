package run

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/matthiasharzer/livebuffer/api/clip"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/stream"
	"github.com/matthiasharzer/livebuffer/util/fsutil"
	"github.com/spf13/cobra"
)

var streamURLString string
var buffer time.Duration
var httpPort int
var httpHost string
var cookiesFile string
var bufferDirectoryArg string
var resumeBuffer bool
var restartOnFailure bool

func init() {
	Command.Flags().StringVarP(&streamURLString, "url", "u", "", "URL of the livestream to snapshot (required)")
	err := Command.MarkFlagRequired("url")
	if err != nil {
		panic(err)
	}

	Command.Flags().DurationVarP(&buffer, "buffer", "b", time.Minute*10, "Duration of the live buffer (default: 10m)")
	Command.Flags().StringVarP(&bufferDirectoryArg, "buffer-dir", "", "", "Directory to store live buffer segments (default: temporary directory)")
	Command.Flags().BoolVarP(&resumeBuffer, "resume-buffer", "", false, "Whether to use existing buffer files in the buffer directory (only applicable if --buffer-dir is set)")
	Command.Flags().IntVarP(&httpPort, "port", "p", 4000, "HTTP server port")
	Command.Flags().StringVarP(&httpHost, "host", "", "", "HTTP server host (default: all interfaces)")
	Command.Flags().StringVarP(&cookiesFile, "cookies-file", "", "", "Path to a file containing cookies for yt-dlp")
	Command.Flags().BoolVarP(&restartOnFailure, "restart-on-failure", "", false, "Whether to automatically restart the live buffer if it fails (e.g. due to stream interruptions)")
}

var Command = &cobra.Command{
	Use:   "run",
	Short: "Run the livebuffer server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if buffer <= 0 {
			return errors.New("buffer duration must be greater than 0")
		}
		if httpPort <= 0 || httpPort > 65535 {
			return errors.New("port must be a valid TCP port number")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		streamURL, err := url.Parse(streamURLString)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}

		var bufferDirectory string
		if bufferDirectoryArg != "" {
			bufferDirectory = bufferDirectoryArg
			err := os.MkdirAll(bufferDirectory, 0755)
			if err != nil {
				return fmt.Errorf("failed to create buffer directory: %w", err)
			}
		} else {
			tmpDir, cleanup, err := fsutil.TemporaryDirectory()
			if err != nil {
				return err
			}
			defer cleanup()
			bufferDirectory = tmpDir
		}

		logging.Info("starting live buffer", "url", streamURLString, "buffer", buffer.String(), "bufferDirectory", bufferDirectory)
		liveBuffer := stream.NewLiveBuffer(streamURL.String(), buffer, bufferDirectory, resumeBuffer, cookiesFile)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			for {
				err := liveBuffer.Start(ctx)
				if err != nil {
					logging.Error("live buffer error", "err", err)
				} else {
					logging.Info("live buffer stopped")
				}
				if !restartOnFailure {
					os.Exit(1)
					return
				}
				logging.Info("restarting live buffer in 5 seconds...")
				time.Sleep(5 * time.Second)
			}
		}()
		defer liveBuffer.Stop()

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
