package providers

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path"
	"slices"
	"strings"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
	"github.com/jnsgruk/concierge/internal/snap"
	snapdClient "github.com/snapcore/snapd/client"
)

// NewMicroK8s constructs a new MicroK8s provider instance.
func NewMicroK8s(config *config.Config) *MicroK8s {
	var channel string

	if config.Overrides.MicroK8s != "" {
		channel = config.Overrides.MicroK8s
	} else if config.Providers.MicroK8s.Channel == "" {
		channel = computeDefaultChannel()
	} else {
		channel = config.Providers.MicroK8s.Channel
	}

	return &MicroK8s{
		Channel: channel,
		Addons:  config.Providers.MicroK8s.Addons,
	}
}

// MicroK8s represents a MicroK8s install on a given machine.
type MicroK8s struct {
	Channel string
	Addons  []string
}

// Init installs and configures MicroK8s such that it can work in testing environments.
// This includes installing the snap, enabling the user who ran concierge to interact
// with MicroK8s without sudo, and sets up the user's kubeconfig file.
func (m *MicroK8s) Init() error {
	err := m.install()
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

	slog.Info("Initialised provider", "provider", m.Name())

	return nil
}

// Name reports the name of the provider for Concierge's purposes.
func (m *MicroK8s) Name() string {
	return "microk8s"
}

// CloudName reports the name of the provider as Juju sees it.
func (m *MicroK8s) CloudName() string {
	return "microk8s"
}

// GroupName reports the name of the POSIX group with permission to use MicroK8s.
func (m *MicroK8s) GroupName() string {
	if strings.Contains(m.Channel, "strict") {
		return "snap_microk8s"
	} else {
		return "microk8s"
	}
}

// install ensures that MicroK8s is installed, minimally configured, and ready.
func (m *MicroK8s) install() error {
	err := snap.NewSnap("microk8s", m.Channel).Install()
	if err != nil {
		return err
	}

	if err := runner.RunCommands(
		runner.NewCommandSudo("snap", []string{"start", "microk8s"}),
		runner.NewCommandSudo("microk8s", []string{"status", "--wait-ready"}),
	); err != nil {
		return err
	}

	return nil
}

// enableAddons iterates over the specified addons, enabling and configuring them.
func (m *MicroK8s) enableAddons() error {
	for _, addon := range m.Addons {
		enableArg := addon

		// If the addon is MetalLB, add the predefined IP range
		if addon == "metallb" {
			enableArg = "metallb:10.64.140.43-10.64.140.49"
		}

		_, err := runner.NewCommandSudo("microk8s", []string{"enable", enableArg}).Run()
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

	_, err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add user '%s' to group 'microk8s': %w", user.Username, err)
	}

	return nil
}

// setupKubectl both installs the kubectl snap, and writes the relevant kubeconfig
// file to the user's home directory such that kubectl works with MicroK8s.
func (m *MicroK8s) setupKubectl() error {
	err := snap.NewSnapFromString("kubectl").Install()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	err = os.MkdirAll(path.Join(home, ".kube"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create '.kube' subdirectory in user's home directory: %w", err)
	}

	result, err := runner.NewCommandSudo("microk8s", []string{"config"}).Run()
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
	snap, _, err := snapdClient.New(nil).FindOne("microk8s")
	if err != nil {
		return "1.30-strict/stable"
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
