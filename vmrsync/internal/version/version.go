// Package version holds the build-injected release string.
package version

// Version is set at build time via -ldflags "-X .../version.Version=<commit>".
// Defaults to "dev" when built without ldflags.
var Version = "dev"
