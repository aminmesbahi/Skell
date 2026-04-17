package skell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeRepoWithSkill creates a manifest + lockfile for a single pinnable skill.
func makeRepoWithSkill(t *testing.T, skillName, version string) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills: map[string]manifest.SkillEntry{
			skillName: {Version: version, Registry: "default"},
		},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills:       []model.InstalledSkill{{Name: skillName, Version: version, Registry: "default"}},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	return repo
}

func TestPinCmd_PinsSkill(t *testing.T) {
	repo := makeRepoWithSkill(t, "pdf-processing", "1.2.0")

	out, err := executeCmd(t, "pin", "--repo", repo, "pdf-processing")
	require.NoError(t, err)
	assert.Contains(t, out, "pin") // action label "pin" printed by PrintAction

	m, readErr := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, readErr)
	assert.True(t, m.Skills["pdf-processing"].Pinned)
}

func TestPinCmd_MissingSkill_ReturnsError(t *testing.T) {
	repo := makeRepoWithSkill(t, "other-skill", "1.0.0")

	_, err := executeCmd(t, "pin", "--repo", repo, "missing-skill")
	assert.Error(t, err)
}

func TestUnpinCmd_UnpinsSkill(t *testing.T) {
	repo := makeRepoWithSkill(t, "code-review", "2.0.0")

	// Pin first.
	_, err := executeCmd(t, "pin", "--repo", repo, "code-review")
	require.NoError(t, err)

	// Then unpin.
	out, err := executeCmd(t, "unpin", "--repo", repo, "code-review")
	require.NoError(t, err)
	assert.Contains(t, out, "unpin") // action label "unpin" printed by PrintAction

	m, readErr := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, readErr)
	assert.False(t, m.Skills["code-review"].Pinned)
}
