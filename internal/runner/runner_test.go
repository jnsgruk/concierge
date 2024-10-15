package runner

import (
	"reflect"
	"testing"
)

func TestNewCommand(t *testing.T) {
	expected := &Command{
		Executable: "juju",
		Args:       []string{"add-model", "testing"},
		User:       "",
		Group:      "",
	}

	command := NewCommand("juju", []string{"add-model", "testing"})
	if !reflect.DeepEqual(expected, command) {
		t.Fatalf("expected: %v, got: %v", expected, command)
	}
}

func TestNewCommandAs(t *testing.T) {
	expected := &Command{
		Executable: "apt-get",
		Args:       []string{"install", "-y", "cowsay"},
		User:       "test-user",
		Group:      "foo",
	}

	command := NewCommandAs("test-user", "foo", "apt-get", []string{"install", "-y", "cowsay"})
	if !reflect.DeepEqual(expected, command) {
		t.Fatalf("expected: %+v, got: %+v", expected, command)
	}
}

func TestNewCommandAsRoot(t *testing.T) {
	expected := &Command{
		Executable: "apt-get",
		Args:       []string{"install", "-y", "cowsay"},
		User:       "",
		Group:      "",
	}

	command := NewCommandAs("root", "foo", "apt-get", []string{"install", "-y", "cowsay"})
	if !reflect.DeepEqual(expected, command) {
		t.Fatalf("expected: %+v, got: %+v", expected, command)
	}
}

func TestCommandString(t *testing.T) {
	type test struct {
		command  *Command
		expected string
	}

	// Use CONCIERGE_TEST_COMMAND to avoid $PATH lookups making tests flaky
	tests := []test{
		{
			command:  NewCommand("CONCIERGE_TEST_COMMAND", []string{"add-model", "testing"}),
			expected: "CONCIERGE_TEST_COMMAND add-model testing",
		},
		{
			command:  NewCommandAs("test-user", "", "CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}),
			expected: "sudo -u test-user CONCIERGE_TEST_COMMAND install -y cowsay",
		},
		{
			command:  NewCommandAs("test-user", "apters", "CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}),
			expected: "sudo -u test-user -g apters CONCIERGE_TEST_COMMAND install -y cowsay",
		},
	}

	for _, tc := range tests {
		commandString := tc.command.commandString()
		if !reflect.DeepEqual(tc.expected, commandString) {
			t.Fatalf("expected: %v, got: %v", tc.expected, commandString)
		}
	}
}
