package config

type ConfigOverrides struct {
	K8sChannel        string
	JujuChannel       string
	MicroK8sChannel   string
	LXDChannel        string
	CharmcraftChannel string
	SnapcraftChannel  string
	RockcraftChannel  string

	GoogleCredentialFile string

	ExtraSnaps []string
	ExtraDebs  []string
}
