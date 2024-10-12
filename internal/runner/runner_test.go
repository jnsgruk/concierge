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

func TestNewCommandSudo(t *testing.T) {
	expected := &Command{
		Executable: "apt-get",
		Args:       []string{"install", "-y", "cowsay"},
		User:       "root",
		Group:      "",
	}

	command := NewCommandSudo("apt-get", []string{"install", "-y", "cowsay"})
	if !reflect.DeepEqual(expected, command) {
		t.Fatalf("expected: %v, got: %v", expected, command)
	}
}

func TestNewCommandWithGroup(t *testing.T) {
	expected := &Command{
		Executable: "apt-get",
		Args:       []string{"install", "-y", "cowsay"},
		User:       "",
		Group:      "foo",
	}

	command := NewCommandWithGroup("apt-get", []string{"install", "-y", "cowsay"}, "foo")
	if !reflect.DeepEqual(expected, command) {
		t.Fatalf("expected: %v, got: %v", expected, command)
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
			command:  NewCommandSudo("CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}),
			expected: "sudo -u root CONCIERGE_TEST_COMMAND install -y cowsay",
		},
		{
			command:  NewCommandWithGroup("CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}, "apters"),
			expected: "sudo -g apters CONCIERGE_TEST_COMMAND install -y cowsay",
		},
	}

	for _, tc := range tests {
		commandString := tc.command.commandString()
		if !reflect.DeepEqual(tc.expected, commandString) {
			t.Fatalf("expected: %v, got: %v", tc.expected, commandString)
		}
	}
}
