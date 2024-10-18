package packages

import (
	"os"
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/runnertest"
)

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

	snaps := []SnapPackage{
		runnertest.NewTestSnap("charmcraft", "latest/stable", true, true),
		runnertest.NewTestSnap("jq", "latest/stable", false, false),
		runnertest.NewTestSnap("microk8s", "1.30-strict/stable", false, false),
	}

	for _, tc := range tests {
		runner := runnertest.NewMockRunner()
		tc.testFunc(NewSnapHandler(runner, snaps))

		if !reflect.DeepEqual(tc.expected, runner.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expected, runner.ExecutedCommands)
		}
	}

}
