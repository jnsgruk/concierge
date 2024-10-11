package packages

import (
	"reflect"
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
		if !reflect.DeepEqual(tc.expected, snap) {
			t.Fatalf("expected: %v, got: %v", tc.expected, snap)
		}
	}
}
