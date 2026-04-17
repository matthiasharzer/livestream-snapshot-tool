package fsutil

import "os"

type TempFileOptions struct {
	Prefix     string
	FileEnding string
}

func applyTempFileOptions(opts ...TempFileOptions) TempFileOptions {
	var merged TempFileOptions
	for _, opt := range opts {
		if opt.Prefix != "" {
			merged.Prefix = opt.Prefix
		}
		if opt.FileEnding != "" {
			merged.FileEnding = opt.FileEnding
		}
	}
	return merged
}

func TemporaryFileWithPrefix(prefix string) TempFileOptions {
	return TempFileOptions{
		Prefix: prefix,
	}
}

func TemporaryFileWithEnding(fileEnding string) TempFileOptions {
	return TempFileOptions{
		FileEnding: fileEnding,
	}
}

func TemporaryFile(options ...TempFileOptions) (string, func(), error) {
	opts := applyTempFileOptions(options...)

	pattern := opts.Prefix + "*"
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
