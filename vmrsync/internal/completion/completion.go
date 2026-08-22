package completion

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed vmrsync.bash-completion
var bashCompletionScript string

//go:embed vmrsync.zsh-completion
var zshCompletionScript string

//go:embed vmrsync.fish-completion
var fishCompletionScript string

// Install writes shell completion system-wide under /usr/local.
// Requires root. shell must be bash, zsh, or fish.
func Install(shell string) error {
	var dest string
	var script string
	switch shell {
	case "bash":
		dest = "/usr/local/share/bash-completion/completions/vmrsync"
		script = bashCompletionScript
	case "zsh":
		dest = "/usr/local/share/zsh/site-functions/_vmrsync"
		script = zshCompletionScript
	case "fish":
		dest = "/usr/local/share/fish/vendor_completions.d/vmrsync.fish"
		script = fishCompletionScript
	case "":
		return fmt.Errorf("completion requires a shell; run:\n  sudo vmrsync completion bash|zsh|fish")
	default:
		return fmt.Errorf("unsupported shell %q; choose bash, zsh, or fish", shell)
	}

	if os.Geteuid() != 0 {
		return fmt.Errorf("completion requires root; run:\n  sudo vmrsync completion %s", shell)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	if err := os.WriteFile(dest, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write completion file to %s: %w", dest, err)
	}

	fmt.Printf("Installed: %s\n", dest)
	return nil
}
