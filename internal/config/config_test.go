package config

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestFlagToEnvVar(t *testing.T) {
	type test struct {
		flag     string
		expected string
	}

	viper.SetEnvPrefix("CONCIERGE")

	tests := []test{
		{flag: "juju-channel", expected: "CONCIERGE_JUJU_CHANNEL"},
		{flag: "rockcraft-channel", expected: "CONCIERGE_ROCKCRAFT_CHANNEL"},
		{flag: "foobar", expected: "CONCIERGE_FOOBAR"},
	}

	for _, tc := range tests {
		ev := flagToEnvVar(tc.flag)
		if !reflect.DeepEqual(tc.expected, ev) {
			t.Fatalf("expected: %v, got: %v", tc.expected, ev)
		}
	}
}
