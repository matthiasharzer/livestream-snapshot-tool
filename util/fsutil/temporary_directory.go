package fsutil

import (
	"os"
)

func TemporaryDirectory() (string, func(), error) {
	dir, err := os.MkdirTemp("", "livestream-snapshot-tool-*")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup, nil
}
