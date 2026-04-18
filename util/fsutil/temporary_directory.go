package fsutil

import (
	"os"
)

func TemporaryDirectory() (string, func(), error) {
	dir, err := os.MkdirTemp("", "livebuffer-*")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup, nil
}
