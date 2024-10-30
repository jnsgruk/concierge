package providers

import (
	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
)

// SupportedProviders is a list of stringified names of supported providers.
var SupportedProviders []string = []string{
	"k8s",
	"google",
	"lxd",
	"microk8s",
}

// Provider describes the set of methods expected to be available on a
// provider that concierge can try to bootstrap Juju onto.
type Provider interface {
	// Prepare is used for installing/configuring the provider.
	Prepare() error
	// Restore is used for uninstalling the provider.
	Restore() error
	// Name reports the name of the provider used internally by concierge.
	Name() string
	// Bootstrap reports whether or not a Juju controller should be bootstrapped on the provider.
	Bootstrap() bool
	// CloudName reports name of the provider as Juju sees it.
	CloudName() string
	// GroupName reports the name of a POSIX user group that can be used
	// to allow non-root users to interact with the provider (where applicable).
	GroupName() string
	// Credentials reports the section of Juju's credentials.yaml for the provider
	Credentials() map[string]interface{}
}

// NewProvider returns a newly constructed provider based on a stringified name of the provider.
func NewProvider(providerName string, runner runner.CommandRunner, config *config.Config) Provider {
	if providerName == "lxd" && config.Providers.LXD.Enable {
		return NewLXD(runner, config)
	} else if providerName == "microk8s" && config.Providers.MicroK8s.Enable {
		return NewMicroK8s(runner, config)
	} else if providerName == "google" && config.Providers.Google.Enable {
		return NewGoogle(runner, config)
	} else if providerName == "k8s" && config.Providers.K8s.Enable {
		return NewK8s(runner, config)
	} else {
		return nil
	}
}
