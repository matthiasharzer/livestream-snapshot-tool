package latestclip

import (
	"io"
	"net/http"

	"github.com/matthiasharzer/livestream-snapshot-tool/showmaster"
)

func Handler(master *showmaster.ShowMaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clipReader, err := master.LatestClip.Reader()
		if err != nil {
			http.Error(w, "Failed to get latest clip", http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = clipReader.Close()
		}()

		w.Header().Set("Content-Type", "video/mp4")

		_, err = io.Copy(w, clipReader)
		if err != nil {
			http.Error(w, "Failed to stream latest clip", http.StatusInternalServerError)
			return
		}
	}
}
