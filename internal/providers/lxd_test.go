package providers

import (
	"os"
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

	runner := runner.NewTestRunner()

	tests := []test{
		{config: noOverrides, expected: &LXD{Channel: "", runner: runner}},
		{config: channelInConfig, expected: &LXD{Channel: "latest/edge", runner: runner}},
		{config: overrides, expected: &LXD{Channel: "5.20/stable", runner: runner}},
	}

	for _, tc := range tests {
		lxd := NewLXD(runner, tc.config)
		if !reflect.DeepEqual(tc.expected, lxd) {
			t.Fatalf("expected: %v, got: %v", tc.expected, lxd)
		}
	}
}

func TestLXDPrepareCommands(t *testing.T) {
	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	defer os.Setenv("PATH", path)
	os.Setenv("PATH", "")

	config := &config.Config{}

	expected := []string{
		"lxd waitready",
		"lxd init --minimal",
		"lxc network set lxdbr0 ipv6.address none",
		"chmod a+wr /var/snap/lxd/common/lxd/unix.socket",
		"usermod -a -G lxd test-user",
		"iptables -F FORWARD",
		"iptables -P FORWARD ACCEPT",
	}

	runner := runner.NewTestRunner()
	lxd := NewLXD(runner, config)
	lxd.Prepare()

	if !reflect.DeepEqual(expected, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expected, runner.ExecutedCommands)
	}
}
