package packages

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/snapcore/snapd/client"
)

// SnapPackage is an interface with methods that can describe a snap package
type SnapPackage interface {
	// Name reports the name of the snap.
	Name() string
	// Installed reports if the snap is currently Installed.
	Installed() bool
	// Classic reports whether or not the snap at the tip of the specified channel uses
	// Classic confinement or not.
	Classic() (bool, error)
	// tracking reports which channel an installed snap is tracking.
	Tracking() (string, error)
	// Channel reports the snap channel.
	Channel() string
	// SetChannel sets the snap channel.
	SetChannel(c string)
}

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
	return &Snap{name: name, channel: channel, client: client.New(nil)}
}

// Snap represents a snap package on the system.
type Snap struct {
	name    string
	channel string

	client *client.Client
}

// Name reports the name of the snap.
func (s *Snap) Name() string { return s.name }

// Channel reports the snap channel.
func (s *Snap) Channel() string { return s.channel }

// SetChannel sets the snap channel.
func (s *Snap) SetChannel(c string) { s.channel = c }

// Installed is a helper that reports if the snap is currently Installed.
func (s *Snap) Installed() bool {
	slog.Debug("Querying snap install status", "snap", s.name)

	snap, _, err := s.client.Snap(s.name)
	if err != nil {
		return false
	}

	return snap.Status == client.StatusActive
}

// Classic reports whether or not the snap at the tip of the specified channel uses
// Classic confinement or not.
func (s *Snap) Classic() (bool, error) {
	slog.Debug("Querying snap confinement", "snap", s.name)

	snap, _, err := s.client.FindOne(s.name)
	if err != nil {
		return false, fmt.Errorf("failed to find snap: %w", err)
	}

	channel, ok := snap.Channels[s.channel]
	if ok {
		return channel.Confinement == "classic", nil
	}

	return snap.Confinement == "classic", nil
}

// Tracking reports which channel an installed snap is tracking.
func (s *Snap) Tracking() (string, error) {
	slog.Debug("Querying snap channel tracking", "snap", s.name)

	snap, _, err := s.client.Snap(s.name)
	if err != nil {
		return "", fmt.Errorf("failed to find snap: %w", err)
	}

	if snap.Status == client.StatusActive {
		return snap.TrackingChannel, nil
	} else {
		return "", fmt.Errorf("snap '%s' is not installed", s.name)
	}
}
