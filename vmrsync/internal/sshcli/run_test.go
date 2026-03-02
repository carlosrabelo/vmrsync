package sshcli

import (
	"errors"
	"os/exec"
	"testing"
)

func TestExitCode(t *testing.T) {
	err := exec.Command("sh", "-c", "exit 255").Run()
	wrapped := &CommandError{Output: "denied", Err: err}

	code, ok := ExitCode(wrapped)
	if !ok || code != 255 {
		t.Fatalf("ExitCode() = (%d, %v), want (255, true)", code, ok)
	}
}

func TestRunSSH_invalidHost(t *testing.T) {
	err := RunSSH([]string{
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=1",
		"-p", "59999",
		"127.0.0.1",
		"echo ok",
	})
	if err == nil {
		t.Fatal("expected error for unreachable SSH")
	}
	var cmdErr *CommandError
	if !errors.As(err, &cmdErr) {
		t.Fatalf("expected *CommandError, got %T: %v", err, err)
	}
}
