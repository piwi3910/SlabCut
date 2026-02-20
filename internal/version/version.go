// Package version holds build-time version information injected via ldflags.
package version

import "fmt"

// Version is the semantic version or git tag, set via -ldflags.
var Version = "dev"

// Commit is the short git commit hash, set via -ldflags.
var Commit = "unknown"

// Short returns a human-readable version string, e.g. "v1.0.0 (abc1234)".
func Short() string {
	if Version == "dev" && Commit == "unknown" {
		return "dev"
	}
	if Commit == "unknown" {
		return Version
	}
	return fmt.Sprintf("%s (%s)", Version, Commit)
}
