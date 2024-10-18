package packages

import (
	"os"
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/runnertest"
)

func TestDebHandlerCommands(t *testing.T) {
	type test struct {
		testFunc func(d *DebHandler)
		expected []string
	}

	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	defer os.Setenv("PATH", path)
	os.Setenv("PATH", "")

	tests := []test{
		{
			func(d *DebHandler) { d.Prepare() },
			[]string{
				"apt-get update",
				"apt-get install -y cowsay",
				"apt-get install -y python3-venv",
			},
		},
		{
			func(d *DebHandler) { d.Restore() },
			[]string{
				"apt-get remove -y cowsay",
				"apt-get remove -y python3-venv",
				"apt-get autoremove -y",
			},
		},
	}

	debs := []*Deb{
		NewDeb("cowsay"),
		NewDeb("python3-venv"),
	}

	for _, tc := range tests {
		runner := runnertest.NewMockRunner()
		tc.testFunc(NewDebHandler(runner, debs))

		if !reflect.DeepEqual(tc.expected, runner.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expected, runner.ExecutedCommands)
		}
	}
}
