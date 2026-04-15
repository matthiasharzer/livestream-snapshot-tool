package showmaster

import (
	"io"
	"os"
	"sync"
)

type Clip struct {
	Path string

	mutex sync.RWMutex
}

func (c *Clip) ReplacePath(newPath string) (oldPath string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	oldPath = c.Path
	c.Path = newPath
	return oldPath
}

func (c *Clip) CopyTo(filePath string) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Path == "" {
		return false, nil
	}

	srcFile, err := os.Open(c.Path)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	destFile, err := os.Create(filePath)
	if err != nil {
		return false, err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return false, err
	}

	return true, nil
}
