package showmaster

import (
	"fmt"
	"os"
	"sync"
)

type ShowMaster struct {
	historicalClips []*Clip // 0 is the oldest clip, len-1 is the newest clip
	historySize     int

	mutex sync.RWMutex
}

func New(historySize int) (*ShowMaster, error) {
	if historySize < 1 {
		return nil, fmt.Errorf("history size must be at least 1")
	}
	return &ShowMaster{
		historicalClips: make([]*Clip, 0, historySize),
		historySize:     historySize,
		mutex:           sync.RWMutex{},
	}, nil
}

func (s *ShowMaster) AddClip(path string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.historicalClips) == s.historySize {
		oldestClip := s.historicalClips[0]
		if err := os.Remove(oldestClip.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove oldest clip: %w", err)
		}
		s.historicalClips = s.historicalClips[1:]
	}
	clip, err := NewClip(path)
	if err != nil {
		return fmt.Errorf("failed to create clip: %w", err)
	}
	s.historicalClips = append(s.historicalClips, clip)
	return nil
}

func (s *ShowMaster) HistorySize() int {
	return s.historySize
}

// NthClip returns the nth most recent clip, where n=0 is the most recent clip, n=1 is the second most recent clip, and so on. If n is out of bounds, it returns nil.
func (s *ShowMaster) NthClip(n int) *Clip {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if n < 0 || n >= len(s.historicalClips) {
		return nil
	}
	return s.historicalClips[len(s.historicalClips)-1-n]
}
