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

	runner := runner.NewRunner(false)

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
		uk8s := NewMicroK8s(runner.NewRunner(false), config)

		if !reflect.DeepEqual(tc.expected, uk8s.GroupName()) {
			t.Fatalf("expected: %v, got: %v", tc.expected, uk8s.GroupName())
		}
	}
}
