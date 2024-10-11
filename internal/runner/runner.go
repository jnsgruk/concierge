package runner

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

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
func (r *Runner) Run(c *Command) (*CommandResult, error) {
	path, err := exec.LookPath(c.Executable)
	if err != nil {
		return nil, fmt.Errorf("could not find '%s' command in path: %w", c.Executable, err)
	}

	logger := slog.Default()
	if len(c.User) > 0 {
		logger = slog.With("user", c.User)
	}
	if len(c.Group) > 0 {
		logger = slog.With("group", c.Group)
	}

	cmd := exec.Command(os.Getenv("SHELL"), "-c", c.commandString())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Debug("running command", "command", fmt.Sprintf("%s %s", path, strings.Join(c.Args, " ")))
	err = cmd.Run()

	return &CommandResult{Stdout: stdout, Stderr: stderr}, err
}
	}

	cmdArgs = append(cmdArgs, c.Executable)
	cmdArgs = append(cmdArgs, c.Args...)

	return strings.Join(cmdArgs, " ")
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
