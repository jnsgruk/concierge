package handlers

import (
	"fmt"
	"log/slog"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/runner"
)

// NewSnapHandler constructs a new instance of a SnapHandler.
func NewSnapHandler(config *config.Config, runner *runner.Runner, snaps []*packages.Snap) *SnapHandler {
	return &SnapHandler{
		Snaps:  snaps,
		config: config,
		runner: runner,
	}
}

// SnapHandler can install or remove a set of snaps.
type SnapHandler struct {
	Snaps  []*packages.Snap
	config *config.Config
	runner *runner.Runner
}

// Prepare installs a set of snaps on the machine.
func (h *SnapHandler) Prepare() error {
	for _, snap := range h.Snaps {
		err := h.installSnap(snap)
		if err != nil {
			return fmt.Errorf("failed to install snap: %w", err)
		}
	}
	return nil
}

// Restore removes a set of snaps from the machine.
func (h *SnapHandler) Restore() error {
	for _, snap := range h.Snaps {
		err := h.removeSnap(snap)
		if err != nil {
			return fmt.Errorf("failed to remove snap: %w", err)
		}
	}
	return nil
}

// installSnap ensures that the specified snap is installed at the specified channel.
// If already installed, but on the wrong channel, the snap is refreshed.
func (h *SnapHandler) installSnap(s *packages.Snap) error {
	var action string
	if s.Installed() {
		action = "refresh"
	} else {
		action = "install"
	}

	args := []string{action, s.Name}

	if s.Channel != "" {
		args = append(args, "--channel", s.Channel)
	}

	classic, err := s.Classic()
	if err != nil {
		return fmt.Errorf("failed to determine if snap '%s' is classic: %w", s.Name, err)
	}

	if classic {
		args = append(args, "--classic")
	}

	cmd := runner.NewCommand("snap", args)
	_, err = h.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	trackingChannel, err := s.Tracking()
	if err != nil {
		return fmt.Errorf("failed to resolve which channel the '%s' snap is tracking: %w", s.Name, err)
	}

	slog.Info("Installed snap", "snap", s.Name, "channel", trackingChannel)
	return nil
}

// Remove uninstalls the specified snap from the system, optionally purging its data.
func (h *SnapHandler) removeSnap(s *packages.Snap) error {
	args := []string{"remove", s.Name, "--purge"}

	cmd := runner.NewCommand("snap", args)
	_, err := h.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove snap '%s': %w", s.Name, err)
	}

	return nil
}
