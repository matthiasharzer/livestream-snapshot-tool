package showmaster

type ShowMaster struct {
	LatestClip *Clip
}

func New() *ShowMaster {
	return &ShowMaster{
		LatestClip: &Clip{},
	}
}
