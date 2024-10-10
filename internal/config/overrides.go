package config

type ConfigOverrides struct {
	JujuChannel       string
	MicroK8sChannel   string
	LXDChannel        string
	CharmcraftChannel string
	SnapcraftChannel  string
	RockcraftChannel  string

	ExtraSnaps []string
	ExtraDebs  []string
}
