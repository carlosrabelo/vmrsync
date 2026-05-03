package sshcli

import (
	"strings"
)

// BuildSSHFlags returns SSH flag args for -p and -i options.
func BuildSSHFlags(port, key string) []string {
	var args []string
	if port != "" {
		args = append(args, "-p", port)
	}
	if key != "" {
		args = append(args, "-i", key)
	}
	return args
}

// ShellQuoteSingle wraps s for POSIX shell single-quote context.
func ShellQuoteSingle(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// TestDirCommand returns a remote shell fragment to verify directory existence.
// Paths are single-quoted; no "test --" because BusyBox/ash test rejects GNU-style "--".
func TestDirCommand(path string) string {
	return "test -d " + ShellQuoteSingle(path)
}

// MkdirPCommand returns a remote shell fragment: mkdir -p <path> (quoted).
func MkdirPCommand(path string) string {
	return "mkdir -p " + ShellQuoteSingle(path)
}
