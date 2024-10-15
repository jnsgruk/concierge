package runnertest

import (
	"os"
	"os/user"
	"time"

	"github.com/jnsgruk/concierge/internal/runner"
)

// NewMockRunner constructs a new mock command runner.
func NewMockRunner() *MockRunner {
	return &MockRunner{
		CreatedFiles: map[string]string{},
	}
}

// MockRunner represents a struct that can emulate running commands.
type MockRunner struct {
	ExecutedCommands   []string
	CreatedFiles       map[string]string
	CreatedDirectories []string
	Deleted            []string
	desiredReturn      []byte
	desiredError       error
}

// SetNextReturn sets a static return value representing command combined output,
// and a desired error return for the next command executed by the runner.
func (r *MockRunner) SetNextReturn(b []byte, err error) {
	r.desiredReturn = b
	r.desiredError = err
}

// User returns the user the runner executes commands on behalf of.
func (r *MockRunner) User() *user.User {
	return &user.User{
		Username: "test-user",
		Uid:      "666",
		Gid:      "666",
		HomeDir:  os.TempDir(),
	}
}

// Run executes the command, returning the stdout/stderr where appropriate.
func (r *MockRunner) Run(c *runner.Command) ([]byte, error) {
	r.ExecutedCommands = append(r.ExecutedCommands, c.CommandString())
	returnValue := r.desiredReturn
	returnErr := r.desiredError
	r.desiredReturn = []byte{}
	r.desiredError = nil
	return returnValue, returnErr
}

// RunWithRetries executes the command, retrying utilising an exponential backoff pattern,
// which starts at 1 second. Retries will be attempted up to the specified maximum duration.
func (r *MockRunner) RunWithRetries(c *runner.Command, maxDuration time.Duration) ([]byte, error) {
	return r.Run(c)
}

// RunMany takes a variadic number of Command's, and runs them in a loop, returning
// and error if any command fails.
func (r *MockRunner) RunMany(commands ...*runner.Command) error {
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
func (r *MockRunner) WriteHomeDirFile(filepath string, contents []byte) error {
	r.CreatedFiles[filepath] = string(contents)
	return nil
}

// MkHomeSubdirectory takes a relative folder path and creates it recursively in the real
// user's home directory.
func (r *MockRunner) MkHomeSubdirectory(subdirectory string) error {
	r.CreatedDirectories = append(r.CreatedDirectories, subdirectory)
	return nil
}

// ReadHomeDirFile takes a path relative to the real user's home dir, and reads the content
// from the file
func (r *MockRunner) ReadHomeDirFile(filepath string) ([]byte, error) {
	return nil, nil
}

// RemoveAllHome recursively removes a file path from the user's home directory.
func (r *MockRunner) RemoveAllHome(filePath string) error {
	r.Deleted = append(r.Deleted, filePath)
	return nil
}