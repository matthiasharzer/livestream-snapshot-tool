package showmaster

import (
	"fmt"
	"os"
)

type ShowMaster struct {
	LatestClip *Clip
}

func New() *ShowMaster {
	return &ShowMaster{
		LatestClip: &Clip{},
	}
}

func (s *ShowMaster) AddClip(path string) error {
	// TODO: introduce a different delete strategy once more than one clip is supported
	oldPath := s.LatestClip.ReplacePath(path)

	hasOldClip := oldPath != ""
	didPathChange := oldPath != path

	if !hasOldClip || !didPathChange {
		// No need to delete the old clip if the path didn't change
		return nil
	}

	err := os.Remove(oldPath)
	if err != nil {
		return fmt.Errorf("failed to remove old clip: %w", err)
	}
	return nil
}
