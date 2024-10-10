package providers

// Provider describes the set of methods expected to be available on a
// provider that concierge can try to bootstrap Juju onto.
type Provider interface {
	// Init is used for installing/configuring the provider.
	Init() error
	// Name reports the name of the provider used internally by concierge.
	Name() string
	// CloudName reports name of the provider as Juju sees it.
	CloudName() string
	// GroupName reports the name of a POSIX user group that can be used
	// to allow non-root users to interact with the provider (where applicable).
	GroupName() string
}
