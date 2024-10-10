package config

type ConfigOverrides struct {
	JujuChannel       string
	MicroK8s          string
	LXDChannel        string
	CharmcraftChannel string
	SnapcraftChannel  string
	RockcraftChannel  string

	ExtraSnaps []string
	ExtraDebs  []string
}
