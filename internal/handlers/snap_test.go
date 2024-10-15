package handlers

import (
	"os"
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/runnertest"
)

func NewTestSnap(name, channel string, classic bool, installed bool) *TestSnap {
	return &TestSnap{name: name, channel: channel, classic: classic, installed: installed}
}

type TestSnap struct {
	name      string
	channel   string
	classic   bool
	installed bool
}

func (ts *TestSnap) Name() string              { return ts.name }
func (ts *TestSnap) Classic() (bool, error)    { return ts.classic, nil }
func (ts *TestSnap) Installed() bool           { return ts.installed }
func (ts *TestSnap) Tracking() (string, error) { return ts.channel, nil }
func (ts *TestSnap) Channel() string           { return ts.channel }
func (ts *TestSnap) SetChannel(c string)       { ts.channel = c }

func TestSnapHandlerCommands(t *testing.T) {
	type test struct {
		testFunc func(s *SnapHandler)
		expected []string
	}

	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	defer os.Setenv("PATH", path)
	os.Setenv("PATH", "")
	// Reset the PATH variable

	tests := []test{
		{
			func(s *SnapHandler) { s.Prepare() },
			[]string{
				"snap refresh charmcraft --channel latest/stable --classic",
				"snap install jq --channel latest/stable",
				"snap install microk8s --channel 1.30-strict/stable",
			},
		},
		{
			func(s *SnapHandler) { s.Restore() },
			[]string{
				"snap remove charmcraft --purge",
				"snap remove jq --purge",
				"snap remove microk8s --purge",
			},
		},
	}

	snaps := []packages.SnapPackage{
		NewTestSnap("charmcraft", "latest/stable", true, true),
		NewTestSnap("jq", "latest/stable", false, false),
		NewTestSnap("microk8s", "1.30-strict/stable", false, false),
	}

	for _, tc := range tests {
		runner := runnertest.NewMockRunner()
		tc.testFunc(NewSnapHandler(runner, snaps))

		if !reflect.DeepEqual(tc.expected, runner.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expected, runner.ExecutedCommands)
		}
	}

}
