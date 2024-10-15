package runner

import (
	"fmt"
	"os"
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

func TestNewCommandAsRealUserWithGroup(t *testing.T) {
	if os.Getenv("SUDO_USER") == "" {
		t.Skip("skipping test because `sudo` is not in use for this invocation")
	}

	expected := &Command{
		Executable: "apt-get",
		Args:       []string{"install", "-y", "cowsay"},
		User:       os.Getenv("SUDO_USER"),
		Group:      "foo",
	}

	command := NewCommandAsRealUserWithGroup("apt-get", []string{"install", "-y", "cowsay"}, "foo")
	if !reflect.DeepEqual(expected, command) {
		t.Fatalf("expected: %+v, got: %+v", expected, command)
	}
}

func TestCommandString(t *testing.T) {
	type test struct {
		command  *Command
		expected string
		skip     bool
	}

	// Use CONCIERGE_TEST_COMMAND to avoid $PATH lookups making tests flaky
	tests := []test{
		{
			command:  NewCommand("CONCIERGE_TEST_COMMAND", []string{"add-model", "testing"}),
			expected: "CONCIERGE_TEST_COMMAND add-model testing",
			skip:     false,
		},
		{
			command:  NewCommandAsRealUser("CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}),
			expected: fmt.Sprintf("sudo -u %s CONCIERGE_TEST_COMMAND install -y cowsay", os.Getenv("SUDO_USER")),
			skip:     os.Getenv("SUDO_USER") == "",
		},
		{
			command:  NewCommandAsRealUserWithGroup("CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}, "apters"),
			expected: fmt.Sprintf("sudo -u %s -g apters CONCIERGE_TEST_COMMAND install -y cowsay", os.Getenv("SUDO_USER")),
			skip:     os.Getenv("SUDO_USER") == "",
		},
	}

	for _, tc := range tests {
		if tc.skip {
			t.Skip("skipping test because `sudo` is not in use for this invocation")
		}

		commandString := tc.command.commandString()
		if !reflect.DeepEqual(tc.expected, commandString) {
			t.Fatalf("expected: %v, got: %v", tc.expected, commandString)
		}
	}
}
