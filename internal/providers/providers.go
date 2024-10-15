package providers

import (
	"github.com/jnsgruk/concierge/internal/packages"
)

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
	// Snaps reports the list of snaps required by the provider.
	Snaps() []packages.SnapPackage
}
