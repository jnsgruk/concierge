package providers

import (
	"fmt"
	"log/slog"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/runner"
	"gopkg.in/yaml.v3"
)

// NewGoogle constructs a new Google provider instance.
func NewGoogle(runner runner.CommandRunner, config *config.Config) *Google {
	credentialsFile := config.Providers.Google.CredentialsFile
	if config.Overrides.GoogleCredentialFile != "" {
		credentialsFile = config.Overrides.GoogleCredentialFile
	}

	return &Google{
		runner:          runner,
		bootstrap:       config.Providers.Google.Bootstrap,
		credentialsFile: credentialsFile,
		credentials:     map[string]interface{}{},
	}
}

// Google represents a Google cloud to bootstrap.
type Google struct {
	bootstrap       bool
	runner          runner.CommandRunner
	credentialsFile string
	credentials     map[string]interface{}
}

// Prepare installs and configures Google such that it can work in testing environments.
// This includes installing the snap, enabling the user who ran concierge to interact
// with Google without sudo, and deconflicting the firewall rules with docker.
func (l *Google) Prepare() error {
	contents, err := l.runner.ReadFile(l.credentialsFile)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	credentials := make(map[string]interface{})

	err = yaml.Unmarshal(contents, &credentials)
	if err != nil {
		return fmt.Errorf("failed to parse google cloud credentials: %w", err)
	}

	l.credentials = credentials

	slog.Info("Prepared provider", "provider", l.Name())
	return nil
}

// Name reports the name of the provider for Concierge's purposes.
func (l *Google) Name() string { return "google" }

// Bootstrap reports whether a Juju controller should be bootstrapped on Google.
func (l *Google) Bootstrap() bool { return l.bootstrap }

// CloudName reports the name of the provider as Juju sees it.
func (l *Google) CloudName() string { return "google" }

// GroupName reports the name of the POSIX group with permissions over the Google socket.
func (l *Google) GroupName() string { return "" }

// Credentials reports the section of Juju's credentials.yaml for the provider
func (l *Google) Credentials() map[string]interface{} {
	return l.credentials
}

// Remove Google provider.
func (l *Google) Restore() error {
	slog.Info("Restored provider", "provider", l.Name())
	return nil
}
