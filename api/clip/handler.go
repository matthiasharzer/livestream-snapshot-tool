package clip

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/stream"
	"github.com/matthiasharzer/livebuffer/util/fsutil"
)

func Handler(buffer *stream.LiveBuffer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

		if startStr == "" || endStr == "" {
			http.Error(w, "missing 'start' or 'end' query parameters (e.g., ?start=15m&end=5m)", http.StatusBadRequest)
			return
		}

		startAgo, err := time.ParseDuration(startStr)
		if err != nil {
			http.Error(w, "invalid 'start' format. Use Go duration strings (e.g., 2m, 30s)", http.StatusBadRequest)
			return
		}

		endAgo, err := time.ParseDuration(endStr)
		if err != nil {
			http.Error(w, "invalid 'end' format. Use Go duration strings (e.g., 2m, 30s)", http.StatusBadRequest)
			return
		}

		if startAgo < 0 || endAgo < 0 {
			http.Error(w, "'start' and 'end' must be non-negative durations", http.StatusBadRequest)
			return
		}
		if startAgo > buffer.BufferDuration {
			http.Error(w, fmt.Sprintf("requested timeframe exceeds the allowed logical buffer of %v", buffer.BufferDuration), http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if startAgo <= endAgo {
			http.Error(w, "start time must be older than end time", http.StatusBadRequest)
			return
		}

		tempMp4Path, cleanup, err := fsutil.TemporaryFile(fsutil.TemporaryFileWithEnding(".mp4"))
		if err != nil {
			logging.Error("failed to create temp file", "err", err)
			http.Error(w, "internal Server Error", http.StatusInternalServerError)
			return
		}
		defer cleanup()

		err = buffer.ExportClip(r.Context(), startAgo, endAgo, tempMp4Path)
		if err != nil {
			logging.Error("failed to export clip", "err", err)
			http.Error(w, "failed to generate clip", http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("clip_%s_to_%s.mp4", startAgo.String(), endAgo.String())

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

		http.ServeFile(w, r, tempMp4Path)
	}
}
