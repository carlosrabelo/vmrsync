package sshcli

import (
	"strings"
)

// BuildSSHFlags returns non-interactive SSH flag args for -p and -i options.
func BuildSSHFlags(port, key string) []string {
	args := []string{
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
	}
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

func shellToken(s string) string {
	if s == "" {
		return "''"
	}
	if strings.ContainsAny(s, " \t\n\\\"'$`|&;<>(){}[]!*?#~") {
		return ShellQuoteSingle(s)
	}
	return s
}

// FormatRsyncSSHShell builds an rsync -e shell command: ssh [quoted flags...].
func FormatRsyncSSHShell(flags []string) string {
	parts := make([]string, 0, len(flags)+1)
	parts = append(parts, "ssh")
	for _, f := range flags {
		parts = append(parts, shellToken(f))
	}
	return strings.Join(parts, " ")
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
