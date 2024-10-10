package juju

import (
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/providers"
)

func TestNewJuju(t *testing.T) {
	type test struct {
		config   *config.Config
		expected *Juju
	}

	noOverrides := &config.Config{}

	channelInConfig := &config.Config{}
	channelInConfig.Juju.Channel = "4.0/edge"

	overrides := &config.Config{}
	overrides.Overrides.JujuChannel = "3.6/beta"

	tests := []test{
		{
			config: noOverrides,
			expected: &Juju{
				Channel:       "",
				ModelDefaults: nil,
				providers:     []providers.Provider{},
			},
		},
		{
			config: channelInConfig,
			expected: &Juju{
				Channel:       "4.0/edge",
				ModelDefaults: nil,
				providers:     []providers.Provider{},
			},
		},
		{
			config: overrides,
			expected: &Juju{
				Channel:       "3.6/beta",
				ModelDefaults: nil,
				providers:     []providers.Provider{},
			},
		},
	}

	for _, tc := range tests {
		juju := NewJuju(tc.config, []providers.Provider{})
		if !reflect.DeepEqual(tc.expected, juju) {
			t.Fatalf("expected: %v, got: %v", tc.expected, juju)
		}
	}
}
