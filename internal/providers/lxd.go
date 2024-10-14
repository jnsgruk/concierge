package providers

import (
	"fmt"
	"log/slog"
	"os/user"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/runner"
)

// NewLXD constructs a new LXD provider instance.
func NewLXD(runner *runner.Runner, config *config.Config) *LXD {
	var channel string
	if config.Overrides.LXDChannel != "" {
		channel = config.Overrides.LXDChannel
	} else {
		channel = config.Providers.LXD.Channel
	}

	return &LXD{
		Channel:   channel,
		runner:    runner,
		bootstrap: config.Providers.LXD.Bootstrap,
	}
}

// LXD represents a LXD install on a given machine.
type LXD struct {
	Channel string

	bootstrap bool
	runner    *runner.Runner
}

// Prepare installs and configures LXD such that it can work in testing environments.
// This includes installing the snap, enabling the user who ran concierge to interact
// with LXD without sudo, and deconflicting the firewall rules with docker.
func (l *LXD) Prepare() error {
	for _, snap := range l.Snaps() {
		if !snap.Installed() {
			return fmt.Errorf("snap '%s' not installed and is required by LXD", snap.Name)
		}
	}

	err := l.init()
	if err != nil {
		return fmt.Errorf("failed to initialise LXD: %w", err)
	}

	err = l.enableNonRootUserControl()
	if err != nil {
		return fmt.Errorf("failed to enable non-root LXD access: %w", err)
	}

	err = l.deconflictFirewall()
	if err != nil {
		return fmt.Errorf("failed to adjust firewall rules for LXD: %w", err)
	}

	slog.Info("Prepared provider", "provider", l.Name())
	return nil
}

// Name reports the name of the provider for Concierge's purposes.
func (l *LXD) Name() string { return "lxd" }

// Bootstrap reports whether a Juju controller should be bootstrapped on LXD.
func (l *LXD) Bootstrap() bool { return l.bootstrap }

// CloudName reports the name of the provider as Juju sees it.
func (l *LXD) CloudName() string { return "localhost" }

// GroupName reports the name of the POSIX group with permissions over the LXD socket.
func (l *LXD) GroupName() string { return "lxd" }

// Snaps reports the snaps required by the LXD provider.
func (l *LXD) Snaps() []*packages.Snap {
	return []*packages.Snap{packages.NewSnap("lxd", l.Channel)}
}

// Remove uninstalls LXD.
func (l *LXD) Restore() error {
	slog.Info("Restored provider", "provider", l.Name())
	return nil
}

// init ensures that LXD is installed, minimally configured, and ready.
func (l *LXD) init() error {
	return l.runner.RunCommands(
		runner.NewCommandSudo("lxd", []string{"waitready"}),
		runner.NewCommandSudo("lxd", []string{"init", "--minimal"}),
	)
}

// enableNonRootUserControl ensures the current user is in the `lxd` group.
func (l *LXD) enableNonRootUserControl() error {
	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not determine current user info: %w", err)
	}

	return l.runner.RunCommands(
		runner.NewCommandSudo("chmod", []string{"a+wr", "/var/snap/lxd/common/lxd/unix.socket"}),
		runner.NewCommandSudo("usermod", []string{"-a", "-G", "lxd", user.Username}),
	)
}

// deconflictFirewall ensures that LXD containers can talk out to the internet.
// This is to avoid a conflict with the default iptables rules that ship with
// docker on Ubuntu.
func (l *LXD) deconflictFirewall() error {
	return l.runner.RunCommands(
		runner.NewCommandSudo("iptables", []string{"-F", "FORWARD"}),
		runner.NewCommandSudo("iptables", []string{"-P", "FORWARD", "ACCEPT"}),
	)
}
