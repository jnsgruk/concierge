package snap

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jnsgruk/concierge/internal/runner"
	snapdClient "github.com/snapcore/snapd/client"
)

// NewSnapFromString returns a constructed snap instance, where the snap is
// specified in shorthand form, i.e. `charmcraft/latest/edge`.
func NewSnapFromString(snap string) Snap {
	before, after, found := strings.Cut(snap, "/")
	if found {
		return Snap{Name: before, Channel: after}
	} else {
		return Snap{Name: before}
	}
}

// NewSnap constructs a new Snap instance.
func NewSnap(name string, channel string) Snap {
	return Snap{Name: name, Channel: channel}
}

// Snap represents a snap package on the system.
type Snap struct {
	Name    string
	Channel string
}

// Install ensures that the given snap is installed at the specified channel.
// If already installed, but on the wrong channel, the snap is refreshed.
func (s Snap) Install() error {
	var action string
	if s.installed() {
		action = "refresh"
	} else {
		action = "install"
	}

	args := []string{action, s.Name}

	if s.Channel != "" {
		args = append(args, "--channel", s.Channel)
	}

	classic, err := s.classic()
	if err != nil {
		return fmt.Errorf("failed to determine if snap '%s' is classic: %w", s.Name, err)
	}

	if classic {
		args = append(args, "--classic")
	}

	_, err = runner.NewCommandSudo("snap", args).Run()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	trackingChannel, err := s.trackingChannel()
	if err != nil {
		return fmt.Errorf("failed to resolve which channel the '%s' snap is tracking: %w", s.Name, err)
	}

	slog.Info("Installed snap", "snap", s.Name, "channel", trackingChannel)
	return nil
}

// Remove uninstalls the snap from the system, optionally purging its data.
func (s Snap) Remove(purge bool) error {
	args := []string{"remove", s.Name}

	if purge {
		args = append(args, "--purge")
	}

	_, err := runner.NewCommandSudo("snap", args).Run()
	if err != nil {
		return fmt.Errorf("failed to remove snap '%s': %w", s.Name, err)
	}

	return nil
}

// installed is a helper that reports if the snap is currently installed.
func (s Snap) installed() bool {
	slog.Debug("querying snap install status", "snap", s.Name)

	snap, _, err := snapdClient.New(nil).Snap(s.Name)
	if err != nil {
		return false
	}

	return snap.Status == snapdClient.StatusActive
}

// classic reports whether or not the snap at the tip of the specified channel uses
// classic confinement or not.
func (s Snap) classic() (bool, error) {
	slog.Debug("querying snap confinement", "snap", s.Name)

	snap, _, err := snapdClient.New(nil).FindOne(s.Name)
	if err != nil {
		return false, fmt.Errorf("failed to find snap: %w", err)
	}

	return snap.Confinement == "classic", nil
}

// tracking reports which channel an installed snap is tracking.
func (s Snap) trackingChannel() (string, error) {
	slog.Debug("querying snap channel tracking", "snap", s.Name)

	snap, _, err := snapdClient.New(nil).Snap(s.Name)
	if err != nil {
		return "", fmt.Errorf("failed to find snap: %w", err)
	}

	if snap.Status == snapdClient.StatusActive {
		return snap.TrackingChannel, nil
	} else {
		return "", fmt.Errorf("snap '%s' is not installed", s.Name)
	}
}
