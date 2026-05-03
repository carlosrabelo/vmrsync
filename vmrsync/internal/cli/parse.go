package cli

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/argv"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/config"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/hostcheck"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/usage"
)

// ParseSyncArgs parses flags and positional args for in/out.
func ParseSyncArgs(command string, args []string) *config.AppConfig {
	cfg := &config.AppConfig{Command: command}
	var excludes config.ExcludeFlags

	fs := flag.NewFlagSet("vmrsync "+command, flag.ExitOnError)
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Print the rsync command that would be executed, without running it")
	fs.Var(&excludes, "exclude", "Exclude files matching pattern")
	fs.StringVar(&cfg.SSHPort, "ssh-port", "", "SSH port")
	fs.StringVar(&cfg.SSHKey, "ssh-key", "", "SSH private key path")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose output")
	fs.BoolVar(&cfg.NoDelete, "no-delete", false, "Do not delete files in destination")
	fs.IntVar(&cfg.TimeoutSeconds, "timeout-seconds", 7200, "Hard timeout for rsync runtime in seconds (0 disables)")
	fs.BoolVar(&cfg.Staging, "staging", false, "Use /vmrsync as remote root instead of mirroring local path")
	fs.Usage = usage.Print

	positional, flagTokens := argv.SplitArgs(args)

	if err := fs.Parse(flagTokens); err != nil {
		log.Fatal(err)
	}

	cfg.Excludes = excludes

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
