package showmaster

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
)

type Clip struct {
	Path string

	mutex sync.RWMutex
}

func (c *Clip) SetPath(path string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Path = path
}

func (c *Clip) Reader() (io.ReadCloser, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.Path == "" {
		return nil, errors.New("no clip path set")
	}

	file, err := os.ReadFile(c.Path)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(file)

	return io.NopCloser(reader), nil
}
