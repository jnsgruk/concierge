package juju

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/runner"
)

var fakeGoogleCreds = []byte(`auth-type: oauth2
client-email: juju-gce-1-sa@concierge.iam.gserviceaccount.com
client-id: "12345678912345"
private-key: |
  -----BEGIN PRIVATE KEY-----
  deadbeef
  -----END PRIVATE KEY-----
project-id: concierge
`)

func setupHandlerWithPreset(preset string) (*runner.MockRunner, *JujuHandler, error) {
	var err error
	var cfg *config.Config
	var provider providers.Provider

	runner := runner.NewMockRunner()
	runner.MockCommandReturn("sudo -u test-user juju show-controller concierge-lxd", []byte("not found"), fmt.Errorf("Test error"))
	runner.MockCommandReturn("sudo -u test-user juju show-controller concierge-microk8s", []byte("not found"), fmt.Errorf("Test error"))
	runner.MockCommandReturn("sudo -u test-user juju show-controller concierge-k8s", []byte("not found"), fmt.Errorf("Test error"))

	cfg, err = config.Preset(preset)
	if err != nil {
		return nil, nil, err
	}

	switch preset {
	case "machine":
		provider = providers.NewLXD(runner, cfg)
	case "microk8s":
		provider = providers.NewMicroK8s(runner, cfg)
	case "k8s":
		provider = providers.NewK8s(runner, cfg)
	}

	handler := NewJujuHandler(cfg, runner, []providers.Provider{provider})

	return runner, handler, nil
}

func setupHandlerWithGoogleProvider() (*runner.MockRunner, *JujuHandler, error) {
	cfg := &config.Config{}
	cfg.Providers.Google.Enable = true
	cfg.Providers.Google.Bootstrap = true
	cfg.Providers.Google.CredentialsFile = "google.yaml"

	runner := runner.NewMockRunner()
	runner.MockFile("google.yaml", fakeGoogleCreds)

	provider := providers.NewProvider("google", runner, cfg)

	err := provider.Prepare()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare google provider: %w", err)
	}

	handler := NewJujuHandler(cfg, runner, []providers.Provider{provider})
	return runner, handler, nil
}
func TestJujuHandlerCommandsPresets(t *testing.T) {
	type test struct {
		preset           string
		expectedCommands []string
		expectedDirs     []string
	}

	tests := []test{
		{
			preset: "machine",
			expectedCommands: []string{
				"snap install juju",
				"sudo -u test-user juju show-controller concierge-lxd",
				"sudo -u test-user -g lxd juju bootstrap localhost concierge-lxd --verbose --model-default automatically-retry-hooks=false --model-default test-mode=true",
				"sudo -u test-user juju add-model -c concierge-lxd testing",
			},
			expectedDirs: []string{".local/share/juju"},
		},
		{
			preset: "microk8s",
			expectedCommands: []string{
				"snap install juju",
				"sudo -u test-user juju show-controller concierge-microk8s",
				"sudo -u test-user -g snap_microk8s juju bootstrap microk8s concierge-microk8s --verbose --model-default automatically-retry-hooks=false --model-default test-mode=true",
				"sudo -u test-user juju add-model -c concierge-microk8s testing",
			},
			expectedDirs: []string{".local/share/juju"},
		},
		{
			preset: "k8s",
			expectedCommands: []string{
				"snap install juju",
				"sudo -u test-user juju show-controller concierge-k8s",
				"sudo -u test-user juju bootstrap k8s concierge-k8s --verbose --model-default automatically-retry-hooks=false --model-default test-mode=true --bootstrap-constraints root-disk=2G",
				"sudo -u test-user juju add-model -c concierge-k8s testing",
			},
			expectedDirs: []string{".local/share/juju"},
		},
	}

	for _, tc := range tests {
		runner, handler, err := setupHandlerWithPreset(tc.preset)
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
		if len(runner.CreatedFiles) > 0 {
			t.Fatalf("expected no files to be created, got: %v", runner.CreatedFiles)
		}
	}
}

func TestJujuHandlerWithCredentialedProvider(t *testing.T) {
	expectedCredsFileContent := []byte(`credentials:
    google:
        concierge:
            auth-type: oauth2
            client-email: juju-gce-1-sa@concierge.iam.gserviceaccount.com
            client-id: "12345678912345"
            private-key: |
                -----BEGIN PRIVATE KEY-----
                deadbeef
                -----END PRIVATE KEY-----
            project-id: concierge
`)

	runner, handler, err := setupHandlerWithGoogleProvider()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = handler.Prepare()
	if err != nil {
		t.Fatal(err.Error())
	}

	expectedFiles := map[string]string{".local/share/juju/credentials.yaml": string(expectedCredsFileContent)}

	if !reflect.DeepEqual(expectedFiles, runner.CreatedFiles) {
		t.Fatalf("expected: %v, got: %v", expectedFiles, runner.CreatedFiles)
	}
}

func TestJujuRestoreNoKillController(t *testing.T) {
	runner, handler, err := setupHandlerWithPreset("machine")
	if err != nil {
		t.Fatal(err.Error())
	}

	handler.Restore()

	expectedDeleted := []string{".local/share/juju"}
	expectedCommands := []string{"snap remove juju --purge"}

	if !reflect.DeepEqual(expectedDeleted, runner.Deleted) {
		t.Fatalf("expected: %v, got: %v", expectedDeleted, runner.Deleted)
	}

	if !reflect.DeepEqual(expectedCommands, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expectedCommands, runner.ExecutedCommands)
	}
}

func TestJujuRestoreKillController(t *testing.T) {
	runner, handler, err := setupHandlerWithGoogleProvider()
	if err != nil {
		t.Fatal(err.Error())
	}

	handler.Restore()

	expectedDeleted := []string{".local/share/juju"}
	expectedCommands := []string{
		"sudo -u test-user juju show-controller concierge-google",
		"sudo -u test-user juju kill-controller --verbose --no-prompt concierge-google",
		"snap remove juju --purge",
	}

	if !reflect.DeepEqual(expectedDeleted, runner.Deleted) {
		t.Fatalf("expected: %v, got: %v", expectedDeleted, runner.Deleted)
	}

	if !reflect.DeepEqual(expectedCommands, runner.ExecutedCommands) {
		t.Fatalf("expected: %v, got: %v", expectedCommands, runner.ExecutedCommands)
	}
}
