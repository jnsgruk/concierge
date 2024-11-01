package concierge

import (
	"fmt"
	"log/slog"
	"path"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/system"
	"gopkg.in/yaml.v3"
)

// NewManager constructs a new instance of the concierge manager.
func NewManager(config *config.Config) (*Manager, error) {
	system, err := system.NewSystem(config.Trace)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise system: %w", err)
	}

	return &Manager{
		config: config,
		system: system,
	}, nil
}

// Manager is a construct for controlling the main execution of concierge.
type Manager struct {
	Plan   *Plan
	system system.Worker
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
	m.Plan = NewPlan(m.config, m.system)
	return m.Plan.Execute(action)
}

// recordRuntimeConfig dumps the current manager config into a file in the user's home
// directory, such that it can be read later and used to restore the machine.
func (m *Manager) recordRuntimeConfig() error {
	configYaml, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config file as yaml: %w", err)
	}

	filepath := path.Join(".cache", "concierge", "concierge.yaml")
	err = m.system.WriteHomeDirFile(filepath, configYaml)
	if err != nil {
		return fmt.Errorf("failed to write runtime config file: %w", err)
	}

	slog.Debug("Merged runtime configuration saved", "path", filepath)

	return nil
}

// loadRuntimeConfig loads a previously cached concierge runtime configuration.
func (m *Manager) loadRuntimeConfig() error {
	recordPath := path.Join(".cache", "concierge", "concierge.yaml")

	contents, err := m.system.ReadHomeDirFile(recordPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config config.Config
	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	m.config = &config

	slog.Debug("Loaded previous runtime configuration", "path", recordPath)

	return nil
}
