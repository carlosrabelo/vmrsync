package usage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/config"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/version"
)

// Print writes CLI help to stdout.
func Print() {
	cmd := filepath.Base(os.Args[0])
	fmt.Printf(`VM RSync v%s - Synchronize files between local and remote machines via rsync

Usage:
  %s [options] <command> <machine> [<folder>]

Commands:
  restore     Sync FROM remote machine TO local machine
  backup      Sync FROM local machine TO remote machine
  check       Verify SSH connectivity and that sync paths exist
  prepare     Create and configure %s on remote machine (requires sudo)
  completion  Install system-wide shell completion (root only): bash|zsh|fish
  version     Show version information
  help        Show help (-h, --help)

Options:
  --dry-run              Print the rsync command that would be executed, without running it
  --exclude <pattern>    Exclude files matching pattern (repeatable)
  --exclude-from <file>  Read exclude patterns from file
  --ssh-port <port>      SSH port
  --ssh-key <path>       SSH private key path
  --verbose              Enable verbose output
  --quiet                Suppress progress output (errors still shown)
  --no-delete            Do not delete files at destination
  --timeout-seconds <n>  Hard timeout for rsync runtime in seconds (0 disables; default: 7200)
  --staging              Use /vmrsync/<folder> as remote root (e.g. /vmrsync/sources)

Environment Variables:
  VMRSYNC_PATH   Sync root directory, local and remote (default: $HOME/Sources)

Exclude files:
  Patterns from .vmrsyncignore in the synced folder are applied automatically.

Examples:
  %s prepare vm21
  %s check vm21 project1
  %s restore vm21 project1
  %s restore vm21 project1 --dry-run
  %s backup vm21 project1 --exclude "*.log" --exclude "node_modules"
  %s restore vm21 --ssh-port 2222 --ssh-key ~/.ssh/id_rsa
`, version.Version, cmd, config.VmrsyncRoot, cmd, cmd, cmd, cmd, cmd, cmd)
}
