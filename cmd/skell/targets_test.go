package skell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCmd_TargetFlag_CreatesInChosenDir(t *testing.T) {
	cases := []struct {
		target string
		dir    string
	}{
		{"codex", ".codex"},
		{"copilot", ".github"},
		{"cursor", ".cursor"},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			repo := t.TempDir()
			out, err := executeCmd(t, "init", "--repo", repo, "--target", tc.target)
			require.NoError(t, err)
			assert.Contains(t, out, "skell.toml created")
			path := filepath.Join(repo, tc.dir, "skell.toml")
			_, statErr := os.Stat(path)
			assert.NoError(t, statErr, "skell.toml should exist in %s", path)
		})
	}
}

func TestInitCmd_TargetFlag_UnknownReturnsError(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "init", "--repo", repo, "--target", "no-such-agent")
	assert.Error(t, err)
}

func TestInitCmd_AutoDetectsExistingLayout(t *testing.T) {
	repo := t.TempDir()
	// Pre-create a .cursor/skills folder to simulate an existing layout.
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".cursor", "skills"), 0o755))

	out, err := executeCmd(t, "init", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "detected layout: cursor")
	_, statErr := os.Stat(filepath.Join(repo, ".cursor", "skell.toml"))
	assert.NoError(t, statErr)
}

func TestTargetsCmd_ListsAllPlatforms(t *testing.T) {
	out, err := executeCmd(t, "targets")
	require.NoError(t, err)
	for _, want := range []string{"claude", "codex", "copilot", "cursor"} {
		assert.Contains(t, out, want)
	}
}
