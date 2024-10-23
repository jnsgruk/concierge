package config

type ConfigOverrides struct {
	CanonicalK8sChannel string
	JujuChannel         string
	MicroK8sChannel     string
	LXDChannel          string
	CharmcraftChannel   string
	SnapcraftChannel    string
	RockcraftChannel    string

	GoogleCredentialFile string

	ExtraSnaps []string
	ExtraDebs  []string
}
