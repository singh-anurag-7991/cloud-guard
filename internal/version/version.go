// Package version exposes build metadata injected at compile time via -ldflags.
//
// This exists so "is my deploy live yet?" is answerable from a browser instead of
// by SSH-ing into the box and reading journalctl. Hit /healthz and compare Commit
// against `git rev-parse --short HEAD`.
package version

var (
	// Commit is the short git SHA the binary was built from.
	Commit = "dev"
	// BuildTime is the UTC timestamp of the build.
	BuildTime = "unknown"
)
