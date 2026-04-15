package lockfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead_ValidLockFile(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "skell_version": "0.1.0",
  "locked_at": "2026-04-12T10:00:00Z",
  "skills": [
    {
      "name": "pdf-processing",
      "version": "1.2.0",
      "registry": "default",
      "source_repo": "https://github.com/mycompany/skills-registry",
      "source_ref": "v1.2.0",
      "installed_path": ".claude/skills/pdf-processing",
      "installed_at": "2026-04-12T10:00:00Z",
      "pinned": false,
      "content_hash": "sha256:abc123"
    }
  ]
}`
	path := filepath.Join(dir, "skell.lock")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	lf, err := lockfile.Read(path)
	require.NoError(t, err)
	assert.Equal(t, "0.1.0", lf.SkellVersion)
	require.Len(t, lf.Skills, 1)
	assert.Equal(t, "pdf-processing", lf.Skills[0].Name)
}

func TestRead_MissingFile(t *testing.T) {
	_, err := lockfile.Read("/nonexistent/skell.lock")
	assert.Error(t, err)
}

func TestWrite_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skell.lock")

	original := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		LockedAt:     "2026-04-12T10:00:00Z",
		Skills: []model.InstalledSkill{
			{Name: "pdf-processing", Version: "1.2.0", ContentHash: "sha256:abc"},
		},
	}

	require.NoError(t, lockfile.Write(path, original))

	parsed, err := lockfile.Read(path)
	require.NoError(t, err)
	assert.Equal(t, "0.1.0", parsed.SkellVersion)
	assert.Equal(t, "pdf-processing", parsed.Skills[0].Name)
}

func TestFindSkill_Found(t *testing.T) {
	lf := &lockfile.LockFile{
		Skills: []model.InstalledSkill{
			{Name: "pdf-processing", Version: "1.2.0"},
		},
	}
	skill := lf.FindSkill("pdf-processing")
	require.NotNil(t, skill)
	assert.Equal(t, "1.2.0", skill.Version)
}

func TestFindSkill_NotFound(t *testing.T) {
	lf := &lockfile.LockFile{Skills: []model.InstalledSkill{}}
	assert.Nil(t, lf.FindSkill("nonexistent"))
}

func TestUpsert_AddsNew(t *testing.T) {
	lf := &lockfile.LockFile{}
	lf.Upsert(model.InstalledSkill{Name: "new-skill", Version: "1.0.0"})
	assert.Len(t, lf.Skills, 1)
}

func TestUpsert_ReplacesExisting(t *testing.T) {
	lf := &lockfile.LockFile{
		Skills: []model.InstalledSkill{{Name: "my-skill", Version: "1.0.0"}},
	}
	lf.Upsert(model.InstalledSkill{Name: "my-skill", Version: "2.0.0"})
	assert.Len(t, lf.Skills, 1)
	assert.Equal(t, "2.0.0", lf.Skills[0].Version)
}

func TestRemove_Existing(t *testing.T) {
	lf := &lockfile.LockFile{
		Skills: []model.InstalledSkill{{Name: "my-skill", Version: "1.0.0"}},
	}
	lf.Remove("my-skill")
	assert.Empty(t, lf.Skills)
}

func TestPath(t *testing.T) {
	path := lockfile.Path("/my/repo")
	assert.Equal(t, filepath.Join("/my/repo", ".claude", "skell.lock"), path)
}
