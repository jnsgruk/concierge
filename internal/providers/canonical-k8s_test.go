package providers

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/runnertest"
)

var defaultFeatureConfig = map[string]map[string]string{
	"load-balancer": {
		"l2-mode": "true",
		"cidrs":   "10.43.45.1/32",
	},
	"local-storage": {},
}

func TestNewCanonicalK8s(t *testing.T) {
	type test struct {
		config   *config.Config
		expected *CanonicalK8s
	}

	noOverrides := &config.Config{}

	channelInConfig := &config.Config{}
	channelInConfig.Providers.CanonicalK8s.Channel = "1.31/candidate"

	overrides := &config.Config{}
	overrides.Overrides.CanonicalK8sChannel = "1.32/edge"
	overrides.Providers.CanonicalK8s.Features = defaultFeatureConfig

	runner := runnertest.NewMockRunner()

	tests := []test{
		{
			config:   noOverrides,
			expected: &CanonicalK8s{Channel: "", runner: runner},
		},
		{
			config:   channelInConfig,
			expected: &CanonicalK8s{Channel: "1.31/candidate", runner: runner},
		},
		{
			config:   overrides,
			expected: &CanonicalK8s{Channel: "1.32/edge", Features: defaultFeatureConfig, runner: runner},
		},
	}

	for _, tc := range tests {
		ck8s := NewCanonicalK8s(runner, tc.config)

		// Check the constructed snaps are correct
		if ck8s.snaps[0].Channel() != tc.expected.Channel {
			t.Fatalf("expected: %v, got: %v", ck8s.snaps[0].Channel(), tc.expected.Channel)
		}

		// Remove the snaps so the rest of the object can be compared
		ck8s.snaps = nil
		if !reflect.DeepEqual(tc.expected, ck8s) {
			t.Fatalf("expected: %v, got: %v", tc.expected, ck8s)
		}
	}
}

func TestCanonicalK8sPrepareCommands(t *testing.T) {
	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	defer os.Setenv("PATH", path)
	os.Setenv("PATH", "")

	config := &config.Config{}
	config.Providers.CanonicalK8s.Channel = ""
	config.Providers.CanonicalK8s.Features = defaultFeatureConfig

	expectedCommands := []string{
		"snap install k8s",
		"k8s bootstrap",
		"k8s status --wait-ready",
		"k8s set load-balancer.l2-mode=true",
		"k8s set load-balancer.cidrs=10.43.45.1/32",
		"k8s enable load-balancer",
		"k8s enable local-storage",
		"k8s kubectl config view --raw",
	}

	expectedFiles := map[string]string{
		".kube/config": "",
	}

	runner := runnertest.NewMockRunner()
	ck8s := NewCanonicalK8s(runner, config)

	// Override the snaps with fake ones that don't call the snapd socket.
	ck8s.snaps = []packages.SnapPackage{
		runnertest.NewTestSnap("k8s", "", false, false),
	}

	ck8s.Prepare()

	slices.Sort(expectedCommands)
	slices.Sort(runner.ExecutedCommands)

	if !reflect.DeepEqual(expectedCommands, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expectedCommands, runner.ExecutedCommands)
	}

	if !reflect.DeepEqual(expectedFiles, runner.CreatedFiles) {
		t.Fatalf("expected: %v, got: %v", expectedFiles, runner.CreatedFiles)
	}
}

func TestCanonicalK8sRestore(t *testing.T) {
	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	defer os.Setenv("PATH", path)
	os.Setenv("PATH", "")

	config := &config.Config{}
	config.Providers.CanonicalK8s.Channel = ""
	config.Providers.CanonicalK8s.Features = defaultFeatureConfig

	runner := runnertest.NewMockRunner()
	ck8s := NewCanonicalK8s(runner, config)
	ck8s.Restore()

	expectedDeleted := []string{".kube"}

	if !reflect.DeepEqual(expectedDeleted, runner.Deleted) {
		t.Fatalf("expected: %v, got: %v", expectedDeleted, runner.Deleted)
	}

	expectedCommands := []string{
		"snap remove k8s --purge",
		"snap remove kubectl --purge",
	}

	if !reflect.DeepEqual(expectedCommands, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expectedCommands, runner.ExecutedCommands)
	}
}
