package juju

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/runner"
	"github.com/jnsgruk/concierge/internal/snap"
	"golang.org/x/sync/errgroup"
)

// NewJuju constructs a new Juju instance.
func NewJuju(config *config.Config, providers []providers.Provider) *Juju {
	var channel string
	if config.Overrides.JujuChannel != "" {
		channel = config.Overrides.JujuChannel
	} else {
		channel = config.Juju.Channel
	}

	return &Juju{
		Channel:       channel,
		ModelDefaults: config.Juju.ModelDefaults,
		providers:     providers,
	}
}

// Juju represents a Juju installation on the system.
type Juju struct {
	Channel       string
	ModelDefaults map[string]string
	providers     []providers.Provider
}

// Init installs juju and bootstraps it on the configured providers.
func (j Juju) Init() error {
	err := j.install()
	if err != nil {
		return fmt.Errorf("failed to install Juju: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	err = os.MkdirAll(path.Join(home, ".local", "share", "juju"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create '.local/share/juju' subdirectory in user's home directory: %w", err)
	}

	err = j.bootstrap()
	if err != nil {
		return fmt.Errorf("failed to bootstrap Juju controller: %w", err)
	}

	return nil
}

// Remove uninstalls Juju from the system.
func (j Juju) Remove() error {
	err := snap.NewSnap("juju", j.Channel).Remove(true)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	err = os.RemoveAll(path.Join(home, ".local", "share", "juju"))
	if err != nil {
		return fmt.Errorf("failed to remove '.local/share/juju' subdirectory from user's home directory: %w", err)
	}

	slog.Info("Removed Juju")

	return nil
}

// install ensures the Juju snap is installed and tracking the specified channel.
func (j Juju) install() error {
	err := snap.NewSnap("juju", j.Channel).Install()
	if err != nil {
		return err
	}

	return nil
}

// bootstrap iterates over the set of configured providers, and bootstraps each of
// them in parallel with a unique controller name.
func (j Juju) bootstrap() error {
	var eg errgroup.Group

	for _, provider := range j.providers {
		eg.Go(func() error {
			err := j.bootstrapProvider(provider)
			if err != nil {
				return err
			}

			slog.Info("Bootstrapped Juju", "provider", provider.Name())
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// bootstrapProvider bootstraps one specific provider.
func (j Juju) bootstrapProvider(provider providers.Provider) error {
	controllerName := fmt.Sprintf("concierge-%s", provider.Name())

	bootstrapped, err := j.checkBootstrapped(controllerName)
	if err != nil {
		return fmt.Errorf("error checking bootstrap status for provider '%s'", provider.Name())
	}

	if bootstrapped {
		return nil
	}

	bootstrapArgs := []string{
		"bootstrap",
		provider.CloudName(),
		controllerName,
		"--verbose",
	}

	for k, v := range j.ModelDefaults {
		bootstrapArgs = append(bootstrapArgs, "--model-default", fmt.Sprintf("%s=%s", k, v))
	}

	if err := runner.RunCommands(
		runner.NewCommandWithGroup("juju", bootstrapArgs, provider.GroupName()),
		runner.NewCommand("juju", []string{"add-model", "-c", controllerName, "testing"}),
	); err != nil {
		return err
	}

	return nil
}

// checkBootstrapped checks whether concierge has already been bootstrapped on a given provider.
func (j Juju) checkBootstrapped(controllerName string) (bool, error) {
	result, err := runner.NewCommand("juju", []string{"show-controller", controllerName}).Run()
	if err != nil && strings.Contains(result.Stderr.String(), "not found") {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
