package completion

import (
	"os"
	"strings"
	"testing"
)

func TestInstall_requiresShell(t *testing.T) {
	err := Install("")
	if err == nil {
		t.Fatal("expected error for empty shell")
	}
	if !strings.Contains(err.Error(), "bash|zsh|fish") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstall_unsupportedShell(t *testing.T) {
	err := Install("tcsh")
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstall_requiresRoot(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root")
	}
	for _, shell := range []string{"bash", "zsh", "fish"} {
		err := Install(shell)
		if err == nil {
			t.Fatalf("expected root error for %s", shell)
		}
		if !strings.Contains(err.Error(), "requires root") {
			t.Fatalf("unexpected error for %s: %v", shell, err)
		}
		if !strings.Contains(err.Error(), "sudo vmrsync completion "+shell) {
			t.Fatalf("error should suggest sudo for %s: %v", shell, err)
		}
	}
}
