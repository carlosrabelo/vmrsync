package main

import "strings"

// VERSION is set at build time via -ldflags "-X main.VERSION=<commit>".
// Defaults to "dev" when built without ldflags.
var VERSION = "dev"

const vmrsyncRoot = "/vmrsync"

// AppConfig stores the configuration for a sync operation.
type AppConfig struct {
	Command   string
	Machine   string
	Folder    string
	DryRun    bool
	Excludes  []string
	SSHPort   string
	SSHKey    string
	Verbose   bool
	NoDelete  bool
	// TimeoutSeconds is a hard upper bound for the rsync process runtime.
	// 0 disables the timeout.
	TimeoutSeconds int
	LocalRoot string
	Staging   bool // when true, uses /vmrsync as remote root instead of mirroring local
}

// effectiveRemoteRoot returns the remote root to use:
// --staging → /vmrsync, otherwise mirrors the local root exactly.
func (c *AppConfig) effectiveRemoteRoot() string {
	if c.Staging {
		return vmrsyncRoot
	}
	return c.LocalRoot
}

// excludeFlags is a custom flag type that supports multiple --exclude flags.
type excludeFlags []string

func (i *excludeFlags) String() string {
	return strings.Join(*i, " ")
}

func (i *excludeFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
