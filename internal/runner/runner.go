package runner

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
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

// Run executes the command, returning the stdout/stderr where appropriate.
func (c *Command) Run() (*CommandResult, error) {
	path, err := exec.LookPath(c.Executable)
	if err != nil {
		return nil, fmt.Errorf("could not find '%s' command in path: %w", c.Executable, err)
	}

	cmdArgs := []string{}
	logger := slog.Default()

	if len(c.User) > 0 || len(c.Group) > 0 {
		cmdArgs = append(cmdArgs, "sudo")
	}

	if len(c.User) > 0 {
		cmdArgs = append(cmdArgs, "-u", c.User)
		logger = slog.With("user", c.User)
	}

	if len(c.Group) > 0 {
		cmdArgs = append(cmdArgs, "-g", c.Group)
		logger = slog.With("group", c.Group)
	}

	cmdArgs = append(cmdArgs, path)
	cmdArgs = append(cmdArgs, c.Args...)

	cmd := exec.Command(os.Getenv("SHELL"), "-c", strings.Join(cmdArgs, " "))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Debug("running command", "command", fmt.Sprintf("%s %s", path, strings.Join(c.Args, " ")))
	err = cmd.Run()

	return &CommandResult{Stdout: stdout, Stderr: stderr}, err
}

// RunCommands takes a variadic number of Command's, and runs them in a loop, returning
// and error if any command fails.
func RunCommands(commands ...*Command) error {
	for _, cmd := range commands {
		_, err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
