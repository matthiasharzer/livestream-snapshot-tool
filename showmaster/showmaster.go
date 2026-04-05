package showmaster

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
	err := s.LatestClip.Clear()
	if err != nil {
		return err
	}
	s.LatestClip.SetPath(path)
	return nil
}
