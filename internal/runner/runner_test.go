package runner

import (
	"fmt"
	"os"
	"os/user"
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
	// Fake a sudo user
	user, _ := user.Current()
	os.Setenv("SUDO_USER", user.Username)

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
	}

	// Fake a SUDO_USER
	user, _ := user.Current()
	os.Setenv("SUDO_USER", user.Username)
	defer os.Setenv("SUDO_USER", "")

	// Use CONCIERGE_TEST_COMMAND to avoid $PATH lookups making tests flaky
	tests := []test{
		{
			command:  NewCommand("CONCIERGE_TEST_COMMAND", []string{"add-model", "testing"}),
			expected: "CONCIERGE_TEST_COMMAND add-model testing",
		},
		{
			command:  NewCommandAsRealUser("CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}),
			expected: fmt.Sprintf("sudo -u %s CONCIERGE_TEST_COMMAND install -y cowsay", os.Getenv("SUDO_USER")),
		},
		{
			command:  NewCommandAsRealUserWithGroup("CONCIERGE_TEST_COMMAND", []string{"install", "-y", "cowsay"}, "apters"),
			expected: fmt.Sprintf("sudo -u %s -g apters CONCIERGE_TEST_COMMAND install -y cowsay", os.Getenv("SUDO_USER")),
		},
	}

	for _, tc := range tests {
		commandString := tc.command.commandString()
		if !reflect.DeepEqual(tc.expected, commandString) {
			t.Fatalf("expected: %v, got: %v", tc.expected, commandString)
		}
	}
}
