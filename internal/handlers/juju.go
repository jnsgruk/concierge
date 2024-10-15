package handlers

import (
	"fmt"
	"log/slog"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/runner"
	"golang.org/x/sync/errgroup"
)

// NewJujuHandler constructs a new JujuHandler instance.
func NewJujuHandler(config *config.Config, runner runner.CommandRunner, providers []providers.Provider) *JujuHandler {
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
	runner        runner.CommandRunner
}

// Prepare bootstraps Juju on the configured providers.
func (j *JujuHandler) Prepare() error {
	dir := path.Join(".local", "share", "juju")

	err := j.runner.MkHomeSubdirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	err = j.bootstrap()
	if err != nil {
		return fmt.Errorf("failed to bootstrap Juju controller: %w", err)
	}

	return nil
}

// Restore uninstalls Juju from the system.
func (j *JujuHandler) Restore() error {
	err := j.runner.RemoveAllHome(path.Join(".local", "share", "juju"))
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
		slog.Info("Previous Juju controller found", "provider", provider.Name())
		return nil
	}

	bootstrapArgs := []string{
		"bootstrap",
		provider.CloudName(),
		controllerName,
		"--verbose",
	}

	// Get a sorted list of the model-default keys
	keys := make([]string, 0, len(j.modelDefaults))
	for k := range j.modelDefaults {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	// Iterate over the model-defaults and append them to the bootstrapArgs
	for _, k := range keys {
		bootstrapArgs = append(bootstrapArgs, "--model-default", fmt.Sprintf("%s=%s", k, j.modelDefaults[k]))
	}

	user := j.runner.User().Username

	cmd := runner.NewCommandAs(user, provider.GroupName(), "juju", bootstrapArgs)
	_, err = j.runner.RunWithRetries(cmd, (5 * time.Minute))
	if err != nil {
		return err
	}

	cmd = runner.NewCommandAs(user, "", "juju", []string{"add-model", "-c", controllerName, "testing"})
	_, err = j.runner.Run(cmd)
	if err != nil {
		return err
	}

	slog.Info("Bootstrapped Juju", "provider", provider.Name())
	return nil
}

// checkBootstrapped checks whether concierge has already been bootstrapped on a given provider.
func (j *JujuHandler) checkBootstrapped(controllerName string) (bool, error) {
	user := j.runner.User().Username
	cmd := runner.NewCommandAs(user, "", "juju", []string{"show-controller", controllerName})

	result, err := j.runner.Run(cmd)
	if err != nil && strings.Contains(string(result), "not found") {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
