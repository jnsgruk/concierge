package config

// Config represents concierge's configuration format.
type Config struct {
	Juju      jujuConfig     `mapstructure:"juju"`
	Providers providerConfig `mapstructure:"providers"`
	Host      hostConfig     `mapstructure:"host"`

	// The following are added at runtime according to CLI flags
	Overrides ConfigOverrides `mapstructure:"overrides"`
	Verbose   bool            `mapstructure:"verbose"`
	Trace     bool            `mapstructure:"trace"`
}

// jujuConfig represents the configuration for juju, including the desired version,
// and defaults/constraints for the bootstrap process.
type jujuConfig struct {
	Channel string `mapstructure:"channel"`
	// The set of model-defaults to be passed to Juju during bootstrap
	ModelDefaults map[string]string `mapstructure:"model-defaults"`
	// The set of bootstrap constraints to be passed to Juju
	BootstrapConstraints map[string]string `mapstructure:"bootstrap-constraints"`
}

// providerConfig represents the set of providers to be configured and bootstrapped.
type providerConfig struct {
	K8s      k8sConfig      `mapstructure:"k8s"`
	LXD      lxdConfig      `mapstructure:"lxd"`
	Google   googleConfig   `mapstructure:"google"`
	MicroK8s microk8sConfig `mapstructure:"microk8s"`
}

// lxdConfig represents how LXD should be configured on the host.
type lxdConfig struct {
	Enable               bool              `mapstructure:"enable"`
	Bootstrap            bool              `mapstructure:"bootstrap"`
	Channel              string            `mapstructure:"channel"`
	ModelDefaults        map[string]string `mapstructure:"model-defaults"`
	BootstrapConstraints map[string]string `mapstructure:"bootstrap-constraints"`
}

// googleConfig represents how Juju should be configured for Google Cloud use.
type googleConfig struct {
	Enable               bool              `mapstructure:"enable"`
	Bootstrap            bool              `mapstructure:"bootstrap"`
	CredentialsFile      string            `mapstructure:"credentials-file"`
	ModelDefaults        map[string]string `mapstructure:"model-defaults"`
	BootstrapConstraints map[string]string `mapstructure:"bootstrap-constraints"`
}

// microk8sConfig represents how MicroK8s should be configured on the host.
type microk8sConfig struct {
	Enable               bool              `mapstructure:"enable"`
	Bootstrap            bool              `mapstructure:"bootstrap"`
	Channel              string            `mapstructure:"channel"`
	Addons               []string          `mapstructure:"addons"`
	ModelDefaults        map[string]string `mapstructure:"model-defaults"`
	BootstrapConstraints map[string]string `mapstructure:"bootstrap-constraints"`
}

// k8sConfig represents how MicroK8s should be configured on the host.
type k8sConfig struct {
	Enable               bool                         `mapstructure:"enable"`
	Bootstrap            bool                         `mapstructure:"bootstrap"`
	Channel              string                       `mapstructure:"channel"`
	Features             map[string]map[string]string `mapstructure:"features"`
	ModelDefaults        map[string]string            `mapstructure:"model-defaults"`
	BootstrapConstraints map[string]string            `mapstructure:"bootstrap-constraints"`
}

// SnapConfig represents the configuration for a specific snap to be installed.
type SnapConfig struct {
	// Channel is the channel from which to install the snap.
	Channel string `mapstructure:"channel"`
	// Connections is a list of snap connections to form.
	Connections []string `mapstructure:"connections"`
}

// hostConfig is a top-level field containing addition configuration for the host being
// configured.
type hostConfig struct {
	// Packages is a of apt packages to be installed from the archive
	Packages []string `mapstructure:"packages"`
	// Snaps is a map of snaps to be installed.
	Snaps map[string]SnapConfig `mapstructure:"snaps"`
}
