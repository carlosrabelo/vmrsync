package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"
)

// valueFlags lists flags that consume the next argument as their value.
var valueFlags = map[string]bool{
	"exclude": true, "ssh-port": true, "ssh-key": true, "timeout-seconds": true,
}

// splitArgs separates positional arguments from flags, supporting any order.
// Flags that take a value (--ssh-port, --ssh-key, --exclude) consume the next token.
func splitArgs(args []string) (positional []string, flags []string) {
	i := 0
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			i++
			continue
		}
		// Extract flag name (strip leading dashes, ignore =value suffix).
		name := strings.TrimLeft(arg, "-")
		if eq := strings.Index(name, "="); eq >= 0 {
			name = name[:eq]
		}
		flags = append(flags, arg)
		// If this flag takes a value and the next token is not a flag, consume it.
		if valueFlags[name] && !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			i++
			flags = append(flags, args[i])
		}
		i++
	}
	return
}

func parseArgs(command string, args []string) *AppConfig {
	config := &AppConfig{Command: command}
	var excludes excludeFlags

	fs := flag.NewFlagSet("vmrsync "+command, flag.ExitOnError)
	fs.BoolVar(&config.DryRun, "dry-run", false, "Print the rsync command that would be executed, without running it")
	fs.Var(&excludes, "exclude", "Exclude files matching pattern")
	fs.StringVar(&config.SSHPort, "ssh-port", "", "SSH port")
	fs.StringVar(&config.SSHKey, "ssh-key", "", "SSH private key path")
	fs.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	fs.BoolVar(&config.NoDelete, "no-delete", false, "Do not delete files in destination")
	fs.IntVar(&config.TimeoutSeconds, "timeout-seconds", 7200, "Hard timeout for rsync runtime in seconds (0 disables)")
	fs.BoolVar(&config.Staging, "staging", false, "Use /vmrsync as remote root instead of mirroring local path")
	fs.Usage = showUsage

	positional, flags := splitArgs(args)

	if err := fs.Parse(flags); err != nil {
		log.Fatal(err)
	}

	config.Excludes = excludes

	if len(positional) > 0 {
		config.Machine = positional[0]
	}
	if len(positional) > 1 {
		config.Folder = positional[1]
	}
	if len(positional) > 2 {
		log.Fatalf("too many positional arguments: %v", positional)
	}
	if config.Machine == "" {
		log.Fatalf("machine not specified")
	}
	if config.Folder != "" {
		config.Folder = strings.Trim(filepath.Clean(config.Folder), "/")
		if strings.Contains(config.Folder, "..") {
			log.Fatalf("security error: folder argument contains '..', which is not allowed")
		}
	}

	if err := ensureRemoteSSHHost(config.Machine); err != nil {
		log.Fatal(err)
	}

	return config
}
