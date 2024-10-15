package runner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	retry "github.com/sethvargo/go-retry"
)

// NewRunner constructs a new command runner.
func NewRunner(trace bool) (*Runner, error) {
	realUser, err := realUser()
	if err != nil {
		return nil, fmt.Errorf("failed to lookup effective user details: %w", err)
	}
	return &Runner{trace: trace, user: realUser}, nil
}

// Runner represents a struct that can run commands.
type Runner struct {
	trace bool
	user  *user.User
}

// User returns a user struct containing details of the "real" user, which
// may differ from the current user when concierge is executed with `sudo`.
func (r *Runner) User() *user.User { return r.user }

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

// RunWithRetries executes the command, retrying utilising an exponential backoff pattern,
// which starts at 1 second. Retries will be attempted up to the specified maximum duration.
func (r *Runner) RunWithRetries(c *Command, maxDuration time.Duration) ([]byte, error) {
	backoff := retry.NewExponential(1 * time.Second)
	backoff = retry.WithMaxDuration(maxDuration, backoff)
	ctx := context.Background()

	return retry.DoValue(ctx, backoff, func(ctx context.Context) ([]byte, error) {
		output, err := r.Run(c)
		if err != nil {
			return nil, retry.RetryableError(err)
		}

		return output, nil
	})
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

// WriteHomeDirFile takes a path relative to the real user's home dir, and writes the contents
// specified to it.
func (r *Runner) WriteHomeDirFile(filePath string, contents []byte) error {
	dir := path.Dir(filePath)

	err := r.MkHomeSubdirectory(dir)
	if err != nil {
		return err
	}

	filePath = path.Join(path.Join(r.user.HomeDir, filePath))

	if err := os.WriteFile(filePath, contents, 0644); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", filePath, err)
	}

	err = r.chownRecursively(filePath, r.user)
	if err != nil {
		return fmt.Errorf("failed to change ownership of file '%s': %w", filePath, err)
	}

	return nil
}

// MkHomeSubdirectory takes a relative folder path and creates it recursively in the real
// user's home directory.
func (r *Runner) MkHomeSubdirectory(subdirectory string) error {
	if path.IsAbs(subdirectory) {
		return fmt.Errorf("only relative paths supported")
	}

	dir := path.Join(r.user.HomeDir, subdirectory)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	parts := strings.Split(subdirectory, "/")
	if len(parts) > 0 {
		dir = path.Join(r.user.HomeDir, parts[0])
	}

	err = r.chownRecursively(dir, r.user)
	if err != nil {
		return fmt.Errorf("failed to change ownership of directory '%s': %w", dir, err)
	}

	return nil
}

// ReadHomeDirFile takes a path relative to the real user's home dir, and reads the content
// from the file
func (r *Runner) ReadHomeDirFile(filePath string) ([]byte, error) {
	homePath := path.Join(r.user.HomeDir, filePath)

	if _, err := os.Stat(homePath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("file '%s' does not exist: %w", homePath, err)
	}

	return os.ReadFile(homePath)
}

// RemoveAllHome recursively removes a file path from the user's home directory.
func (r *Runner) RemoveAllHome(filePath string) error {
	return os.RemoveAll(path.Join(r.user.HomeDir, filePath))
}

// ChownRecursively recursively changes ownership of a given filepath to the uid/gid of
// the specified user.
func (r *Runner) chownRecursively(path string, user *user.User) error {
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return fmt.Errorf("failed to convert user id string to int: %w", err)
	}
	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return fmt.Errorf("failed to convert group id string to int: %w", err)
	}

	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		err = os.Chown(path, uid, gid)
		if err != nil {
			return err
		}

		return nil
	})

	slog.Debug("Filesystem ownership changed", "user", user.Username, "group", user.Gid, "path", path)
	return err
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

// realUser returns a user struct containing details of the "real" user, which
// may differ from the current user when concierge is executed with `sudo`.
func realUser() (*user.User, error) {
	realUser := os.Getenv("SUDO_USER")
	if len(realUser) == 0 {
		return user.Lookup("root")
	}

	return user.Lookup(realUser)
}
