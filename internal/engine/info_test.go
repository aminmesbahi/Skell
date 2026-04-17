package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo_LocalSkillWithLockfile(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "pdf-processing",
		"---\nname: pdf-processing\ndescription: PDF tool.\nmetadata:\n  version: \"1.2.0\"\n  owner: test-team\n  lifecycle: stable\n---\n")

	skillDir := filepath.Join(repo, ".claude", "skills", "pdf-processing")
	hash, err := hasher.HashDir(skillDir)
	require.NoError(t, err)

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{
				Name:        "pdf-processing",
				Version:     "1.2.0",
				Registry:    "default",
				InstalledAt: "2026-04-12T10:00:00Z",
				ContentHash: hash,
			},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	eng := newWithProvider(nil)
	result, err := eng.Info(repo, "pdf-processing", "local")
	require.NoError(t, err)

	require.NotNil(t, result.Skill)
	assert.Equal(t, "pdf-processing", result.Skill.Name)
	assert.Equal(t, "1.2.0", result.Skill.Metadata.Version)
	assert.Equal(t, model.LifecycleStable, result.Skill.Metadata.Lifecycle)

	require.NotNil(t, result.Lock)
	assert.Equal(t, "1.2.0", result.Lock.Version)
	assert.Equal(t, model.StatusUpToDate, result.Status)
}

func TestInfo_LocalSkillWithoutLockfile(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "code-review",
		"---\nname: code-review\ndescription: Review.\nmetadata:\n  version: \"2.0.0\"\n---\n")

	eng := newWithProvider(nil)
	result, err := eng.Info(repo, "code-review", "")
	require.NoError(t, err)

	require.NotNil(t, result.Skill)
	assert.Equal(t, "code-review", result.Skill.Name)
	assert.Nil(t, result.Lock, "no lock file means Lock should be nil")
}

func TestInfo_SkillNotFound(t *testing.T) {
	repo := makeRepo(t)

	eng := newWithProvider(nil)
	_, err := eng.Info(repo, "nonexistent-skill", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-skill")
}

func TestInfo_LocallyModifiedSkill(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "pdf-processing",
		"---\nname: pdf-processing\nmetadata:\n  version: \"1.2.0\"\n---\n")

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{
				Name:        "pdf-processing",
				Version:     "1.2.0",
				ContentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	eng := newWithProvider(nil)
	result, err := eng.Info(repo, "pdf-processing", "")
	require.NoError(t, err)

	assert.Equal(t, model.StatusLocallyModified, result.Status)
}

func TestInfo_UpToDateStatus(t *testing.T) {
	repo := makeRepo(t)
	content := "---\nname: pdf-processing\nmetadata:\n  version: \"1.2.0\"\n---\n"
	makeInstalledSkill(t, repo, "pdf-processing", content)

	skillDir := filepath.Join(repo, ".claude", "skills", "pdf-processing")
	realHash, err := hasher.HashDir(skillDir)
	require.NoError(t, err)

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "pdf-processing", Version: "1.2.0", ContentHash: realHash},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	eng := newWithProvider(nil)
	result, err := eng.Info(repo, "pdf-processing", "")
	require.NoError(t, err)

	assert.Equal(t, model.StatusUpToDate, result.Status)
}

func TestInfo_OnlyLockfileEntry(t *testing.T) {
	repo := makeRepo(t)

	// Create lock file but no SKILL.md — skill dir doesn't exist
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "legacy-skill", Version: "0.9.0", InstalledAt: "2026-01-01T00:00:00Z"},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	eng := newWithProvider(nil)
	result, err := eng.Info(repo, "legacy-skill", "")
	require.NoError(t, err)

	assert.Nil(t, result.Skill, "no SKILL.md present, Skill should be nil")
	require.NotNil(t, result.Lock)
	assert.Equal(t, "0.9.0", result.Lock.Version)
}
