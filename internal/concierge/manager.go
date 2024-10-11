package concierge

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
	"gopkg.in/yaml.v3"
)

// NewManager constructs a new instance of the concierge manager.
func NewManager(config *config.Config) *Manager {
	return &Manager{
		config: config,
		runner: runner.NewRunner(config.Trace),
	}
}

// Manager is a construct for controlling the main execution of concierge.
type Manager struct {
	Plan   *Plan
	runner *runner.Runner
	config *config.Config
}

// Prepare runs the steps required for provisioning the machine according to
// the config.
func (m *Manager) Prepare() error {
	return m.execute(PrepareAction)
}

// Restore reverses the provisioning process, returning the machine to its.
func (m *Manager) Restore() error {
	return m.execute(RestoreAction)
}

// execute runs the overlord with a specified action.
func (m *Manager) execute(action string) error {
	switch action {
	case PrepareAction:
		err := m.recordRuntimeConfig()
		if err != nil {
			return fmt.Errorf("failed to record config file: %w", err)
		}
	case RestoreAction:
		err := m.loadRuntimeConfig()
		if err != nil {
			return fmt.Errorf("failed to load previous runtime configuration: %w", err)
		}
	default:
		return fmt.Errorf("unknown handler action: %s", action)
	}

	// Create the installation/preparation plan
	m.Plan = NewPlan(m.config, m.runner)
	return m.Plan.Execute(action)
}

// recordRuntimeConfig dumps the current manager config into a file in the user's home
// directory, such that it can be read later and used to restore the machine.
func (m *Manager) recordRuntimeConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	err = os.MkdirAll(path.Join(home, ".cache", "concierge"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create '.cache/concierge' subdirectory in user's home directory: %w", err)
	}

	configYaml, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config file as yaml: %w", err)
	}

	recordPath := path.Join(home, ".cache", "concierge", "concierge.yaml")

	if err := os.WriteFile(recordPath, configYaml, 0644); err != nil {
		return fmt.Errorf("failed to write config record file: %w", err)
	}

	slog.Debug("merged runtime configuration saved", "path", recordPath)

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
