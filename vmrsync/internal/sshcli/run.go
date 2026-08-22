package sshcli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const checkTimeout = 30 * time.Second

// CommandError holds combined output from a failed ssh invocation.
type CommandError struct {
	Output string
	Err    error
}

func (e *CommandError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("%v: %s", e.Err, e.Output)
	}
	return e.Err.Error()
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// ExitCode returns the process exit code when err wraps *exec.ExitError.
func ExitCode(err error) (int, bool) {
	var cmdErr *CommandError
	if errors.As(err, &cmdErr) {
		err = cmdErr.Err
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), true
	}
	return 0, false
}

// RunSSH runs ssh with BatchMode, a connect timeout, and an overall deadline.
func RunSSH(args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("SSH timed out after %s: %w", checkTimeout, err)
		}
		return &CommandError{Output: string(out), Err: err}
	}
	return nil
}
