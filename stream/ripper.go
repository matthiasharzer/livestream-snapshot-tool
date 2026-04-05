package stream

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// SegmentCallback is the function signature for handling completed downloads.
// It provides the path to the downloaded file, or an error if the segment failed.
type SegmentCallback func(filePath string, err error)

// Ripper manages the asynchronous downloading of livestream segments.
type Ripper struct {
	URL       url.URL
	Interval  time.Duration
	OutputDir string
	Callback  SegmentCallback

	mu         sync.Mutex
	cancelFunc context.CancelFunc
	isRunning  bool
}

func NewRipper(url url.URL, interval time.Duration, outDir string, cb SegmentCallback) *Ripper {
	return &Ripper{
		URL:       url,
		Interval:  interval,
		OutputDir: outDir,
		Callback:  cb,
	}
}

func (r *Ripper) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isRunning {
		return fmt.Errorf("ripper is already running")
	}

	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFunc = cancel
	r.isRunning = true

	go r.loop(ctx)

	return nil
}

func (r *Ripper) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isRunning && r.cancelFunc != nil {
		r.cancelFunc()
		r.isRunning = false
	}
}

func (r *Ripper) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		timestamp := time.Now().Format("20060102_150405")
		outPath := filepath.Join(r.OutputDir, fmt.Sprintf("segment_%s.mp4", timestamp))
		seconds := int(r.Interval.Seconds())

		cmd := exec.CommandContext(ctx, "yt-dlp",
			r.URL.String(),
			"--downloader", "ffmpeg",
			"--downloader-args", fmt.Sprintf("ffmpeg:-t %d", seconds),
			"--merge-output-format", "mp4",
			"-o", outPath,
		)

		err := cmd.Run()

		if ctx.Err() != nil {
			return
		}

		if err == nil {
			if _, statErr := os.Stat(outPath); os.IsNotExist(statErr) {
				err = fmt.Errorf("yt-dlp succeeded but output file %s is missing", outPath)
			}
		}

		go r.Callback(outPath, err)
	}
}
