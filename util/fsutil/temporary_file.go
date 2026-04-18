package fsutil

import "os"

func TemporaryFile() (string, func(), error) {
	file, err := os.CreateTemp("", "livebuffer-")
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	cleanup := func() {
		_ = os.Remove(file.Name())
	}

	return file.Name(), cleanup, nil
}
