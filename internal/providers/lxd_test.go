package providers

import (
	"reflect"
	"testing"

	"github.com/jnsgruk/concierge/internal/config"
)

func TestNewLXD(t *testing.T) {
	type test struct {
		config   *config.Config
		expected *LXD
	}

	noOverrides := &config.Config{}

	channelInConfig := &config.Config{}
	channelInConfig.Providers.LXD.Channel = "latest/edge"

	overrides := &config.Config{}
	overrides.Overrides.LXDChannel = "5.20/stable"

	tests := []test{
		{config: noOverrides, expected: &LXD{Channel: ""}},
		{config: channelInConfig, expected: &LXD{Channel: "latest/edge"}},
		{config: overrides, expected: &LXD{Channel: "5.20/stable"}},
	}

	for _, tc := range tests {
		lxd := NewLXD(tc.config)
		if !reflect.DeepEqual(tc.expected, lxd) {
			t.Fatalf("expected: %v, got: %v", tc.expected, lxd)
		}
	}
}
