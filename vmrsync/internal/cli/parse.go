package cli

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/argv"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/config"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/excludes"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/hostcheck"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/usage"
)

// ParseSyncArgs parses flags and positional args for sync and check commands.
func ParseSyncArgs(command string, args []string) *config.AppConfig {
	cfg := &config.AppConfig{Command: command}
	var excludesFlag config.ExcludeFlags

	fs := flag.NewFlagSet("vmrsync "+command, flag.ExitOnError)
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Print the rsync command that would be executed, without running it")
	fs.Var(&excludesFlag, "exclude", "Exclude files matching pattern")
	fs.StringVar(&cfg.ExcludeFrom, "exclude-from", "", "Read exclude patterns from file")
	fs.StringVar(&cfg.SSHPort, "ssh-port", "", "SSH port")
	fs.StringVar(&cfg.SSHKey, "ssh-key", "", "SSH private key path")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose output")
	fs.BoolVar(&cfg.NoDelete, "no-delete", false, "Do not delete files in destination")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "Suppress progress output (errors still shown)")
	fs.IntVar(&cfg.TimeoutSeconds, "timeout-seconds", 7200, "Hard timeout for rsync runtime in seconds (0 disables)")
	fs.BoolVar(&cfg.Staging, "staging", false, "Use /vmrsync as remote root instead of mirroring local path")
	fs.Usage = usage.Print

	positional, flagTokens := argv.SplitArgs(args)

	if err := fs.Parse(flagTokens); err != nil {
		log.Fatal(err)
	}
	if cfg.TimeoutSeconds < 0 {
		log.Fatalf("--timeout-seconds must be >= 0")
	}

	cfg.Excludes = excludesFlag
	if cfg.ExcludeFrom != "" {
		fromFile, err := excludes.LoadFile(cfg.ExcludeFrom)
		if err != nil {
			log.Fatal(err)
		}
		cfg.Excludes = append(fromFile, cfg.Excludes...)
	}

	if len(positional) > 0 {
		cfg.Machine = positional[0]
	}
	if len(positional) > 1 {
		cfg.Folder = positional[1]
	}
	if len(positional) > 2 {
		log.Fatalf("too many positional arguments: %v", positional)
	}
	if cfg.Machine == "" {
		log.Fatalf("machine not specified")
	}
	if cfg.Folder != "" {
		cfg.Folder = strings.Trim(filepath.Clean(cfg.Folder), "/")
		if strings.Contains(cfg.Folder, "..") {
			log.Fatalf("security error: folder argument contains '..', which is not allowed")
		}
	}

	if err := hostcheck.EnsureRemoteSSHHost(cfg.Machine); err != nil {
		log.Fatal(err)
	}

	return cfg
}
