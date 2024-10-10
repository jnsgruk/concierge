package providers

import (
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
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

	tests := []test{
		{
			config:   noOverrides,
			expected: &MicroK8s{Channel: "1.31-strict/stable"},
		},
		{
			config:   channelInConfig,
			expected: &MicroK8s{Channel: "1.29-strict/stable"},
		},
		{
			config:   overrides,
			expected: &MicroK8s{Channel: "1.30/edge", Addons: defaultAddons},
		},
	}

	for _, tc := range tests {
		uk8s := NewMicroK8s(tc.config)
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
		uk8s := NewMicroK8s(config)

		if !reflect.DeepEqual(tc.expected, uk8s.GroupName()) {
			t.Fatalf("expected: %v, got: %v", tc.expected, uk8s.GroupName())
		}
	}
}

// func TestRunCommands(t *testing.T) {
// 	if err := RunCommands(
// 		NewCommand("go", []string{"env", "GOBIN"}),
// 		NewCommand("go", []string{"env", "GOPATH"}),
// 	); err != nil {
// 		t.Fatalf("expected commands to succeed; got error: %s", err.Error())
// 	}

// 	if err := RunCommands(
// 		NewCommand("go", []string{"env", "GOBIN"}),
// 		NewCommand("gopp", []string{"env", "GOPATH"}),
// 	); err == nil {
// 		t.Fatalf("expected commands to fail; but they succeeded")
// 	}
// }
