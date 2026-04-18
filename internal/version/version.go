// Package version holds the build-time version string.
// It is set at build time via:
//
//	go build -ldflags "-X github.com/aminmesbahi/skell/internal/version.Version=v1.2.3"
package version

// Version is the current skell CLI version. Defaults to "dev" for local builds.
var Version = "dev"
