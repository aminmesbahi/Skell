package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRepoWithRegistry(t *testing.T, alias, url string) string {
	t.Helper()
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, alias, url)
	return repo
}

func TestAddFromURL_InvalidURL_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")
	eng := newWithProvider(&fakeProvider{})
	_, err := eng.AddFromURL(repo, "not-a-url", false)
	assert.Error(t, err)
}

func TestAddFromURL_RegistryRoot_AddsToManifest(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	eng := newWithProvider(&fakeProvider{})
	res, err := eng.AddFromURL(repo, "https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills", false)
	require.NoError(t, err)
	assert.True(t, res.Registered)
	assert.Empty(t, res.SkillName)
	assert.Equal(t, "dotnet-skills", res.Alias)

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/Aaronontheweb/dotnet-skills", m.Registries["dotnet-skills"])
}

func TestAddFromURL_RegistryRoot_DryRun_DoesNotWrite(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	eng := newWithProvider(&fakeProvider{})
	res, err := eng.AddFromURL(repo, "https://github.com/owner/skills-repo/tree/main/skills", true)
	require.NoError(t, err)
	assert.False(t, res.Registered)

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Empty(t, m.Registries, "dry-run must not write to manifest")
}

func TestAddFromURL_RegistryRoot_AlreadyRegistered_NoError(t *testing.T) {
	repo := makeRepoWithRegistry(t, "dotnet-skills", "https://github.com/Aaronontheweb/dotnet-skills")

	eng := newWithProvider(&fakeProvider{})
	res, err := eng.AddFromURL(repo, "https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills", false)
	require.NoError(t, err)
	assert.False(t, res.Registered)
}

func TestAddFromURL_RegistryRoot_NoManifest_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	eng := newWithProvider(&fakeProvider{})
	_, err := eng.AddFromURL(repo, "https://github.com/owner/repo/tree/main/skills", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}

func TestAddFromURL_SpecificSkill_Installs(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	fp := &fakeProvider{
		skill: &model.RegistrySkill{
			Name:     "akka-testing-patterns",
			Metadata: model.SkillMetadata{Version: "1.0.0"},
		},
	}
	eng := newWithProvider(fp)

	res, err := eng.AddFromURL(repo,
		"https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills/akka-testing-patterns",
		false,
	)
	require.NoError(t, err)
	assert.True(t, res.Installed)
	assert.Equal(t, "akka-testing-patterns", res.SkillName)
	assert.Equal(t, "dotnet-skills", res.Alias)

	// Skill files must exist.
	skillDir := filepath.Join(repo, ".claude", "skills", "akka-testing-patterns")
	_, err = os.Stat(filepath.Join(skillDir, "SKILL.md"))
	assert.NoError(t, err)
}

func TestAddFromURL_SpecificSkill_DryRun_NoFiles(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	fp := &fakeProvider{
		skill: &model.RegistrySkill{Name: "akka-testing-patterns"},
	}
	eng := newWithProvider(fp)

	res, err := eng.AddFromURL(repo,
		"https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills/akka-testing-patterns",
		true,
	)
	require.NoError(t, err)
	assert.False(t, res.Installed)

	skillDir := filepath.Join(repo, ".claude", "skills", "akka-testing-patterns")
	_, err = os.Stat(skillDir)
	assert.True(t, os.IsNotExist(err))
}

func TestAddFromURL_SpecificSkill_NoManifest_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	fp := &fakeProvider{skill: &model.RegistrySkill{Name: "my-skill"}}
	eng := newWithProvider(fp)
	_, err := eng.AddFromURL(repo,
		"https://github.com/owner/repo/tree/main/skills/my-skill",
		false,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}

func TestAddFromURL_PlainGitURL_RegistersRegistry(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	eng := newWithProvider(&fakeProvider{})
	res, err := eng.AddFromURL(repo, "https://github.com/davidfowl/dotnet-skillz", false)
	require.NoError(t, err)
	assert.True(t, res.Registered)
	assert.Empty(t, res.SkillName)

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/davidfowl/dotnet-skillz", m.Registries["dotnet-skillz"])
}

// TestAddFromURL_SkillsRootSubdir_RegistersAsRegistry covers the case where
// the URL contains a multi-segment subpath that points to a skills-root
// subdirectory (e.g. /tree/main/ai/claude) rather than a specific skill.
// The engine should detect this and register the repository as a registry
// instead of returning an error.
func TestAddFromURL_SkillsRootSubdir_RegistersAsRegistry(t *testing.T) {
	// Create a fake cache that contains the "ai/claude" subdirectory as if
	// the registry clone already happened.
	cacheRoot := t.TempDir()
	subDir := filepath.Join(cacheRoot, "myrepo", "ai", "claude")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	// fakeProvider returns ErrSkillNotFound so Install fails with the sentinel.
	fp := &fakeProvider{getErr: registry.ErrSkillNotFound}
	eng := &Engine{
		provider:  fp,
		cacheRoot: cacheRoot,
		logger:    defaultAuditLogger(),
		pol:       loadPolicy(),
	}

	res, err := eng.AddFromURL(repo,
		"https://github.com/owner/myrepo/tree/main/ai/claude",
		false,
	)
	require.NoError(t, err)
	assert.True(t, res.Registered)
	assert.Empty(t, res.SkillName, "skills-root URL should not set SkillName")
	assert.Equal(t, "myrepo", res.Alias)
}

func TestAddFromURL_RegistryRoot_NilRegistriesInManifest(t *testing.T) {
	// When the manifest has a nil Registries map (no [registries] section in TOML),
	// AddFromURL must initialise it before writing the new registry entry.
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	// Write manifest with nil Registries (Skills only).
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Skills: map[string]manifest.SkillEntry{},
	}))

	eng := newWithProvider(&fakeProvider{})
	res, err := eng.AddFromURL(repo, "https://github.com/owner/nil-reg-repo/tree/main/skills", false)
	require.NoError(t, err)
	assert.True(t, res.Registered)
	assert.Equal(t, "nil-reg-repo", res.Alias)

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/nil-reg-repo", m.Registries["nil-reg-repo"])
}

func TestAddFromURL_SpecificSkill_InstallError_NotSubpathDir_ReturnsError(t *testing.T) {
	// When Install fails and the cache does not contain the subpath as a directory,
	// AddFromURL returns "add from URL: ..." wrapping the install error.
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "myrepo", "https://example.com/myrepo")

	fp := &fakeProvider{getErr: errors.New("skill not found in registry")}
	eng := newWithProvider(fp) // cacheRoot="" → isSubPathDir always false

	_, err := eng.AddFromURL(repo,
		"https://github.com/owner/myrepo/tree/main/skills/missing-skill",
		false,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "add from URL")
}

func TestAddFromURL_SpecificSkill_NonSentinelInstallError_PropagatesEvenIfSubpathExists(t *testing.T) {
	cacheRoot := t.TempDir()
	subDir := filepath.Join(cacheRoot, "myrepo", "skills", "my-skill")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	repo := makeRepoWithRegistry(t, "myrepo", "https://example.com/myrepo")

	fp := &fakeProvider{getErr: errors.New("network unreachable")}
	eng := &Engine{
		provider:  fp,
		cacheRoot: cacheRoot,
		logger:    defaultAuditLogger(),
		pol:       loadPolicy(),
	}

	_, err := eng.AddFromURL(repo,
		"https://github.com/owner/myrepo/tree/main/skills/my-skill",
		false,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "add from URL")
	assert.Contains(t, err.Error(), "network unreachable")
}
