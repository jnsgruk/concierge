package providers

import (
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
)

var defaultAddons []string = []string{
	"hostpath-storage",
	"dns",
	"rbac",
	"metallb:10.64.140.43-10.64.140.49",
}

func TestNewMicroK8s(t *testing.T) {
	type test struct {
		config   *config.Config
		expected *MicroK8s
	}

	noOverrides := &config.Config{}

	channelInConfig := &config.Config{}
	channelInConfig.Providers.MicroK8s.Channel = "1.29-strict/stable"

	overrides := &config.Config{}
	overrides.Overrides.MicroK8sChannel = "1.30/edge"
	overrides.Providers.MicroK8s.Addons = defaultAddons

	runner := runner.NewMockRunner()

	tests := []test{
		{
			config:   noOverrides,
			expected: &MicroK8s{Channel: "1.31-strict/stable", runner: runner},
		},
		{
			config:   channelInConfig,
			expected: &MicroK8s{Channel: "1.29-strict/stable", runner: runner},
		},
		{
			config:   overrides,
			expected: &MicroK8s{Channel: "1.30/edge", Addons: defaultAddons, runner: runner},
		},
	}

	for _, tc := range tests {
		uk8s := NewMicroK8s(runner, tc.config)

		// Check the constructed snaps are correct
		if uk8s.snaps[0].Channel != tc.expected.Channel {
			t.Fatalf("expected: %v, got: %v", uk8s.snaps[0].Channel, tc.expected.Channel)
		}

		// Remove the snaps so the rest of the object can be compared
		uk8s.snaps = nil
		if !reflect.DeepEqual(tc.expected, uk8s) {
			t.Fatalf("expected: %v, got: %v", tc.expected, uk8s)
		}
	}
}

func TestMicroK8sGroupName(t *testing.T) {
	type test struct {
		channel  string
		expected string
	}

	tests := []test{
		{channel: "1.30-strict/stable", expected: "snap_microk8s"},
		{channel: "1.30/stable", expected: "microk8s"},
	}

	for _, tc := range tests {
		config := &config.Config{}
		config.Providers.MicroK8s.Channel = tc.channel
		uk8s := NewMicroK8s(runner.NewMockRunner(), config)

		if !reflect.DeepEqual(tc.expected, uk8s.GroupName()) {
			t.Fatalf("expected: %v, got: %v", tc.expected, uk8s.GroupName())
		}
	}
}

func TestMicroK8sPrepareCommands(t *testing.T) {
	config := &config.Config{}
	config.Providers.MicroK8s.Channel = "1.31-strict/stable"
	config.Providers.MicroK8s.Addons = defaultAddons

	expectedCommands := []string{
		"snap install microk8s --channel 1.31-strict/stable",
		"snap install kubectl --channel stable",
		"microk8s status --wait-ready",
		"microk8s enable hostpath-storage",
		"microk8s enable dns",
		"microk8s enable rbac",
		"microk8s enable metallb:10.64.140.43-10.64.140.49",
		"usermod -a -G snap_microk8s test-user",
		"microk8s config",
	}

	expectedFiles := map[string]string{
		".kube/config": "",
	}

	runner := runner.NewMockRunner()
	uk8s := NewMicroK8s(runner, config)
	uk8s.Prepare()

	if !reflect.DeepEqual(expectedCommands, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expectedCommands, runner.ExecutedCommands)
	}

	if !reflect.DeepEqual(expectedFiles, runner.CreatedFiles) {
		t.Fatalf("expected: %v, got: %v", expectedFiles, runner.CreatedFiles)
	}
}

func TestMicroK8sRestore(t *testing.T) {
	config := &config.Config{}
	config.Providers.MicroK8s.Channel = "1.31-strict/stable"
	config.Providers.MicroK8s.Addons = defaultAddons

	runner := runner.NewMockRunner()
	uk8s := NewMicroK8s(runner, config)
	uk8s.Restore()

	expectedDeleted := []string{".kube"}

	if !reflect.DeepEqual(expectedDeleted, runner.Deleted) {
		t.Fatalf("expected: %v, got: %v", expectedDeleted, runner.Deleted)
	}

	expectedCommands := []string{
		"snap remove microk8s --purge",
		"snap remove kubectl --purge",
	}

	if !reflect.DeepEqual(expectedCommands, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expectedCommands, runner.ExecutedCommands)
	}
}
