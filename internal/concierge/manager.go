package concierge

import (
	"fmt"

	"github.com/jnsgruk/concierge/internal/apt"
	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/juju"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/snap"
	"golang.org/x/sync/errgroup"
)

// NewManager constructs a new instance of the concierge manager.
func NewManager(config *config.Config) *Manager {
	return &Manager{config: config}
}

// Manager is a construct for controlling the main execution of concierge.
type Manager struct {
	Providers []providers.Provider
	config    *config.Config
}

// Execute runs the steps required for provisioning the machine according to
// the config.
func (m *Manager) Execute() error {
	var eg errgroup.Group

	// Install snaps in one goroutine
	eg.Go(func() error {
		err := m.installSnaps()
		if err != nil {
			return fmt.Errorf("failed to install snap packages: %w", err)
		}
		return nil
	})

	// In another goroutine, start install apt packages which can be done
	// without conflict with the snap operations above
	eg.Go(func() error {
		err := m.installAptPackages()
		if err != nil {
			return fmt.Errorf("failed to install apt packages: %w", err)
		}

		return nil
	})

	// Wait for all the other installations to have happened
	if err := eg.Wait(); err != nil {
		return err
	}

	err := m.setupProviders()
	if err != nil {
		return fmt.Errorf("failed to setup providers: %w", err)
	}

	err = m.setupJuju()
	if err != nil {
		return fmt.Errorf("failed to setup juju: %w", err)
	}

	return nil
}

// setupProviders iterates over the providers specified in the configuration,
// installing and performing minimal setup for each to ensure they're functional
// to the extent that a Juju controller can be bootstrapped on each provider.
func (m *Manager) setupProviders() error {
	providerConfig := m.config.Providers

	if providerConfig.MicroK8s.Enable {
		p := providers.NewMicroK8s(m.config)
		m.Providers = append(m.Providers, p)
	}

	if providerConfig.LXD.Enable {
		p := providers.NewLXD(m.config)
		m.Providers = append(m.Providers, p)
	}

	// Range over the providers and initialise them
	for _, provider := range m.Providers {
		err := provider.Init()
		if err != nil {
			return fmt.Errorf("failed to set up %s: %w", provider.Name(), err)
		}

	}

	return nil
}

// installSnaps iterates over the list of host snaps to be installed
// and installs them one by one.
func (m *Manager) installSnaps() error {
	for _, s := range append(m.config.Host.Snaps, m.config.Overrides.ExtraSnaps...) {
		snap := snap.NewSnapFromString(s)

		// Check if the channel has been overridden by a CLI argument/env var
		channelOverride := m.getSnapChannelOverride(snap.Name)
		if channelOverride != "" {
			snap.Channel = channelOverride
		}

		err := snap.Install()
		if err != nil {
			return fmt.Errorf("failed to install snap: %w", err)
		}
	}

	return nil
}

// installAptPackages both updates the apt cache for the system, and installs
// the specified packages.
func (m *Manager) installAptPackages() error {
	packages := append(m.config.Host.Packages, m.config.Overrides.ExtraDebs...)
	if len(packages) == 0 {
		return nil
	}

	err := apt.Update()
	if err != nil {
		return fmt.Errorf("failed to update apt cache: %w", err)
	}

	for _, p := range packages {
		err := apt.NewAptPackage(p).Install()
		if err != nil {
			return fmt.Errorf("failed to install apt package: %w", err)
		}
	}

	return nil
}

// setupJuju installs, configures and bootstraps juju on the specified providers.
func (m *Manager) setupJuju() error {
	err := juju.NewJuju(m.config, m.Providers).Init()
	if err != nil {
		return err
	}

	return nil
}

// getSnapChannelOverride takes the name of a snap. If the snap's version
// is overridden, the overridden channel is returned.
func (m *Manager) getSnapChannelOverride(snap string) string {
	switch snap {
	case "charmcraft":
		return m.config.Overrides.CharmcraftChannel
	case "snapcraft":
		return m.config.Overrides.SnapcraftChannel
	case "rockcraft":
		return m.config.Overrides.RockcraftChannel
	default:
		return ""
	}
}
