package handlers

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/runner"
)

func setupHandler(preset string) (*runner.TestRunner, *JujuHandler, error) {
	var err error
	var cfg *config.Config
	var provider providers.Provider

	runner := runner.NewTestRunner()
	runner.SetNextReturn([]byte("not found"), fmt.Errorf("Test error"))

	cfg, err = config.Preset(preset)
	if err != nil {
		return nil, nil, err
	}

	switch preset {
	case "machine":
		provider = providers.NewLXD(runner, cfg)
	case "k8s":
		provider = providers.NewMicroK8s(runner, cfg)
	}

	return runner, NewJujuHandler(cfg, runner, []providers.Provider{provider}), nil
}

func TestJujuHandlerCommandsMicroK8s(t *testing.T) {
	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", path)

	type test struct {
		preset           string
		expectedCommands []string
		expectedDirs     []string
	}

	tests := []test{
		{
			preset: "machine",
			expectedCommands: []string{
				"sudo -u test-user juju show-controller concierge-lxd",
				"sudo -u test-user -g lxd juju bootstrap localhost concierge-lxd --verbose --model-default automatically-retry-hooks=false --model-default test-mode=true",
				"sudo -u test-user juju add-model -c concierge-lxd testing",
			},
			expectedDirs: []string{".local/share/juju"},
		},
		{
			preset: "k8s",
			expectedCommands: []string{
				"sudo -u test-user juju show-controller concierge-microk8s",
				"sudo -u test-user -g snap_microk8s juju bootstrap microk8s concierge-microk8s --verbose --model-default automatically-retry-hooks=false --model-default test-mode=true",
				"sudo -u test-user juju add-model -c concierge-microk8s testing",
			},
			expectedDirs: []string{".local/share/juju"},
		},
	}

	for _, tc := range tests {
		runner, handler, err := setupHandler(tc.preset)
		if err != nil {
			t.Fatal(err.Error())
		}

		err = handler.Prepare()
		if err != nil {
			t.Fatal(err.Error())
		}

		if !reflect.DeepEqual(tc.expectedCommands, runner.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expectedCommands, runner.ExecutedCommands)
		}
		if !reflect.DeepEqual(tc.expectedDirs, runner.CreatedDirectories) {
			t.Fatalf("expected: %v, got: %v", tc.expectedDirs, runner.CreatedDirectories)
		}
	}

}

func TestJujuRestore(t *testing.T) {
	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", path)

	runner, handler, err := setupHandler("machine")
	if err != nil {
		t.Fatal(err.Error())
	}

	handler.Restore()

	expectedDeleted := []string{".local/share/juju"}

	if !reflect.DeepEqual(expectedDeleted, runner.Deleted) {
		t.Fatalf("expected: %v, got: %v", expectedDeleted, runner.Deleted)
	}
}
