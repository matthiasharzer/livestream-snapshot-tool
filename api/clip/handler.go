package clip

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/matthiasharzer/livebuffer/showmaster"
	"github.com/matthiasharzer/livebuffer/util/fsutil"
)

func resolveClipNumber(clipText string) (int, error) {
	if clipText == "latest" {
		return 0, nil
	}

	clipNr, err := strconv.Atoi(clipText)
	if err != nil {
		return 0, fmt.Errorf("failed to convert %s to integer: %v", clipText, err)
	}

	return clipNr, nil
}

func Handler(master *showmaster.ShowMaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clip := r.PathValue("clip")

		clipNr, err := resolveClipNumber(clip)
		if err != nil {
			http.Error(w, "invalid clip number", http.StatusBadRequest)
			return
		}

		if clipNr < 0 || clipNr >= master.HistorySize() {
			http.Error(w, "clip number out of bounds", http.StatusBadRequest)
			return
		}

		requestedClip := master.NthClip(clipNr)
		if requestedClip == nil {
			http.Error(w, "clip unavailable", http.StatusNotFound)
			return
		}

		tempFile, cleanup, err := fsutil.TemporaryFile()
		if err != nil {
			http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
			return
		}
		defer cleanup()

		hasPath, err := requestedClip.CopyTo(tempFile)
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
