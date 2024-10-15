package packages

import (
	"testing"
)

func TestNewSnapFromString(t *testing.T) {
	type test struct {
		input    string
		expected *Snap
	}

	tests := []test{
		{input: "juju", expected: &Snap{name: "juju"}},
		{input: "juju/latest/edge", expected: &Snap{name: "juju", channel: "latest/edge"}},
		{input: "juju/stable", expected: &Snap{name: "juju", channel: "stable"}},
	}

	for _, tc := range tests {
		snap := NewSnapFromString(tc.input)
		if tc.expected.channel != snap.channel {
			t.Fatalf("incorrect snap channel; expected: %v, got: %v", tc.expected, snap)
		}
		if tc.expected.name != snap.name {
			t.Fatalf("incorrect snap name; expected: %v, got: %v", tc.expected, snap)
		}
	}
}
