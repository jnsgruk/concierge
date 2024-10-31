package providers

import (
	"fmt"
	"log/slog"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/system"
)

// NewLXD constructs a new LXD provider instance.
func NewLXD(r system.Worker, config *config.Config) *LXD {
	var channel string
	if config.Overrides.LXDChannel != "" {
		channel = config.Overrides.LXDChannel
	} else {
		channel = config.Providers.LXD.Channel
	}

	return &LXD{
		Channel:   channel,
		system:    r,
		bootstrap: config.Providers.LXD.Bootstrap,
		snaps:     []*system.Snap{{Name: "lxd", Channel: channel}},
	}
}

// LXD represents a LXD install on a given machine.
type LXD struct {
	Channel string

	bootstrap bool
	system    system.Worker
	snaps     []*system.Snap
}

// Prepare installs and configures LXD such that it can work in testing environments.
// This includes installing the snap, enabling the user who ran concierge to interact
// with LXD without sudo, and deconflicting the firewall rules with docker.
func (l *LXD) Prepare() error {
	err := l.install()
	if err != nil {
		return fmt.Errorf("failed to install LXD: %w", err)
	}

	err = l.init()
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

// Credentials reports the section of Juju's credentials.yaml for the provider
func (l *LXD) Credentials() map[string]interface{} { return nil }

// Remove uninstalls LXD.
func (l *LXD) Restore() error {
	snapHandler := packages.NewSnapHandler(l.system, l.snaps)

	err := snapHandler.Restore()
	if err != nil {
		return err
	}

	slog.Info("Restored provider", "provider", l.Name())
	return nil
}

// install ensures that LXD is installed.
func (l *LXD) install() error {
	snapHandler := packages.NewSnapHandler(l.system, l.snaps)

	err := snapHandler.Prepare()
	if err != nil {
		return err
	}

	return nil
}

// init ensures that LXD is minimally configured, and ready.
func (l *LXD) init() error {
	return l.system.RunMany(
		system.NewCommand("lxd", []string{"waitready"}),
		system.NewCommand("lxd", []string{"init", "--minimal"}),
		system.NewCommand("lxc", []string{"network", "set", "lxdbr0", "ipv6.address", "none"}),
	)
}

// enableNonRootUserControl ensures the current user is in the `lxd` group.
func (l *LXD) enableNonRootUserControl() error {
	username := l.system.User().Username

	return l.system.RunMany(
		system.NewCommand("chmod", []string{"a+wr", "/var/snap/lxd/common/lxd/unix.socket"}),
		system.NewCommand("usermod", []string{"-a", "-G", "lxd", username}),
	)
}

// deconflictFirewall ensures that LXD containers can talk out to the internet.
// This is to avoid a conflict with the default iptables rules that ship with
// docker on Ubuntu.
func (l *LXD) deconflictFirewall() error {
	return l.system.RunMany(
		system.NewCommand("iptables", []string{"-F", "FORWARD"}),
		system.NewCommand("iptables", []string{"-P", "FORWARD", "ACCEPT"}),
	)
}
