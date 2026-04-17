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

func makeDoctorRepo(t *testing.T) string {
	t.Helper()
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     map[string]manifest.SkillEntry{"pdf": {Version: "1.0.0", Registry: "default"}},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills:       []model.InstalledSkill{{Name: "pdf", Version: "1.0.0", Registry: "default"}},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	makeInstalledSkill(t, repo, "pdf", "---\nname: pdf\n---\n")
	return repo
}

func TestDoctor_CleanRepo_ReturnsNoIssues(t *testing.T) {
	repo := makeDoctorRepo(t)
	issues, err := newWithProvider(nil).Doctor(repo)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestDoctor_NoManifest_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	issues, err := newWithProvider(nil).Doctor(repo)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, SeverityError, issues[0].Severity)
	assert.Equal(t, "no-manifest", issues[0].Code)
}

func TestDoctor_NoLockFile_ReturnsWarning(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{Registries: map[string]string{}, Skills: map[string]manifest.SkillEntry{}}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	issues, err := newWithProvider(nil).Doctor(repo)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, SeverityWarning, issues[0].Severity)
	assert.Equal(t, "no-lockfile", issues[0].Code)
}

func TestDoctor_MissingSkillDir_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     map[string]manifest.SkillEntry{"pdf": {Version: "1.0.0", Registry: "default"}},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))
	lf := &lockfile.LockFile{
		Skills: []model.InstalledSkill{{Name: "pdf", Version: "1.0.0", Registry: "default"}},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	// No skill directory created.

	issues, err := newWithProvider(nil).Doctor(repo)
	require.NoError(t, err)
	found := false
	for _, i := range issues {
		if i.Code == "missing-dir" {
			found = true
			assert.Equal(t, SeverityError, i.Severity)
		}
	}
	assert.True(t, found, "expected missing-dir issue")
}

func TestDoctor_UntrackedSkill_ReturnsWarning(t *testing.T) {
	repo := makeDoctorRepo(t)
	// Install an extra skill not in the manifest.
	makeInstalledSkill(t, repo, "extra-skill", "---\nname: extra-skill\n---\n")

	issues, err := newWithProvider(nil).Doctor(repo)
	require.NoError(t, err)
	found := false
	for _, i := range issues {
		if i.Code == "untracked-skill" {
			found = true
			assert.Equal(t, SeverityWarning, i.Severity)
		}
	}
	assert.True(t, found, "expected untracked-skill issue")
}

func TestDoctor_HashMismatch_ReturnsWarning(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     map[string]manifest.SkillEntry{"pdf": {Version: "1.0.0", Registry: "default"}},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))
	lf := &lockfile.LockFile{
		Skills: []model.InstalledSkill{{Name: "pdf", Version: "1.0.0", Registry: "default", ContentHash: "bad-hash"}},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	makeInstalledSkill(t, repo, "pdf", "---\nname: pdf\n---\n")

	issues, err := newWithProvider(nil).Doctor(repo)
	require.NoError(t, err)
	found := false
	for _, i := range issues {
		if i.Code == "locally-modified" {
			found = true
			assert.Equal(t, SeverityWarning, i.Severity)
		}
	}
	assert.True(t, found, "expected locally-modified issue for hash mismatch")
}
