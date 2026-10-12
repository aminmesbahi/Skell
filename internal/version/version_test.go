package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func reset(t *testing.T) {
	t.Helper()
	v, c, d := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = v, c, d })
	Version, Commit, Date = "dev", "none", "unknown"
}

func TestFillFromBuildInfo_UsesModuleVersionAndVCS(t *testing.T) {
	reset(t)
	fillFromBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.time", Value: "2026-01-02T03:04:05Z"},
		},
	})
	assert.Equal(t, "v1.2.3", Version)
	assert.Equal(t, "0123456", Commit)
	assert.Equal(t, "2026-01-02T03:04:05Z", Date)
}

func TestFillFromBuildInfo_KeepsLdflagsAndIgnoresDevel(t *testing.T) {
	reset(t)
	fillFromBuildInfo(&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}})
	assert.Equal(t, "dev", Version)

	Version, Commit = "v9.9.9", "abc1234"
	fillFromBuildInfo(&debug.BuildInfo{
		Main:     debug.Module{Version: "v1.0.0"},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "ffffffffff"}},
	})
	assert.Equal(t, "v9.9.9", Version)
	assert.Equal(t, "abc1234", Commit)
}
