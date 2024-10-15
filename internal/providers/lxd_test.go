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

	runner := runner.NewRunner(false)

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
	type test struct {
		testFunc func(l *LXD)
		expected []string
	}

	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	os.Setenv("PATH", "")

	tests := []test{
		{
			func(l *LXD) { l.init() },
			[]string{
				"lxd waitready",
				"lxd init --minimal",
			},
		},
		{
			func(l *LXD) { l.enableNonRootUserControl() },
			[]string{
				"chmod a+wr /var/snap/lxd/common/lxd/unix.socket",
				"usermod -a -G lxd root",
			},
		},
		{
			func(l *LXD) { l.deconflictFirewall() },
			[]string{
				"iptables -F FORWARD",
				"iptables -P FORWARD ACCEPT",
			},
		},
	}

	config := &config.Config{}

	for _, tc := range tests {
		runner := runner.NewTestRunner()
		tc.testFunc(NewLXD(runner, config))

		if !reflect.DeepEqual(tc.expected, runner.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expected, runner.ExecutedCommands)
		}
	}

	// Reset the PATH variable
	os.Setenv("PATH", path)
}
