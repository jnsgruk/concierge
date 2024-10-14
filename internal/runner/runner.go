package runner

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

// NewRunner constructs a new command runner.
func NewRunner(trace bool) *Runner {
	return &Runner{trace: trace}
}

// Runner represents a struct that can run commands.
type Runner struct {
	trace bool
}

// Run executes the command, returning the stdout/stderr where appropriate.
func (r *Runner) Run(c *Command) ([]byte, error) {
	logger := slog.Default()
	if len(c.User) > 0 {
		logger = slog.With("user", c.User)
	}
	if len(c.Group) > 0 {
		logger = slog.With("group", c.Group)
	}

	shell, err := getShellPath()
	if err != nil {
		return nil, fmt.Errorf("unable to determine shell path to run command")
	}

	cmd := exec.Command(shell, "-c", c.commandString())

	logger.Debug("Running command", "command", c.commandString())

	output, err := cmd.CombinedOutput()

	if r.trace {
		fmt.Print(generateTraceMessage(c.commandString(), output))
	}

	return output, err
}

// RunCommands takes a variadic number of Command's, and runs them in a loop, returning
// and error if any command fails.
func (r *Runner) RunCommands(commands ...*Command) error {
	for _, cmd := range commands {
		_, err := r.Run(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// generateTraceMessage creates a formatted string that is written to stdout, representing
// a command and it's output when concierge is run with `--trace`.
func generateTraceMessage(cmd string, output []byte) string {
	green := color.New(color.FgGreen, color.Bold, color.Underline)
	bold := color.New(color.Bold)

	result := fmt.Sprintf("%s %s\n", green.Sprintf("Command:"), bold.Sprintf(cmd))
	if len(output) > 0 {
		result = fmt.Sprintf("%s%s\n%s", result, green.Sprintf("Output:"), string(output))
	}
	return result
}

// getShellPath tries to find the path to the user's preferred shell, as per the `SHELL“
// environment variable. If that cannot be found, it looks for a path to "bash", and to
// "sh" in that order. If no shell can be found, then an error is returned.
func getShellPath() (string, error) {
	// If the `SHELL` var is set, return that.
	shellVar := os.Getenv("SHELL")
	if len(shellVar) > 0 {
		return shellVar, nil
	}

	// Try both the command name (to lookup in PATH), and common default paths.
	for _, shell := range []string{"bash", "/bin/bash", "sh", "/bin/sh"} {
		// Check if the shell path exists
		if _, err := os.Stat(shell); errors.Is(err, os.ErrNotExist) {
			// If the path doesn't exist, the lookup the value in the `PATH` variable
			path, err := exec.LookPath(shell)
			if err != nil {
				continue
			}
			return path, nil
		}
		return shell, nil
	}

	return "", fmt.Errorf("could not find path to a shell")
}
