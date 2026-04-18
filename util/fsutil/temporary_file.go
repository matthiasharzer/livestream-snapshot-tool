package fsutil

import "os"

type TempFileOptions struct {
	FileEnding string
}

func applyTempFileOptions(opts ...TempFileOptions) TempFileOptions {
	var merged TempFileOptions
	for _, opt := range opts {
		if opt.FileEnding != "" {
			merged.FileEnding = opt.FileEnding
		}
	}
	return merged
}

func TemporaryFileWithEnding(fileEnding string) TempFileOptions {
	return TempFileOptions{
		FileEnding: fileEnding,
	}
}

func TemporaryFile(options ...TempFileOptions) (string, func(), error) {
	opts := applyTempFileOptions(options...)

	pattern := "livebuffer-*"
	if opts.FileEnding != "" {
		pattern += opts.FileEnding
	}

	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	cleanup := func() {
		_ = os.Remove(file.Name())
	}

	return file.Name(), cleanup, nil
}
