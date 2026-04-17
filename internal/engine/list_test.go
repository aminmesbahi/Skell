package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_ReadsFromLockFile(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "pdf-processing", Version: "1.2.0"},
			{Name: "code-review", Version: "2.0.0"},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	eng := newWithProvider(nil)
	skills, err := eng.List(repo)

	require.NoError(t, err)
	assert.Len(t, skills, 2)
	assert.Equal(t, "pdf-processing", skills[0].Name)
	assert.Equal(t, "code-review", skills[1].Name)
}

func TestList_FallsBackToScanWhenNoLockFile(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "bare-skill", "# skill")

	eng := newWithProvider(nil)
	skills, err := eng.List(repo)

	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "bare-skill", skills[0].Name)
}

func TestList_EmptyRepo_ReturnsEmpty(t *testing.T) {
	repo := makeRepo(t)

	eng := newWithProvider(nil)
	skills, err := eng.List(repo)

	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestListRegistry_ReturnsSkillsFromProvider(t *testing.T) {
	repo := makeRepo(t)
	fp := &fakeProvider{
		listSkills: []model.RegistrySkill{{Name: "pdf-processing"}},
	}
	eng := newWithProvider(fp)

	makeManifestWithRegistry(t, repo, "default", "https://example.com/registry")
	m, err := manifest.Resolve(repo)
	require.NoError(t, err)

	skills, err := eng.ListRegistry(m)
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "pdf-processing", skills[0].Name)
}

func TestListRegistry_PropagatesProviderError(t *testing.T) {
	repo := makeRepo(t)
	fp := &fakeProvider{listErr: errors.New("registry: list skills not yet implemented")}
	eng := newWithProvider(fp)

	makeManifestWithRegistry(t, repo, "default", "https://example.com/registry")
	m, err := manifest.Resolve(repo)
	require.NoError(t, err)

	_, err = eng.ListRegistry(m)
	assert.Error(t, err)
}
