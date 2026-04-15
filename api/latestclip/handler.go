package latestclip

import (
	"net/http"

	"github.com/matthiasharzer/livestream-snapshotting-tool/showmaster"
	"github.com/matthiasharzer/livestream-snapshotting-tool/util/fsutil"
)

func Handler(latestClip *showmaster.Clip) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tempFile, cleanup, err := fsutil.TemporaryFile()
		if err != nil {
			http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
			return
		}
		defer cleanup()

		hasPath, err := latestClip.CopyTo(tempFile)
		if err != nil {
			http.Error(w, "failed to copy clip to temporary file", http.StatusInternalServerError)
			return
		}
		if !hasPath {
			http.Error(w, "no clip available", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, tempFile)
	}
}
