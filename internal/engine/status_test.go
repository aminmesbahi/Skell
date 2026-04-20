package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeStatusFixture(t *testing.T, repoRoot, skillName, version, hash string, pinned bool) {
	t.Helper()
	claudeDir := filepath.Join(repoRoot, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: skillName, Version: version, ContentHash: hash, Pinned: pinned, Registry: "default"},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repoRoot), lf))
}

func TestStatus_NoManifest_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	eng := newWithProvider(&fakeProvider{})
	_, err := eng.Status(repo)
	assert.Error(t, err)
}

func TestStatus_NoLockFile_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	eng := newWithProvider(&fakeProvider{})
	_, err := eng.Status(repo)
	assert.Error(t, err)
}

func TestStatus_PinnedSkill_ReturnsPinnedStatus(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	makeStatusFixture(t, repo, "pdf", "1.0.0", "", true)

	eng := newWithProvider(&fakeProvider{})
	entries, err := eng.Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusPinned, entries[0].Status)
}

func TestStatus_RegistryUnavailable_ReturnsUnknown(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	makeStatusFixture(t, repo, "pdf", "1.0.0", "", false)

	fp := &fakeProvider{getErr: errors.New("not implemented")}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusUnknown, entries[0].Status)
}

func TestStatus_OutdatedSkill_ReturnsOutdated(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	makeStatusFixture(t, repo, "pdf", "1.0.0", "", false)

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "1.1.0", Lifecycle: model.LifecycleStable},
	}}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusOutdated, entries[0].Status)
	assert.Equal(t, "1.1.0", entries[0].Latest)
}

func TestStatus_UpToDate_ReturnsUpToDate(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	makeStatusFixture(t, repo, "pdf", "1.0.0", "", false)

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "1.0.0", Lifecycle: model.LifecycleStable},
	}}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusUpToDate, entries[0].Status)
}

func TestStatus_DeprecatedInRegistry_ReturnsDeprecated(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	makeStatusFixture(t, repo, "old-skill", "1.0.0", "", false)

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "old-skill",
		Metadata: model.SkillMetadata{Version: "1.0.0", Lifecycle: model.LifecycleDeprecated},
	}}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusDeprecated, entries[0].Status)
}

func TestStatus_ArchivedInRegistry_ReturnsArchived(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	makeStatusFixture(t, repo, "archived-skill", "1.0.0", "", false)

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "archived-skill",
		Metadata: model.SkillMetadata{Version: "1.0.0", Lifecycle: model.LifecycleArchived},
	}}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusArchived, entries[0].Status)
}

func TestStatus_RegistrySkillHasNoVersion_ReturnsUnversioned(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	makeStatusFixture(t, repo, "noversion-skill", "1.0.0", "", false)

	// Registry skill with empty version → StatusUnversioned.
	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "noversion-skill",
		Metadata: model.SkillMetadata{Version: "", Lifecycle: model.LifecycleStable},
	}}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusUnversioned, entries[0].Status)
}

func TestStatus_InstalledSkillHasNoVersion_ReturnsMissingMetadata(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	// InstalledSkill with empty version but registry has a version.
	makeStatusFixture(t, repo, "nover-installed", "", "", false)

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "nover-installed",
		Metadata: model.SkillMetadata{Version: "1.0.0", Lifecycle: model.LifecycleStable},
	}}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusMissingMetadata, entries[0].Status)
}

func TestStatus_HashMismatch_ReturnsLocallyModified(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	// Use a content hash that won't match the (nonexistent) skill directory.
	makeStatusFixture(t, repo, "mod-skill", "1.0.0",
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false)
	// Create the skill directory so hasher can run.
	makeInstalledSkill(t, repo, "mod-skill", "---\nname: mod-skill\n---\n")

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "mod-skill",
		Metadata: model.SkillMetadata{Version: "1.0.0"},
	}}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusLocallyModified, entries[0].Status)
}

func TestStatus_UnknownRegistry_ReturnsUnknown(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	// Lock file references a registry alias not in manifest.
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "ghost-skill", Version: "1.0.0", Registry: "nonexistent-registry"},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	fp := &fakeProvider{}
	entries, err := newWithProvider(fp).Status(repo)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, model.StatusUnknown, entries[0].Status)
}
