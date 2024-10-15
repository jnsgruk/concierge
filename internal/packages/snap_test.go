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
		{input: "juju", expected: &Snap{Name: "juju"}},
		{input: "juju/latest/edge", expected: &Snap{Name: "juju", Channel: "latest/edge"}},
		{input: "juju/stable", expected: &Snap{Name: "juju", Channel: "stable"}},
	}

	for _, tc := range tests {
		snap := NewSnapFromString(tc.input)
		if tc.expected.Channel != snap.Channel {
			t.Fatalf("incorrect snap channel; expected: %v, got: %v", tc.expected, snap)
		}
		if tc.expected.Name != snap.Name {
			t.Fatalf("incorrect snap name; expected: %v, got: %v", tc.expected, snap)
		}
	}
}
