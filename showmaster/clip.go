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

func (c *Clip) SetPath(path string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Path = path
}

func (c *Clip) GetPath() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Path
}

func (c *Clip) CopyTo(filePath string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	srcFile, err := os.Open(c.Path)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func (c *Clip) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Path == "" {
		return nil
	}

	err := os.Remove(c.Path)
	if err != nil {
		return err
	}

	c.Path = ""
	return nil
}
