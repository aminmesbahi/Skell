package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProvider implements RegistryProvider for testing.
type fakeProvider struct {
	skill      *model.RegistrySkill
	getErr     error
	copyErr    error
	copyCalls  int
	listSkills []model.RegistrySkill
	listErr    error
}

func (f *fakeProvider) GetSkill(_ registry.Registry, _ string) (*model.RegistrySkill, error) {
	return f.skill, f.getErr
}

func (f *fakeProvider) CopySkillTo(_ registry.Registry, name, _, destPath string) error {
	f.copyCalls++
	if f.copyErr != nil {
		return f.copyErr
	}
	// Create a minimal SKILL.md so the hasher has something to hash.
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(destPath, "SKILL.md"),
		[]byte("---\nname: "+name+"\n---\n"), 0600)
}

func (f *fakeProvider) ListSkills(_ registry.Registry) ([]model.RegistrySkill, error) {
	return f.listSkills, f.listErr
}

func makeManifestWithRegistry(t *testing.T, repoRoot, alias, url string) {
	t.Helper()
	claudeDir := filepath.Join(repoRoot, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{
		Registries: map[string]string{alias: url},
		Skills:     map[string]manifest.SkillEntry{},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repoRoot), m))
}

func TestInstall_NoManifest_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	eng := newWithProvider(&fakeProvider{})
	err := eng.Install(repo, "pdf-processing", "default", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}

func TestInstall_UnknownRegistryAlias_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")

	eng := newWithProvider(&fakeProvider{skill: &model.RegistrySkill{Name: "pdf-processing"}})
	err := eng.Install(repo, "pdf-processing", "nonexistent-alias", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-alias")
}

func TestInstall_ProviderGetSkillError_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")

	provider := &fakeProvider{getErr: fmt.Errorf("network unreachable")}
	eng := newWithProvider(provider)
	err := eng.Install(repo, "pdf-processing", "default", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pdf-processing")
}

func TestInstall_ProviderCopyError_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")

	provider := &fakeProvider{
		skill:   &model.RegistrySkill{Name: "pdf-processing", Metadata: model.SkillMetadata{Version: "1.2.0"}},
		copyErr: fmt.Errorf("disk full"),
	}
	eng := newWithProvider(provider)
	err := eng.Install(repo, "pdf-processing", "default", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pdf-processing")
}

func TestInstall_DryRun_WritesNoFiles(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")

	provider := &fakeProvider{
		skill: &model.RegistrySkill{Name: "pdf-processing", Metadata: model.SkillMetadata{Version: "1.2.0"}},
	}
	eng := newWithProvider(provider)
	require.NoError(t, eng.Install(repo, "pdf-processing", "default", true))

	// No files should have been copied
	assert.Equal(t, 0, provider.copyCalls)

	// No lock file should have been created
	_, err := os.Stat(lockfile.Path(repo))
	assert.True(t, os.IsNotExist(err), "lock file should not exist after dry-run")
}

func TestInstall_Success_CreatesLockAndManifestEntries(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/registry")

	provider := &fakeProvider{
		skill: &model.RegistrySkill{
			Name:        "pdf-processing",
			Description: "PDF tool",
			Metadata: model.SkillMetadata{
				Version:    "1.2.0",
				SourceRepo: "https://example.com/registry",
			},
		},
	}
	eng := newWithProvider(provider)
	require.NoError(t, eng.Install(repo, "pdf-processing", "default", false))

	// Skill directory should exist
	skillDir := filepath.Join(repo, ".claude", "skills", "pdf-processing")
	_, err := os.Stat(skillDir)
	require.NoError(t, err, "skill directory should be created")

	// Lock file should record the install
	lf, err := lockfile.Read(lockfile.Path(repo))
	require.NoError(t, err)
	entry := lf.FindSkill("pdf-processing")
	require.NotNil(t, entry)
	assert.Equal(t, "1.2.0", entry.Version)
	assert.Equal(t, "default", entry.Registry)
	assert.NotEmpty(t, entry.ContentHash)
	assert.Equal(t, "https://example.com/registry", entry.SourceRepo)

	// Manifest should record the install
	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	me, ok := m.Skills["pdf-processing"]
	require.True(t, ok)
	assert.Equal(t, "1.2.0", me.Version)
	assert.Equal(t, "default", me.Registry)
}

func TestInstall_DefaultsToDefaultRegistry(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/registry")

	provider := &fakeProvider{
		skill: &model.RegistrySkill{
			Name:     "my-skill",
			Metadata: model.SkillMetadata{Version: "1.0.0"},
		},
	}
	eng := newWithProvider(provider)
	// Pass empty registry alias — should default to "default"
	require.NoError(t, eng.Install(repo, "my-skill", "", false))

	lf, err := lockfile.Read(lockfile.Path(repo))
	require.NoError(t, err)
	assert.NotNil(t, lf.FindSkill("my-skill"))
}

func TestInstall_UpdatesExistingLockFile(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")

	// Pre-populate the lock file with an existing skill
	existingLF := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "existing-skill", Version: "0.5.0"},
		},
	}
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".claude"), 0755))
	require.NoError(t, lockfile.Write(lockfile.Path(repo), existingLF))

	provider := &fakeProvider{
		skill: &model.RegistrySkill{
			Name:     "new-skill",
			Metadata: model.SkillMetadata{Version: "2.0.0"},
		},
	}
	eng := newWithProvider(provider)
	require.NoError(t, eng.Install(repo, "new-skill", "default", false))

	lf, err := lockfile.Read(lockfile.Path(repo))
	require.NoError(t, err)
	assert.Len(t, lf.Skills, 2, "existing skill should be preserved")
	assert.NotNil(t, lf.FindSkill("existing-skill"))
	assert.NotNil(t, lf.FindSkill("new-skill"))
}
