package concierge

import (
	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/providers"
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
