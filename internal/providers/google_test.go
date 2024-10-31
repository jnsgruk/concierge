package providers

import (
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
	"gopkg.in/yaml.v3"
)

func TestNewGoogle(t *testing.T) {
	type test struct {
		config   *config.Config
		expected *Google
	}

	noOverrides := &config.Config{}

	credsInConfig := &config.Config{}
	credsInConfig.Providers.Google.CredentialsFile = "/home/ubuntu/credentials.yaml"

	overrides := &config.Config{}
	overrides.Overrides.GoogleCredentialFile = "/home/ubuntu/alternate-credentials.yaml"

	runner := runner.NewMockRunner()

	tests := []test{
		{
			config: noOverrides,
			expected: &Google{
				runner:      runner,
				credentials: map[string]interface{}{},
			},
		},
		{
			config: credsInConfig,
			expected: &Google{
				runner:          runner,
				credentialsFile: "/home/ubuntu/credentials.yaml",
				credentials:     map[string]interface{}{},
			},
		},
		{
			config: overrides,
			expected: &Google{
				runner:          runner,
				credentialsFile: "/home/ubuntu/alternate-credentials.yaml",
				credentials:     map[string]interface{}{},
			},
		},
	}

	for _, tc := range tests {
		uk8s := NewGoogle(runner, tc.config)
		if !reflect.DeepEqual(tc.expected, uk8s) {
			t.Fatalf("expected: %v, got: %v", tc.expected, uk8s)
		}
	}
}

func TestGooglePrepareCommands(t *testing.T) {
	config := &config.Config{}
	config.Providers.Google.CredentialsFile = "/home/ubuntu/credentials.yaml"

	runner := runner.NewMockRunner()
	uk8s := NewGoogle(runner, config)
	uk8s.Prepare()

	if len(runner.ExecutedCommands) != 0 {
		t.Fatalf("expected no commands to have been run")
	}

	if len(runner.CreatedFiles) != 0 {
		t.Fatalf("expected no files to have been created")
	}
}

func TestGoogleReadCredentials(t *testing.T) {
	config := &config.Config{}
	config.Providers.Google.CredentialsFile = "credentials.yaml"

	runner := runner.NewMockRunner()

	creds := []byte(`auth-type: oauth2
client-email: juju-gce-1-sa@concierge.iam.gserviceaccount.com
client-id: "12345678912345"
private-key: |
  -----BEGIN PRIVATE KEY-----
  deadbeef
  -----END PRIVATE KEY-----
project-id: concierge
`)

	fakeCredsMarshalled := make(map[string]interface{})
	err := yaml.Unmarshal(creds, &fakeCredsMarshalled)
	if err != nil {
		t.Fatal(err)
	}

	runner.MockFile("credentials.yaml", creds)

	google := NewGoogle(runner, config)
	google.Prepare()

	if !reflect.DeepEqual(google.Credentials(), fakeCredsMarshalled) {
		t.Fatalf("expected: %v, got: %v", fakeCredsMarshalled, google.Credentials())
	}
}
