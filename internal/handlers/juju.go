package handlers

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/runner"
	"golang.org/x/sync/errgroup"
)

// NewJujuHandler constructs a new JujuHandler instance.
func NewJujuHandler(config *config.Config, runner *runner.Runner, providers []providers.Provider) *JujuHandler {
	return &JujuHandler{
		modelDefaults: config.Juju.ModelDefaults,
		providers:     providers,
		runner:        runner,
	}
}

// JujuHandler represents a Juju installation on the system.
type JujuHandler struct {
	modelDefaults map[string]string
	providers     []providers.Provider
	runner        *runner.Runner
}

// Prepare bootstraps Juju on the configured providers.
func (j *JujuHandler) Prepare() error {
	snap := packages.NewSnapFromString("juju")
	if !snap.Installed() {
		return fmt.Errorf("juju snap not installed")
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

// Restore uninstalls Juju from the system.
func (j *JujuHandler) Restore() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	err = os.RemoveAll(path.Join(home, ".local", "share", "juju"))
	if err != nil {
		return fmt.Errorf("failed to remove '.local/share/juju' subdirectory from user's home directory: %w", err)
	}

	slog.Info("Restored Juju")

	return nil
}

// bootstrap iterates over the set of configured providers, and bootstraps each of
// them in parallel with a unique controller name.
func (j *JujuHandler) bootstrap() error {
	var eg errgroup.Group

	for _, provider := range j.providers {
		eg.Go(func() error { return j.bootstrapProvider(provider) })
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// bootstrapProvider bootstraps one specific provider.
func (j *JujuHandler) bootstrapProvider(provider providers.Provider) error {
	if !provider.Bootstrap() {
		return nil
	}

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

	for k, v := range j.modelDefaults {
		bootstrapArgs = append(bootstrapArgs, "--model-default", fmt.Sprintf("%s=%s", k, v))
	}

	if err := j.runner.RunCommands(
		runner.NewCommandWithGroup("juju", bootstrapArgs, provider.GroupName()),
		runner.NewCommand("juju", []string{"add-model", "-c", controllerName, "testing"}),
	); err != nil {
		return err
	}

	slog.Info("Bootstrapped Juju", "provider", provider.Name())
	return nil
}

// checkBootstrapped checks whether concierge has already been bootstrapped on a given provider.
func (j *JujuHandler) checkBootstrapped(controllerName string) (bool, error) {
	cmd := runner.NewCommand("juju", []string{"show-controller", controllerName})

	result, err := j.runner.Run(cmd)
	if err != nil && strings.Contains(string(result), "not found") {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
