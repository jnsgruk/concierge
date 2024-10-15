package runner

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"os/user"

	"github.com/canonical/x-go/strutil/shlex"
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

// NewCommandAsRealUser constructs a command to be run as the real user/group, which is
// different to the current user when concierge is executed with sudo.
func NewCommandAsRealUser(executable string, args []string) *Command {
	return &Command{
		Executable: executable,
		Args:       args,
		User:       os.Getenv("SUDO_USER"),
		Group:      "",
	}
}

// NewCommandAsRealUser constructs a command to be run as the real user/group, which is
// different to the current user when concierge is executed with sudo.
func NewCommandAsRealUserWithGroup(executable string, args []string, group string) *Command {
	realUser, err := RealUser()
	if err != nil {
		slog.Warn("failed to lookup user, defaulting to 'root'", "error", err.Error())
		realUser = &user.User{Username: "root"}
		group = ""
	}

	// Don't try to drop privileges with a group if the real user is actually root
	if realUser.Uid == "0" {
		group = ""
	}

	return &Command{
		Executable: executable,
		Args:       args,
		User:       realUser.Username,
		Group:      group,
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

	return shlex.Join(cmdArgs)
}
