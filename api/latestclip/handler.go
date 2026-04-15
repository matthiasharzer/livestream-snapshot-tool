package latestclip

import (
	"net/http"

	"github.com/matthiasharzer/livestream-snapshotting-tool/showmaster"
	"github.com/matthiasharzer/livestream-snapshotting-tool/util/fsutil"
)

func Handler(latestClip *showmaster.Clip) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clipPath := latestClip.GetPath()
		if clipPath == "" {
			http.Error(w, "no clip available", http.StatusNotFound)
			return
		}

		tempFile, cleanup, err := fsutil.TemporaryFile()
		if err != nil {
			http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
			return
		}
		defer cleanup()

		err = latestClip.CopyTo(tempFile)
		if err != nil {
			http.Error(w, "failed to copy clip to temporary file", http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, tempFile)
	}
}
