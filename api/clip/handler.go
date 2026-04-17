package clip

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matthiasharzer/livestream-snapshotting-tool/logging"
	"github.com/matthiasharzer/livestream-snapshotting-tool/stream"
	"github.com/matthiasharzer/livestream-snapshotting-tool/util/fsutil"
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

		tempMp4Path, cleanup, err := fsutil.TemporaryFile(fsutil.TemporaryFileWithEnding(".mp4"))
		if err != nil {
			logging.Error("failed to create temp file: %v", err)
			http.Error(w, "internal Server Error", http.StatusInternalServerError)
			return
		}
		defer cleanup()

		err = buffer.ExportClip(startAgo, endAgo, tempMp4Path)
		if err != nil {
			logging.Error("failed to export clip: %v", err)
			http.Error(w, fmt.Sprintf("failed to generate clip: %v", err), http.StatusRequestedRangeNotSatisfiable)
			return
		}

		filename := fmt.Sprintf("clip_%s_to_%s.mp4", startStr, endStr)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

		http.ServeFile(w, r, tempMp4Path)
	}
}
