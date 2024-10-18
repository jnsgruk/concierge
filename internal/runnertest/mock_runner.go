package runnertest

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/jnsgruk/concierge/internal/runner"
)

// NewMockRunner constructs a new mock command runner.
func NewMockRunner() *MockRunner {
	return &MockRunner{
		CreatedFiles: map[string]string{},
		mockReturns:  map[string]MockCommandReturn{},
		mockFiles:    map[string][]byte{},
	}
}

// MockCommandReturn contains mocked Output and Error from a given command.
type MockCommandReturn struct {
	Output []byte
	Error  error
}

// MockRunner represents a struct that can emulate running commands.
type MockRunner struct {
	ExecutedCommands   []string
	CreatedFiles       map[string]string
	CreatedDirectories []string
	Deleted            []string

	mockReturns map[string]MockCommandReturn
	mockFiles   map[string][]byte
}

// MockCommandReturn sets a static return value representing command combined output,
// and a desired error return for the specified command.
func (r *MockRunner) MockCommandReturn(command string, b []byte, err error) {
	r.mockReturns[command] = MockCommandReturn{Output: b, Error: err}
}

// MockFile sets a faked expected file contents for a given file.
func (r *MockRunner) MockFile(filePath string, contents []byte) {
	r.mockFiles[filePath] = contents
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
	cmd := c.CommandString()

	r.ExecutedCommands = append(r.ExecutedCommands, cmd)

	val, ok := r.mockReturns[cmd]
	if ok {
		return val.Output, val.Error
	}
	return []byte{}, nil
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

// RunExclusive is a wrapper around Run that uses a mutex to ensure that only one of that
// particular command can be run at a time.
func (r *MockRunner) RunExclusive(c *runner.Command) ([]byte, error) {
	return r.Run(c)
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
func (r *MockRunner) ReadHomeDirFile(filePath string) ([]byte, error) {
	val, ok := r.mockFiles[filePath]
	if !ok {
		return nil, fmt.Errorf("file not found")
	}
	return val, nil
}

// ReadFile takes a path and reads the content from the specified file.
func (r *MockRunner) ReadFile(filePath string) ([]byte, error) {
	val, ok := r.mockFiles[filePath]
	if !ok {
		return nil, fmt.Errorf("file not found")
	}
	return val, nil
}

// RemoveAllHome recursively removes a file path from the user's home directory.
func (r *MockRunner) RemoveAllHome(filePath string) error {
	r.Deleted = append(r.Deleted, filePath)
	return nil
}
