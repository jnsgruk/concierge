package providers

import (
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
)

func TestNewLXD(t *testing.T) {
	type test struct {
		config   *config.Config
		expected *LXD
	}

	noOverrides := &config.Config{}

	channelInConfig := &config.Config{}
	channelInConfig.Providers.LXD.Channel = "latest/edge"

	overrides := &config.Config{}
	overrides.Overrides.LXDChannel = "5.20/stable"

	runner := runner.NewMockRunner()

	tests := []test{
		{config: noOverrides, expected: &LXD{Channel: "", runner: runner}},
		{config: channelInConfig, expected: &LXD{Channel: "latest/edge", runner: runner}},
		{config: overrides, expected: &LXD{Channel: "5.20/stable", runner: runner}},
	}

	for _, tc := range tests {
		lxd := NewLXD(runner, tc.config)

		// Check the constructed snaps are correct
		if lxd.snaps[0].Channel != tc.expected.Channel {
			t.Fatalf("expected: %v, got: %v", lxd.snaps[0].Channel, tc.expected.Channel)
		}

		// Remove the snaps so the rest of the object can be compared
		lxd.snaps = nil
		if !reflect.DeepEqual(tc.expected, lxd) {
			t.Fatalf("expected: %v, got: %v", tc.expected, lxd)
		}
	}
}

func TestLXDPrepareCommands(t *testing.T) {
	config := &config.Config{}

	expected := []string{
		"snap install lxd",
		"lxd waitready",
		"lxd init --minimal",
		"lxc network set lxdbr0 ipv6.address none",
		"chmod a+wr /var/snap/lxd/common/lxd/unix.socket",
		"usermod -a -G lxd test-user",
		"iptables -F FORWARD",
		"iptables -P FORWARD ACCEPT",
	}

	runner := runner.NewMockRunner()
	lxd := NewLXD(runner, config)
	lxd.Prepare()

	if !reflect.DeepEqual(expected, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expected, runner.ExecutedCommands)
	}
}

func TestLXDRestore(t *testing.T) {
	config := &config.Config{}

	runner := runner.NewMockRunner()
	lxd := NewLXD(runner, config)
	lxd.Restore()

	expectedCommands := []string{"snap remove lxd --purge"}

	if !reflect.DeepEqual(expectedCommands, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expectedCommands, runner.ExecutedCommands)
	}
}
