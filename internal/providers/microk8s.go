package providers

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path"
	"slices"
	"strings"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/runner"
	snapdClient "github.com/snapcore/snapd/client"
)

// Default channel from which MicroK8s is installed when the latest strict
// version cannot be determined.
const defaultChannel = "1.31-strict/stable"

// NewMicroK8s constructs a new MicroK8s provider instance.
func NewMicroK8s(runner *runner.Runner, config *config.Config) *MicroK8s {
	var channel string

	if config.Overrides.MicroK8sChannel != "" {
		channel = config.Overrides.MicroK8sChannel
	} else if config.Providers.MicroK8s.Channel == "" {
		channel = computeDefaultChannel()
	} else {
		channel = config.Providers.MicroK8s.Channel
	}

	return &MicroK8s{
		Channel:   channel,
		Addons:    config.Providers.MicroK8s.Addons,
		bootstrap: config.Providers.MicroK8s.Bootstrap,
		runner:    runner,
	}
}

// MicroK8s represents a MicroK8s install on a given machine.
type MicroK8s struct {
	Channel string
	Addons  []string

	bootstrap bool
	runner    *runner.Runner
}

// Prepare installs and configures MicroK8s such that it can work in testing environments.
// This includes installing the snap, enabling the user who ran concierge to interact
// with MicroK8s without sudo, and sets up the user's kubeconfig file.
func (m *MicroK8s) Prepare() error {
	err := m.init()
	if err != nil {
		return fmt.Errorf("failed to install MicroK8s: %w", err)
	}

	err = m.enableAddons()
	if err != nil {
		return fmt.Errorf("failed to enable MicroK8s addons: %w", err)
	}

	err = m.enableNonRootUserControl()
	if err != nil {
		return fmt.Errorf("failed to enable non-root MicroK8s access: %w", err)
	}

	err = m.setupKubectl()
	if err != nil {
		return fmt.Errorf("failed to setup kubectl for MicroK8s: %w", err)
	}

	slog.Info("Prepared provider", "provider", m.Name())

	return nil
}

// Name reports the name of the provider for Concierge's purposes.
func (m *MicroK8s) Name() string { return "microk8s" }

// Bootstrap reports whether a Juju controller should be bootstrapped onto the provider.
func (m *MicroK8s) Bootstrap() bool { return m.bootstrap }

// CloudName reports the name of the provider as Juju sees it.
func (m *MicroK8s) CloudName() string { return "microk8s" }

// GroupName reports the name of the POSIX group with permission to use MicroK8s.
func (m *MicroK8s) GroupName() string {
	if strings.Contains(m.Channel, "strict") {
		return "snap_microk8s"
	} else {
		return "microk8s"
	}
}

// Snaps reports the snaps required by the MicroK8s provider.
func (m *MicroK8s) Snaps() []*packages.Snap {
	return []*packages.Snap{
		packages.NewSnap("microk8s", m.Channel),
		packages.NewSnap("kubectl", "stable"),
	}
}

// Remove uninstalls MicroK8s and kubectl.
func (m *MicroK8s) Restore() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	err = os.RemoveAll(path.Join(home, ".kube"))
	if err != nil {
		return fmt.Errorf("failed to remove '.kube' subdirectory from user's home directory: %w", err)
	}

	slog.Info("Removed provider", "provider", m.Name())

	return nil
}

// init ensures that MicroK8s is installed, minimally configured, and ready.
func (m *MicroK8s) init() error {
	return m.runner.RunCommands(
		runner.NewCommandSudo("snap", []string{"start", "microk8s"}),
		runner.NewCommandSudo("microk8s", []string{"status", "--wait-ready"}),
	)
}

// enableAddons iterates over the specified addons, enabling and configuring them.
func (m *MicroK8s) enableAddons() error {
	for _, addon := range m.Addons {
		enableArg := addon

		// If the addon is MetalLB, add the predefined IP range
		if addon == "metallb" {
			enableArg = "metallb:10.64.140.43-10.64.140.49"
		}

		cmd := runner.NewCommandSudo("microk8s", []string{"enable", enableArg})
		_, err := m.runner.Run(cmd)
		if err != nil {
			return fmt.Errorf("failed to enable MicroK8s addon '%s': %w", addon, err)
		}
	}

	return nil
}

// enableNonRootUserControl ensures the current user is in the correct POSIX group
// that allows them to interact with MicroK8s.
func (m *MicroK8s) enableNonRootUserControl() error {
	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not determine current user info: %w", err)
	}

	cmd := runner.NewCommandSudo("usermod", []string{"-a", "-G", m.GroupName(), user.Username})

	_, err = m.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to add user '%s' to group 'microk8s': %w", user.Username, err)
	}

	return nil
}

// setupKubectl both installs the kubectl snap, and writes the relevant kubeconfig
// file to the user's home directory such that kubectl works with MicroK8s.
func (m *MicroK8s) setupKubectl() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	err = os.MkdirAll(path.Join(home, ".kube"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create '.kube' subdirectory in user's home directory: %w", err)
	}

	cmd := runner.NewCommandSudo("microk8s", []string{"config"})
	result, err := m.runner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to fetch MicroK8s configuration: %w", err)
	}

	kubeconfig := path.Join(home, ".kube", "config")
	if err := os.WriteFile(kubeconfig, result.Stdout.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig file: %w", err)
	}

	return nil
}

// Try to compute the "correct" default channel. Concerige prefers that the 'strict'
// variants are installed, so we filter available channels and sort descending by
// version. If the list cannot be retrieved, default to a know good version.
func computeDefaultChannel() string {
	// If the snapd socket doesn't exist on the system, return a default value
	if _, err := os.Stat("/run/snapd.socket"); errors.Is(err, os.ErrNotExist) {
		return defaultChannel
	}

	snap, _, err := snapdClient.New(nil).FindOne("microk8s")
	if err != nil {
		return defaultChannel
	}

	keys := []string{}
	for k := range snap.Channels {
		if strings.Contains(k, "strict") && strings.Contains(k, "stable") {
			keys = append(keys, k)
		}
	}

	slices.Sort(keys)
	slices.Reverse(keys)

	return keys[0]
}
