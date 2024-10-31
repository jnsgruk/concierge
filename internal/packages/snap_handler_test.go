package packages

import (
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/runner"
)

func TestSnapHandlerCommands(t *testing.T) {
	type test struct {
		testFunc func(s *SnapHandler)
		expected []string
	}

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

	for _, tc := range tests {
		r := runner.NewMockRunner()
		r.MockSnapStoreLookup("charmcraft", "latest/stable", true, true)

		snaps := []*runner.Snap{
			r.NewSnap("charmcraft", "latest/stable"),
			r.NewSnap("jq", "latest/stable"),
			r.NewSnap("microk8s", "1.30-strict/stable"),
		}

		tc.testFunc(NewSnapHandler(r, snaps))

		if !reflect.DeepEqual(tc.expected, r.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expected, r.ExecutedCommands)
		}
	}

}
