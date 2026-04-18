package stream

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matthiasharzer/livestream-snapshotting-tool/logging"
	"github.com/matthiasharzer/livestream-snapshotting-tool/util/fsutil"
)

const DiskRetentionMargin = 5 * time.Minute

type hlsSegment struct {
	filename string
	duration time.Duration
}

// LiveBuffer manages a rolling HLS buffer of a livestream.
type LiveBuffer struct {
	url             string
	bufferDuration  time.Duration
	segmentDuration time.Duration
	outputDir       string
	resume          bool
	cookieFile      string

	mu          sync.Mutex // Protects state
	ytCmd       *exec.Cmd
	ffmpegCmd   *exec.Cmd
	ffmpegStdin io.WriteCloser
	stopped     chan struct{} // Used to block until fully shut down
	isRunning   bool
}

func NewLiveBuffer(streamURL string, bufferDuration time.Duration, bufferDirectory string, keepOldFiles bool, cookieFile string) *LiveBuffer {
	return &LiveBuffer{
		url:             streamURL,
		bufferDuration:  bufferDuration,
		segmentDuration: 60 * time.Second,
		outputDir:       bufferDirectory,
		resume:          keepOldFiles,
		cookieFile:      cookieFile,
	}
}

func (b *LiveBuffer) playlistFilePath() string {
	return filepath.Join(b.outputDir, "live.m3u8")
}

func (b *LiveBuffer) clearOutputDir() error {
	files, err := os.ReadDir(b.outputDir)
	if err != nil {
		return err
	}
	err = os.Remove(b.playlistFilePath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".ts") {
			continue
		}
		err := os.Remove(filepath.Join(b.outputDir, file.Name()))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// Start begins capturing the stream. It runs asynchronously until ctx is canceled.
func (b *LiveBuffer) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.resume {
		err := b.clearOutputDir()
		if err != nil {
			logging.Warn("failed to clear old buffer dir", "err", err)
		}
	}
	if err := os.MkdirAll(b.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create buffer dir: %w", err)
	}

	physicalDuration := b.bufferDuration + DiskRetentionMargin
	listSize := int(physicalDuration.Seconds() / b.segmentDuration.Seconds())

	b.ytCmd = exec.CommandContext(ctx, "yt-dlp", "-q", "-o", "-", b.url)
	if b.cookieFile != "" {
		b.ytCmd.Args = append(b.ytCmd.Args, "--cookies", b.cookieFile)
	}

	hlsFlags := "delete_segments"
	if b.resume {
		// append_list tells ffmpeg to read the existing .m3u8 and continue adding to it
		hlsFlags = "delete_segments+append_list"
	}

	playlistPath := b.playlistFilePath()
	b.ffmpegCmd = exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0",
		"-c", "copy",
		"-f", "hls",
		"-hls_time", strconv.Itoa(int(b.segmentDuration.Seconds())),
		// ffmpeg now keeps 65 files on disk
		"-hls_list_size", strconv.Itoa(listSize),
		"-hls_flags", hlsFlags,
		playlistPath,
	)

	var err error
	b.ffmpegStdin, err = b.ffmpegCmd.StdinPipe()
	if err != nil {
		return err
	}
	b.ytCmd.Stdout = b.ffmpegStdin

	if err := b.ffmpegCmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}
	if err := b.ytCmd.Start(); err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}

	b.isRunning = true
	b.stopped = make(chan struct{})
	logging.Info("LiveBuffer: Capture started.")

	go func() {
		defer close(b.stopped)
		ytCmdErr := b.ytCmd.Wait()
		ffmpegStdinErr := b.ffmpegStdin.Close()
		ffmpegCmdErr := b.ffmpegCmd.Wait()
		logging.Info("LiveBuffer: Capture stopped.")
		if ytCmdErr != nil {
			logging.Error("yt-dlp error", "err", ytCmdErr)
		}
		if ffmpegStdinErr != nil {
			logging.Error("ffmpeg stdin close error", "err", ffmpegStdinErr)
		}
		if ffmpegCmdErr != nil {
			logging.Error("ffmpeg error", "err", ffmpegCmdErr)
		}
	}()

	return nil
}

// Stop gracefully terminates the stream capture and blocks until all files are finalized.
func (b *LiveBuffer) Stop() {
	b.mu.Lock()
	if !b.isRunning {
		b.mu.Unlock()
		return
	}
	b.isRunning = false

	logging.Info("LiveBuffer: Initiating graceful shutdown...")

	if b.ytCmd != nil && b.ytCmd.Process != nil {
		b.ytCmd.Process.Signal(os.Interrupt)
	}

	if b.ffmpegStdin != nil {
		b.ffmpegStdin.Close()
	}

	b.mu.Unlock()

	<-b.stopped
}

