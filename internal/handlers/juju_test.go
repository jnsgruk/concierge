package handlers

import (
	"fmt"
	"os"
	"os/user"
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/runner"
)

func fakeSudoEnv() (string, func()) {
	// Prevent the path of the test machine interfering with the test results.
	path := os.Getenv("PATH")
	os.Setenv("PATH", "")
	// Fake a sudo user
	user, _ := user.Current()
	os.Setenv("SUDO_USER", user.Username)

	return user.Username, func() {
		os.Setenv("PATH", path)
		os.Setenv("SUDO_USER", "")
	}
}

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
	username, reset := fakeSudoEnv()
	defer reset()

	type test struct {
		preset   string
		expected []string
	}

	tests := []test{
		{
			preset: "machine",
			expected: []string{
				fmt.Sprintf("sudo -u %s juju show-controller concierge-lxd", username),
				fmt.Sprintf("sudo -u %s -g lxd juju bootstrap localhost concierge-lxd --verbose --model-default automatically-retry-hooks=false --model-default test-mode=true", username),
				fmt.Sprintf("sudo -u %s juju add-model -c concierge-lxd testing", username),
			},
		},
		{
			preset: "k8s",
			expected: []string{
				fmt.Sprintf("sudo -u %s juju show-controller concierge-microk8s", username),
				fmt.Sprintf("sudo -u %s -g snap_microk8s juju bootstrap microk8s concierge-microk8s --verbose --model-default automatically-retry-hooks=false --model-default test-mode=true", username),
				fmt.Sprintf("sudo -u %s juju add-model -c concierge-microk8s testing", username),
			},
		},
	}

	for _, tc := range tests {
		runner, handler, err := setupHandler(tc.preset)
		if err != nil {
			t.Fatal(err.Error())
		}

		err = handler.bootstrap()
		if err != nil {
			t.Fatal(err.Error())
		}

		if !reflect.DeepEqual(tc.expected, runner.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expected, runner.ExecutedCommands)
		}
	}

}
