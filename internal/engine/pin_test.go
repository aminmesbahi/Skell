package engine

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

// makePinnableSkill creates a manifest + lock file with a single installed skill.
func makePinnableSkill(t *testing.T, repoRoot, skillName, version string) {
	t.Helper()
	claudeDir := filepath.Join(repoRoot, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills: map[string]manifest.SkillEntry{
			skillName: {Version: version, Registry: "default"},
		},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repoRoot), m))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: skillName, Version: version, Registry: "default"},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repoRoot), lf))
}

func TestPin_SetsFlag(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "pdf-processing", "1.2.0")

	eng := newWithProvider(nil)
	require.NoError(t, eng.Pin(repo, "pdf-processing", ""))

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.True(t, m.Skills["pdf-processing"].Pinned)

	lf, err := lockfile.Read(lockfile.Path(repo))
	require.NoError(t, err)
	locked := lf.FindSkill("pdf-processing")
	require.NotNil(t, locked)
	assert.True(t, locked.Pinned)
}

func TestPin_WithExplicitVersion(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "pdf-processing", "1.2.0")

	eng := newWithProvider(nil)
	require.NoError(t, eng.Pin(repo, "pdf-processing", "1.1.0"))

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Equal(t, "1.1.0", m.Skills["pdf-processing"].Version)
	assert.True(t, m.Skills["pdf-processing"].Pinned)
}

func TestPin_SkillNotInManifest_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "other-skill", "1.0.0")

	eng := newWithProvider(nil)
	err := eng.Pin(repo, "missing-skill", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in manifest")
}

func TestUnpin_ClearsFlag(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "code-review", "2.0.0")

	// First pin it.
	eng := newWithProvider(nil)
	require.NoError(t, eng.Pin(repo, "code-review", ""))

	// Then unpin.
	require.NoError(t, eng.Unpin(repo, "code-review"))

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.False(t, m.Skills["code-review"].Pinned)

	lf, err := lockfile.Read(lockfile.Path(repo))
	require.NoError(t, err)
	locked := lf.FindSkill("code-review")
	require.NotNil(t, locked)
	assert.False(t, locked.Pinned)
}

func TestUnpin_SkillNotInManifest_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "other-skill", "1.0.0")

	eng := newWithProvider(nil)
	err := eng.Unpin(repo, "missing-skill")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in manifest")
}
