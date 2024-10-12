package runner

import (
	"bytes"
	"log/slog"
	"os/exec"
	"os/user"
	"strings"
)

// Command represents a given command to be executed by Concierge, along with the
// user and group that should be assumed if required.
type Command struct {
	Executable string
	Args       []string
	User       string
	Group      string
}

// CommandResult carries the output from an executed command.
type CommandResult struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

// NewCommand constructs a command to be run as the current user/group.
func NewCommand(executable string, args []string) *Command {
	return &Command{
		Executable: executable,
		Args:       args,
		User:       "",
		Group:      "",
	}
}

// NewCommandSudo constructs a command to be run as the "root" user.
func NewCommandSudo(executable string, args []string) *Command {
	return &Command{
		Executable: executable,
		Args:       args,
		User:       "root",
		Group:      "",
	}
}

// NewCommandWithGroup constructs a command to be run, assuming membership in the specified group.
func NewCommandWithGroup(executable string, args []string, group string) *Command {
	assumedGroup := group

	// We don't check the error here, and instead carry on as we were, trying to use
	// the specified group.
	currentUser, _ := user.Current()

	// Don't try to switch group if we're already root
	if currentUser.Uid == "0" {
		assumedGroup = ""
	}

	return &Command{
		Executable: executable,
		Args:       args,
		User:       "",
		Group:      assumedGroup,
	}
}

// commandString puts together a command to be executed in a shell, including the `sudo`
// command and its arguments where appropriate.
func (c *Command) commandString() string {
	path, err := exec.LookPath(c.Executable)
	if err != nil {
		slog.Warn("Failed to lookup command in path", "command", c.Executable)
		path = c.Executable
	}

	cmdArgs := []string{}

	if len(c.User) > 0 || len(c.Group) > 0 {
		cmdArgs = append(cmdArgs, "sudo")
	}

	if len(c.User) > 0 {
		cmdArgs = append(cmdArgs, "-u", c.User)
	}

	if len(c.Group) > 0 {
		cmdArgs = append(cmdArgs, "-g", c.Group)
	}

	cmdArgs = append(cmdArgs, path)
	cmdArgs = append(cmdArgs, c.Args...)

	return strings.Join(cmdArgs, " ")
}
