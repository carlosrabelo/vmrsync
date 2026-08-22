package sshcli

import (
	"strings"
	"testing"
)

func TestTestDirCommand_noDoubleDash(t *testing.T) {
	// BusyBox/ash `test` treats GNU's "--" as invalid; use `test -d 'path'` only.
	got := TestDirCommand("/home/carlos/Sources")
	want := "test -d '/home/carlos/Sources'"
	if got != want {
		t.Errorf("TestDirCommand() = %q, want %q", got, want)
	}
	if strings.Contains(got, "test -d --") {
		t.Errorf("must not use GNU-only `test -d --`: %q", got)
	}
}

func TestFormatRsyncSSHShell_quotesKeyWithSpaces(t *testing.T) {
	flags := BuildSSHFlags("2222", "/tmp/my key")
	got := FormatRsyncSSHShell(flags)
	if !strings.Contains(got, "-i '/tmp/my key'") {
		t.Fatalf("expected quoted key path, got %q", got)
	}
	if !strings.HasPrefix(got, "ssh ") {
		t.Fatalf("expected ssh prefix, got %q", got)
	}
}

func TestShellQuoteSingle_embedsSingleQuote(t *testing.T) {
	got := ShellQuoteSingle("a'b")
	want := `'a'"'"'b'`
	if got != want {
		t.Fatalf("ShellQuoteSingle() = %q, want %q", got, want)
	}
}
