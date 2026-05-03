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
