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

// ── Info extra paths ──────────────────────────────────────────────────────────

func TestInfo_RegistrySource_Found(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf-processing",
		Metadata: model.SkillMetadata{Version: "1.2.0"},
	}}
	eng := newWithProvider(fp)

	result, err := eng.Info(repo, "pdf-processing", "registry")
	require.NoError(t, err)
	require.NotNil(t, result.Skill)
	assert.Equal(t, "pdf-processing", result.Skill.Name)
}

func TestInfo_RegistrySource_NotFound(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{getErr: errors.New("not found in registry")}
	eng := newWithProvider(fp)

	_, err := eng.Info(repo, "missing-skill", "registry")
	assert.Error(t, err)
}

func TestInfo_RegistrySource_NoManifest(t *testing.T) {
	repo := makeRepo(t)
	eng := newWithProvider(nil)

	_, err := eng.Info(repo, "skill", "registry")
	assert.Error(t, err)
}

func TestInfo_FallbackToRegistry_WhenNotFoundLocally(t *testing.T) {
	// source is "" — first tries local, falls through to registry lookup.
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "remote-skill",
		Metadata: model.SkillMetadata{Version: "3.0.0"},
	}}
	eng := newWithProvider(fp)

	result, err := eng.Info(repo, "remote-skill", "")
	require.NoError(t, err)
	require.NotNil(t, result.Skill)
	assert.Equal(t, "remote-skill", result.Skill.Name)
}

// ── Upgrade extra paths ───────────────────────────────────────────────────────

func TestUpgrade_UnknownRegistry_IsSkipped(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	// Lock file has skill from an alias not in manifest.
	claudeDir := makeClaudeDir(t, repo)
	_ = claudeDir
	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "old-skill", Version: "1.0.0", Registry: "nonexistent-registry"},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	eng := newWithProvider(&fakeProvider{})
	report, err := eng.Upgrade(repo, "", false, false)
	require.NoError(t, err)
	require.Len(t, report.Skipped, 1)
	assert.Contains(t, report.Skipped[0], "unknown registry")
}

func TestUpgrade_Force_UpgradesLocally_Modified(t *testing.T) {
	repo := makeUpgradeFixture(t, "pdf", "1.0.0")
	// Also create the actual skill directory so hasher.Verify can run.
	makeInstalledSkill(t, repo, "pdf", "---\nname: pdf\n---\n")

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "2.0.0"},
	}}

	// Set a wrong content hash to simulate a locally modified skill.
	lf, err := lockfile.Read(lockfile.Path(repo))
	require.NoError(t, err)
	for i, s := range lf.Skills {
		if s.Name == "pdf" {
			lf.Skills[i].ContentHash = "sha256:wronghashwronghashwronghashwronghashwronghashwronghashwronghashwronghash"
		}
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	// With force=true the local modification should be overwritten.
	report, err := newWithProvider(fp).Upgrade(repo, "pdf", true, false)
	require.NoError(t, err)
	assert.Len(t, report.Upgraded, 1)
}

func TestUpgrade_LocallyModified_WithoutForce_ReturnsError(t *testing.T) {
	repo := makeUpgradeFixture(t, "pdf", "1.0.0")
	// Also create the actual skill directory so hasher.Verify can run.
	makeInstalledSkill(t, repo, "pdf", "---\nname: pdf\n---\n")

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "2.0.0"},
	}}

	// Set a wrong content hash to simulate a locally modified skill.
	lf, err := lockfile.Read(lockfile.Path(repo))
	require.NoError(t, err)
	for i, s := range lf.Skills {
		if s.Name == "pdf" {
			lf.Skills[i].ContentHash = "sha256:wronghashwronghashwronghashwronghashwronghashwronghashwronghashwronghash"
		}
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	// Without force, should return error about local modifications.
	_, err = newWithProvider(fp).Upgrade(repo, "pdf", false, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "local modifications")
}

func TestUpgrade_ActualUpgrade_WritesFiles(t *testing.T) {
	repo := makeUpgradeFixture(t, "pdf", "1.0.0")
	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "2.0.0"},
	}}

	report, err := newWithProvider(fp).Upgrade(repo, "pdf", false, false)
	require.NoError(t, err)
	assert.Len(t, report.Upgraded, 1)
	assert.Contains(t, report.Upgraded[0], "1.0.0")
	assert.Contains(t, report.Upgraded[0], "2.0.0")
	assert.Equal(t, 1, fp.copyCalls)
}

// ── Install extra paths ───────────────────────────────────────────────────────

func TestInstall_AutoAddsRegistry_WhenURLProvided(t *testing.T) {
	repo := makeRepo(t)
	// Create a manifest with NO registries.
	claudeDir := makeClaudeDir(t, repo)
	_ = claudeDir
	m := &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "auto-skill",
		Metadata: model.SkillMetadata{Version: "1.0.0"},
	}}
	eng := newWithProvider(fp)

	err := eng.Install(repo, "auto-skill", "new-reg", "https://new.example.com/reg", false)
	require.NoError(t, err)

	// Verify registry was added to manifest.
	updated, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Equal(t, "https://new.example.com/reg", updated.Registries["new-reg"])
}

func TestInstall_MissingRegistry_NoURL_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := makeClaudeDir(t, repo)
	_ = claudeDir
	m := &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	eng := newWithProvider(&fakeProvider{})
	err := eng.Install(repo, "skill", "missing-reg", "", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// ── helper ────────────────────────────────────────────────────────────────────

func makeClaudeDir(t *testing.T, repoRoot string) string {
	t.Helper()
	claudeDir := filepath.Join(repoRoot, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	return claudeDir
}
