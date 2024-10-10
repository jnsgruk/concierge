package providers

import (
	"fmt"
	"log/slog"
	"os/user"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
	"github.com/jnsgruk/concierge/internal/snap"
)

// NewLXD constructs a new LXD provider instance.
func NewLXD(config *config.Config) *LXD {
	var channel string
	if config.Overrides.LXDChannel != "" {
		channel = config.Overrides.LXDChannel
	} else {
		channel = config.Providers.LXD.Channel
	}

	return &LXD{Channel: channel}
}

// LXD represents a LXD install on a given machine.
type LXD struct {
	Channel string
}

// Init installs and configures LXD such that it can work in testing environments.
// This includes installing the snap, enabling the user who ran concierge to interact
// with LXD without sudo, and deconflicting the firewall rules with docker.
func (l *LXD) Init() error {
	err := l.install()
	if err != nil {
		return fmt.Errorf("failed to install LXD: %w", err)
	}

	err = l.enableNonRootUserControl()
	if err != nil {
		return fmt.Errorf("failed to enable non-root LXD access: %w", err)
	}

	err = l.deconflictFirewall()
	if err != nil {
		return fmt.Errorf("failed to adjust firewall rules for LXD: %w", err)
	}

	slog.Info("Initialised provider", "provider", l.Name())

	return nil
}

// Name reports the name of the provider for Concierge's purposes.
func (l *LXD) Name() string {
	return "lxd"
}

// CloudName reports the name of the provider as Juju sees it.
func (l *LXD) CloudName() string {
	return "localhost"
}

// GroupName reports the name of the POSIX group with permissions over the LXD socket.
func (l *LXD) GroupName() string {
	return "lxd"
}

// Remove uninstalls LXD.
func (l *LXD) Remove() error {
	err := snap.NewSnapFromString("lxd").Remove(true)
	if err != nil {
		return err
	}

	slog.Info("Removed provider", "provider", l.Name())

	return nil
}

// install ensures that LXD is installed, minimally configured, and ready.
func (l *LXD) install() error {
	err := snap.NewSnap("lxd", l.Channel).Install()
	if err != nil {
		return err
	}

	if err = runner.RunCommands(
		runner.NewCommandSudo("lxd", []string{"waitready"}),
		runner.NewCommandSudo("lxd", []string{"init", "--minimal"}),
	); err != nil {
		return err
	}

	return nil
}

// enableNonRootUserControl ensures the current user is in the `lxd` group.
func (l *LXD) enableNonRootUserControl() error {
	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not determine current user info: %w", err)
	}

	if err = runner.RunCommands(
		runner.NewCommandSudo("chmod", []string{"a+wr", "/var/snap/lxd/common/lxd/unix.socket"}),
		runner.NewCommandSudo("usermod", []string{"-a", "-G", "lxd", user.Username}),
	); err != nil {
		return err
	}

	return nil
}

// deconflictFirewall ensures that LXD containers can talk out to the internet.
// This is to avoid a conflict with the default iptables rules that ship with
// docker on Ubuntu.
func (l *LXD) deconflictFirewall() error {
	if err := runner.RunCommands(
		runner.NewCommandSudo("iptables", []string{"-F", "FORWARD"}),
		runner.NewCommandSudo("iptables", []string{"-P", "FORWARD", "ACCEPT"}),
	); err != nil {
		return err
	}

	return nil
}
