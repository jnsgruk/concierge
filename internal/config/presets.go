package config

import "fmt"

// Preset returns a configuration preset by name.
func Preset(preset string) (*Config, error) {
	switch preset {
	case "k8s":
		return k8sPreset, nil
	case "machine":
		return machinePreset, nil
	case "dev":
		return devPreset, nil
	default:
		return nil, fmt.Errorf("unknown preset '%s'", preset)
	}
}

// defaultPackages is the default Juju config for all presets.
var defaultJujuConfig jujuConfig = jujuConfig{
	ModelDefaults: map[string]string{
		"test-mode":                 "true",
		"automatically-retry-hooks": "false",
	},
}

// defaultPackages is the set of packages installed for all presets.
var defaultPackages []string = []string{
	"python3-pip",
	"python3-venv",
}

// defaultSnaps is the set of snaps installed for all presets.
var defaultSnaps []string = []string{
	"charmcraft/latest/stable",
	"jq/latest/stable",
	"yq/latest/stable",
}

// defaultLXDConfig is the standard LXD config used throughout presets.
var defaultLXDConfig lxdConfig = lxdConfig{
	Enable: true,
}

// defaultK8sConfig is the standard MicroK8s config used throughout presets.
var defaultK8sConfig microk8sConfig = microk8sConfig{
	Enable: true,
	Addons: []string{
		"hostpath-storage",
		"dns",
		"rbac",
		"metallb:10.64.140.43-10.64.140.49",
	},
}

// machinePreset is a configuration preset designed to be used when testing
// machine charms.
var machinePreset *Config = &Config{
	Juju: defaultJujuConfig,
	Providers: providerConfig{
		LXD: defaultLXDConfig,
	},
	Host: hostConfig{
		Packages: defaultPackages,
		Snaps:    append(defaultSnaps, "snapcraft/latest/stable"),
	},
}

// k8sPreset is a configuration preset designed to be used when testing
// k8s charms.
var k8sPreset *Config = &Config{
	Juju: defaultJujuConfig,
	Providers: providerConfig{
		MicroK8s: defaultK8sConfig,
	},
	Host: hostConfig{
		Packages: defaultPackages,
		Snaps:    append(defaultSnaps, "rockcraft/latest/stable"),
	},
}

// devPreset combines both the LXD and K8s presets, designed to be used by
// developers when iterating on charms.
var devPreset *Config = &Config{
	Juju: defaultJujuConfig,
	Providers: providerConfig{
		LXD:      defaultLXDConfig,
		MicroK8s: defaultK8sConfig,
	},
	Host: hostConfig{
		Packages: defaultPackages,
		Snaps:    append(defaultSnaps, "rockcraft/latest/stable", "snapcraft/latest/stable"),
	},
}