// ExportClip safely extracts a timeframe and merges it into a valid .mp4 file.
// startAgo and endAgo represent how far back in time to grab (e.g., 30m ago to 10m ago).
func (b *LiveBuffer) ExportClip(startAgo, endAgo time.Duration, outputPath string) error {
	if startAgo > b.bufferDuration {
		return fmt.Errorf("requested timeframe exceeds the allowed logical buffer of %v", b.bufferDuration)
	}
	if startAgo <= endAgo {
		return fmt.Errorf("start time must be older than end time")
	}

	safeSegments, err := b.getSafeHlsSegments()
	if err != nil {
		return fmt.Errorf("failed to get safe segments: %w", err)
	}

	targetSegments, startTime, err := trimSegments(safeSegments, startAgo, endAgo)
	if err != nil {
		return fmt.Errorf("failed to trim segments: %w", err)
	}

	concatFile, cleanup, err := fsutil.TemporaryFile()
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer cleanup()

	err = b.concatInto(targetSegments, concatFile)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}

	trimDuration := startAgo - endAgo

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute output path: %w", err)
	}

	mergeCmd := exec.Command("ffmpeg", "-y",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-ss", fmt.Sprintf("%.3f", startTime.Seconds()),
		"-t", fmt.Sprintf("%.3f", trimDuration.Seconds()),
		"-c", "copy", // No re-encoding, extremely fast
		absOutputPath,
	)

	if err := mergeCmd.Run(); err != nil {
		return fmt.Errorf("failed to merge clip: %w", err)
	}

	return nil
}

func (b *LiveBuffer) getSafeHlsSegments() ([]hlsSegment, error) {
	playlistPath := b.playlistFilePath()

	file, err := os.Open(playlistPath)
	if err != nil {
		return nil, errors.New("buffer not ready or missing")
	}
	defer file.Close()

	var segments []hlsSegment
	var currentDuration float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXTINF:") {
			parsed, err := strconv.ParseFloat(strings.TrimPrefix(strings.TrimSuffix(line, ","), "#EXTINF:"), 64)
			if err != nil {
				logging.Warn("failed to parse segment duration from playlist", "line", line, "err", err)
				continue
			}
			currentDuration = parsed
		} else if strings.HasSuffix(line, ".ts") {
			segments = append(segments, hlsSegment{
				filename: line,
				duration: time.Duration(currentDuration * float64(time.Second)),
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read playlist: %w", err)
	}

	if len(segments) < 2 {
		return nil, fmt.Errorf("not enough segments in buffer to export")
	}

	safeSegments := segments[:len(segments)-1]
	return safeSegments, nil
}

func (b *LiveBuffer) concatInto(segments []hlsSegment, concatFilePath string) error {
	concatFile, err := os.Create(concatFilePath)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer concatFile.Close()
	for _, segment := range segments {
		// Format required by ffmpeg concat demuxer
		relFilePath := filepath.Join(b.outputDir, segment.filename)
		absFilePath, err := filepath.Abs(relFilePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for segment: %w", err)
		}
		escapedFilePath := strings.ReplaceAll(absFilePath, "'", "'\\''")
		_, err = fmt.Fprintf(concatFile, "file '%s'\n", escapedFilePath)
		if err != nil {
			return fmt.Errorf("failed to write to concat file: %w", err)
		}
	}
	return nil
}

func trimSegments(safeSegments []hlsSegment, startAgo time.Duration, endAgo time.Duration) ([]hlsSegment, time.Duration, error) {
	var totalSafeTime time.Duration
	for _, seg := range safeSegments {
		totalSafeTime += seg.duration
	}

	var targetSegments []hlsSegment
	var currentTime time.Duration
	var startTime time.Duration

	targetStart := totalSafeTime - startAgo
	targetEnd := totalSafeTime - endAgo

	for _, seg := range safeSegments {
		segmentStartTime := currentTime
		segmentEndTime := currentTime + seg.duration

		if segmentEndTime > targetStart && segmentStartTime < targetEnd {
			if len(targetSegments) == 0 {
				startTime = max(targetStart-segmentStartTime, 0)
			}
			targetSegments = append(targetSegments, seg)
		}
		currentTime += seg.duration
	}

	if len(targetSegments) == 0 {
		return nil, 0, fmt.Errorf("requested timeframe is outside the current buffer window")
	}

	return targetSegments, startTime, nil
}
