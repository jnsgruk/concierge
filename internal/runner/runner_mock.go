package runner

import (
	"context"
	"time"

	retry "github.com/sethvargo/go-retry"
)

// NewTestRunner constructs a new command runner.
func NewTestRunner() *TestRunner {
	return &TestRunner{}
}

// TestRunner represents a struct that can run commands.
type TestRunner struct {
	ExecutedCommands []string
	desiredReturn    []byte
	desiredError     error
}

// SetNextReturn sets a static return value representing command combined output,
// and a desired error return for the next command executed by the runner.
func (r *TestRunner) SetNextReturn(b []byte, err error) {
	r.desiredReturn = b
	r.desiredError = err
}

// Run executes the command, returning the stdout/stderr where appropriate.
func (r *TestRunner) Run(c *Command) ([]byte, error) {
	r.ExecutedCommands = append(r.ExecutedCommands, c.commandString())
	returnValue := r.desiredReturn
	returnErr := r.desiredError
	r.desiredReturn = []byte{}
	r.desiredError = nil
	return returnValue, returnErr
}

// RunWithRetries executes the command, retrying utilising an exponential backoff pattern,
// which starts at 1 second. Retries will be attempted up to the specified maximum duration.
func (r *TestRunner) RunWithRetries(c *Command, maxDuration time.Duration) ([]byte, error) {
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
func (r *TestRunner) RunCommands(commands ...*Command) error {
	for _, cmd := range commands {
		_, err := r.Run(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}
