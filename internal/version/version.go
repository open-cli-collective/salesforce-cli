// Package version provides build-time version information.
// Variables are set via ldflags during build (see Makefile).
package version

// Build-time variables set via ldflags
var (
	// Version is the semantic version (from git tag or "dev")
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"
)

// Info returns the version string
func Info() string {
	return Version
}

// Full returns the full version information including commit and build date
func Full() string {
	return Version + " (commit: " + Commit + ", built: " + BuildDate + ")"
}
