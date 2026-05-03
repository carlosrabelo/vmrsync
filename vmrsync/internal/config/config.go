package config

import "strings"

// VmrsyncRoot is the fixed remote staging directory used with --staging.
const VmrsyncRoot = "/vmrsync"

// AppConfig stores the configuration for a sync operation.
type AppConfig struct {
	Command  string
	Machine  string
	Folder   string
	DryRun   bool
	Excludes []string
	SSHPort  string
	SSHKey   string
	Verbose  bool
	NoDelete bool
	// TimeoutSeconds is a hard upper bound for the rsync process runtime.
	// 0 disables the timeout.
	TimeoutSeconds int
	LocalRoot      string
	Staging        bool // when true, uses /vmrsync as remote root instead of mirroring local
}

// EffectiveRemoteRoot returns the remote root to use:
// --staging → /vmrsync, otherwise mirrors the local root exactly.
func (c *AppConfig) EffectiveRemoteRoot() string {
	if c.Staging {
		return VmrsyncRoot
	}
	return c.LocalRoot
}

// ExcludeFlags is a flag.Value that supports multiple --exclude flags.
type ExcludeFlags []string

func (i *ExcludeFlags) String() string {
	return strings.Join(*i, " ")
}

func (i *ExcludeFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
