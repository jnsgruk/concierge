package concierge

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/jnsgruk/concierge/internal/apt"
	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/juju"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/snap"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// Restore reverses the provisioning process, returning the machine to its
func (m *Manager) Restore() error {
	err := m.loadRuntimeConfig()
	if err != nil {
		return fmt.Errorf("failed to load previous runtime configuration: %w", err)
	}

	var eg errgroup.Group

	// Remove snaps in one goroutine
	eg.Go(func() error {
		err := m.removeSnaps()
		if err != nil {
			return fmt.Errorf("failed to remove snap packages: %w", err)
		}
		return nil
	})

	// In another goroutine, start removing apt packages which can be done
	// without conflict with the snap operations above
	eg.Go(func() error {
		err := m.removeAptPackages()
		if err != nil {
			return fmt.Errorf("failed to install apt packages: %w", err)
		}

		return nil
	})

	// Wait for all the other installations to have happened
	if err := eg.Wait(); err != nil {
		return err
	}

	err = m.removeProviders()
	if err != nil {
		return fmt.Errorf("failed to remove providers: %w", err)
	}

	err = m.removeJuju()
	if err != nil {
		return fmt.Errorf("failed to remove juju: %w", err)
	}

	err = m.recordRuntimeConfig()
	if err != nil {
		return fmt.Errorf("failed to remove cached runtime config: %w", err)
	}

	return nil
}

// loadRuntimeConfig loads a previously cached concierge runtime configuration.
func (m *Manager) loadRuntimeConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	recordPath := path.Join(home, ".cache", "concierge", "concierge.yaml")

	if _, err := os.Stat(recordPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no previous runtime configuration found: %w", err)
	}

	contents, err := os.ReadFile(recordPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config config.Config
	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	m.config = &config

	slog.Debug("loaded previous runtime configuration", "path", recordPath)

	return nil
}

// removeSnaps iterates over the list of host snaps that were installed
// and removes them one by one.
func (m *Manager) removeSnaps() error {
	for _, s := range append(m.config.Host.Snaps, m.config.Overrides.ExtraSnaps...) {
		snap := snap.NewSnapFromString(s)
		if err := snap.Remove(true); err != nil {
			return fmt.Errorf("failed to remove snap: %w", err)
		}
	}

	return nil
}

// removeAptPackages removes previously installed packages.
func (m *Manager) removeAptPackages() error {
	packages := append(m.config.Host.Packages, m.config.Overrides.ExtraDebs...)

	for _, p := range packages {
		err := apt.NewAptPackage(p).Remove()
		if err != nil {
			return fmt.Errorf("failed to remove apt package: %w", err)
		}
	}

	return nil
}

// removeProviders iterates over the configured providers and removes them.
func (m *Manager) removeProviders() error {
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
		err := provider.Remove()
		if err != nil {
			return fmt.Errorf("failed to decommission %s: %w", provider.Name(), err)
		}

	}

	return nil
}

// removeJuju removes juju and its config from the host.
func (m *Manager) removeJuju() error {
	return juju.NewJuju(m.config, m.Providers).Remove()
}

// removeRuntimeConfig deletes the configuration record once the machine is restored.
func (m *Manager) removeRuntimeConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	return os.RemoveAll(path.Join(home, ".cache", "concierge"))
}
