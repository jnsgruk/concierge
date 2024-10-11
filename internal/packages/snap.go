package packages

import (
	"fmt"
	"log/slog"
	"strings"

	snapdClient "github.com/snapcore/snapd/client"
)

// NewSnapFromString returns a constructed snap instance, where the snap is
// specified in shorthand form, i.e. `charmcraft/latest/edge`.
func NewSnapFromString(snap string) *Snap {
	before, after, found := strings.Cut(snap, "/")
	if found {
		return NewSnap(before, after)
	} else {
		return NewSnap(before, "")
	}
}

// NewSnap constructs a new Snap instance.
func NewSnap(name string, channel string) *Snap {
	return &Snap{Name: name, Channel: channel}
}

// Snap represents a snap package on the system.
type Snap struct {
	Name    string
	Channel string
}

// Installed is a helper that reports if the snap is currently Installed.
func (s *Snap) Installed() bool {
	slog.Debug("Querying snap install status", "snap", s.Name)

	snap, _, err := snapdClient.New(nil).Snap(s.Name)
	if err != nil {
		return false
	}

	return snap.Status == snapdClient.StatusActive
}

// Classic reports whether or not the snap at the tip of the specified channel uses
// Classic confinement or not.
func (s *Snap) Classic() (bool, error) {
	slog.Debug("Querying snap confinement", "snap", s.Name)

	snap, _, err := snapdClient.New(nil).FindOne(s.Name)
	if err != nil {
		return false, fmt.Errorf("failed to find snap: %w", err)
	}

	return snap.Confinement == "classic", nil
}

// tracking reports which channel an installed snap is tracking.
func (s *Snap) Tracking() (string, error) {
	slog.Debug("Querying snap channel tracking", "snap", s.Name)

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
