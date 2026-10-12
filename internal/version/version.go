// Package version holds the build-time version string.
// It is set at build time via:
//
//	go build -ldflags "-X github.com/aminmesbahi/skell/internal/version.Version=v1.2.3"
//
// Binaries built without those flags (for example 'go install
// github.com/aminmesbahi/skell@v1.2.3') fall back to the module version and
// VCS information the Go toolchain embeds.
package version

import "runtime/debug"

// Version is the current skell CLI version. Defaults to "dev" for local builds.
var Version = "dev"

// Commit is the git commit SHA injected at build time.
var Commit = "none"

// Date is the build date injected at build time.
var Date = "unknown"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		fillFromBuildInfo(info)
	}
}

// fillFromBuildInfo replaces the defaults with what the toolchain recorded.
// Values injected with -ldflags are never overwritten.
func fillFromBuildInfo(info *debug.BuildInfo) {
	if Version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if Commit == "none" && s.Value != "" {
				Commit = s.Value
				if len(Commit) > 7 {
					Commit = Commit[:7]
				}
			}
		case "vcs.time":
			if Date == "unknown" && s.Value != "" {
				Date = s.Value
			}
		}
	}
}
