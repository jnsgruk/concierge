package handlers

import (
	"fmt"
	"log/slog"

	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/runner"
)

// NewDebHandler constructs a new instance of a DebHandler.
func NewDebHandler(runner runner.CommandRunner, debs []*packages.Deb) *DebHandler {
	return &DebHandler{
		Debs:   debs,
		runner: runner,
	}
}

// DebHandler can install or remove a set of debs.
type DebHandler struct {
	Debs   []*packages.Deb
	runner runner.CommandRunner
}

// Prepare updates the apt cache and installs a set of debs from the archive.
func (h *DebHandler) Prepare() error {
	if len(h.Debs) == 0 {
		return nil
	}

	err := h.updateAptCache()
	if err != nil {
		return fmt.Errorf("failed to update apt cache: %w", err)
	}

	for _, deb := range h.Debs {
		err := h.installDeb(deb)
		if err != nil {
			return fmt.Errorf("failed to install deb: %w", err)
		}
	}
	return nil
}

// Restore removes a set of debs from the machine.
func (h *DebHandler) Restore() error {
	for _, deb := range h.Debs {
		err := h.removeDeb(deb)
		if err != nil {
			return fmt.Errorf("failed to remove deb: %w", err)
		}
	}

	cmd := runner.NewCommand("apt-get", []string{"autoremove", "-y"})

	_, err := h.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to install apt package: %w", err)
	}

	return nil
}

// installDeb uses `apt` to install the package on the system from the archives.
func (h *DebHandler) installDeb(d *packages.Deb) error {
	cmd := runner.NewCommand("apt-get", []string{"install", "-y", d.Name})

	_, err := h.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to install apt package '%s': %w", d.Name, err)
	}

	slog.Info("Installed apt package", "package", d.Name)
	return nil
}

// Remove uninstalls the deb from the system with `apt`.
func (h *DebHandler) removeDeb(d *packages.Deb) error {
	cmd := runner.NewCommand("apt-get", []string{"remove", "-y", d.Name})

	_, err := h.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove apt package '%s': %w", d.Name, err)
	}

	slog.Info("Removed apt package", "package", d.Name)
	return nil
}

// updateAptCache is a helper method to update the host's package cache.
func (h *DebHandler) updateAptCache() error {
	cmd := runner.NewCommand("apt-get", []string{"update"})

	_, err := h.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to update apt package lists: %w", err)
	}

	return nil
}
