package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeSyncFixture creates a repo with a manifest declaring one skill and one skill installed on disk.
// manifestSkill and installedSkill may differ to test sync behaviour.
func makeSyncFixture(t *testing.T, manifestSkills, installedSkills []string) string {
	t.Helper()
	return makeSyncFixtureWithLock(t, manifestSkills, installedSkills, nil)
}

// makeSyncFixtureWithLock is like makeSyncFixture but also writes a lock file
// recording lockedSkills as previously installed by Skell.
func makeSyncFixtureWithLock(t *testing.T, manifestSkills, installedSkills, lockedSkills []string) string {
	t.Helper()
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	skills := make(map[string]manifest.SkillEntry)
	for _, name := range manifestSkills {
		skills[name] = manifest.SkillEntry{Version: "1.0.0", Registry: "default"}
	}
	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     skills,
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	for _, name := range installedSkills {
		makeInstalledSkill(t, repo, name, "---\nname: "+name+"\n---\n")
	}

	if lockedSkills != nil {
		lf := &lockfile.LockFile{}
		for _, name := range lockedSkills {
			lf.Skills = append(lf.Skills, model.InstalledSkill{Name: name, Registry: "default"})
		}
		require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	}
	return repo
}

func TestSync_AlreadyInSync_ReturnsEmptyReport(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, false, false)
	require.NoError(t, err)
	assert.Empty(t, report.Installed)
	assert.Empty(t, report.Removed)
}

func TestSync_DryRun_ReportsDiffWithoutWriting(t *testing.T) {
	// old-skill is tracked in the lock file, so it is removable on sync.
	repo := makeSyncFixtureWithLock(t, []string{"pdf", "code-review"}, []string{"pdf", "old-skill"}, []string{"pdf", "old-skill"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, true, false)
	require.NoError(t, err)
	assert.Contains(t, report.Installed, "code-review")
	assert.Contains(t, report.Removed, "old-skill")
	// dry-run: no files written (old-skill dir still present)
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "old-skill"))
	assert.NoError(t, statErr, "dry-run should not delete files")
}

func TestSync_CheckOnly_ReturnsDiffError(t *testing.T) {
	// extra-skill is lock-tracked, so it represents managed drift.
	repo := makeSyncFixtureWithLock(t, []string{"pdf"}, []string{"pdf", "extra-skill"}, []string{"pdf", "extra-skill"})
	fp := &fakeProvider{}
	_, err := newWithProvider(fp).Sync(repo, true, false, false)
	require.Error(t, err)
	var diffErr *SyncDiffError
	assert.ErrorAs(t, err, &diffErr)
	assert.Contains(t, diffErr.Extra, "extra-skill")
}

func TestSync_CheckOnly_NoChange_ReturnsNil(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, true, false, false)
	require.NoError(t, err)
	assert.Empty(t, report.Installed)
	assert.Empty(t, report.Removed)
}

func TestSync_RemovesLockTrackedExtraSkills(t *testing.T) {
	repo := makeSyncFixtureWithLock(t, []string{"pdf"}, []string{"pdf", "old-skill"}, []string{"pdf", "old-skill"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, false, false)
	require.NoError(t, err)
	assert.Contains(t, report.Removed, "old-skill")
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "old-skill"))
	assert.True(t, os.IsNotExist(statErr), "lock-tracked old-skill directory should be removed")
}

// Untracked (hand-authored) skills must NOT be deleted by default — they are
// reported instead. This is the core safety guarantee (design §15).
func TestSync_KeepsUntrackedSkills(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf", "hand-authored"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, false, false)
	require.NoError(t, err)
	assert.NotContains(t, report.Removed, "hand-authored")
	assert.Contains(t, report.Untracked, "hand-authored")
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "hand-authored"))
	assert.NoError(t, statErr, "untracked skill must be preserved")
}

// check mode must not flag untracked skills as managed drift.
func TestSync_CheckOnly_IgnoresUntracked(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf", "hand-authored"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, true, false, false)
	require.NoError(t, err)
	assert.Contains(t, report.Untracked, "hand-authored")
}

// With --prune, untracked skills are removed too.
func TestSync_Prune_RemovesUntracked(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf", "hand-authored"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, false, true)
	require.NoError(t, err)
	assert.Contains(t, report.Removed, "hand-authored")
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "hand-authored"))
	assert.True(t, os.IsNotExist(statErr), "prune should remove untracked skill")
}

func TestSync_InstallMissingFails_WhenRegistryUnavailable(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{})
	fp := &fakeProvider{getErr: errors.New("registry unavailable")}
	_, err := newWithProvider(fp).Sync(repo, false, false, false)
	assert.Error(t, err)
}

// lockInstalledHashes records the current content hash of each installed skill
// in the lock file, as an install would.
func lockInstalledHashes(t *testing.T, repo string, names ...string) {
	t.Helper()
	lf := &lockfile.LockFile{}
	for _, name := range names {
		hash, err := hasher.HashDir(filepath.Join(repo, ".claude", "skills", name))
		require.NoError(t, err)
		lf.Skills = append(lf.Skills, model.InstalledSkill{Name: name, Registry: "default", ContentHash: hash})
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
}

func editSkill(t *testing.T, repo, name string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(repo, ".claude", "skills", name, "SKILL.md"), os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteString("edited\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())
}

func TestSync_CheckOnly_ReportsLocallyModifiedSkill(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf", "docx"}, []string{"pdf", "docx"})
	lockInstalledHashes(t, repo, "pdf", "docx")
	_, err := newWithProvider(&fakeProvider{}).Sync(repo, true, false, false)
	require.NoError(t, err, "untouched skills are in sync")

	editSkill(t, repo, "pdf")
	_, err = newWithProvider(&fakeProvider{}).Sync(repo, true, false, false)
	var diffErr *SyncDiffError
	require.ErrorAs(t, err, &diffErr)
	assert.Equal(t, []string{"pdf"}, diffErr.Modified)
	assert.Contains(t, err.Error(), "modified: pdf")
}

func TestSync_ModifiedSkillIsKeptWithoutForce(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf"})
	lockInstalledHashes(t, repo, "pdf")
	editSkill(t, repo, "pdf")

	report, err := newWithProvider(&fakeProvider{}).Sync(repo, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"pdf"}, report.Modified)
	assert.Empty(t, report.Restored)
	data, err := os.ReadFile(filepath.Join(repo, ".claude", "skills", "pdf", "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "edited", "local edits must survive a plain sync")
}

func TestSync_ForceRestoresModifiedSkill(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf"})
	lockInstalledHashes(t, repo, "pdf")
	editSkill(t, repo, "pdf")
	fp := &fakeProvider{skill: &model.RegistrySkill{Name: "pdf"}}

	report, err := newWithProvider(fp).SyncWith(repo, SyncOptions{Force: true})
	require.NoError(t, err)
	assert.Equal(t, []string{"pdf"}, report.Restored)
	assert.Empty(t, report.Installed, "a restore is not a fresh install")
	data, err := os.ReadFile(filepath.Join(repo, ".claude", "skills", "pdf", "SKILL.md"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "edited")
}

func TestSync_ForceDryRunReportsRestoreWithoutWriting(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf"})
	lockInstalledHashes(t, repo, "pdf")
	editSkill(t, repo, "pdf")

	report, err := newWithProvider(&fakeProvider{}).SyncWith(repo, SyncOptions{Force: true, DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, []string{"pdf"}, report.Restored)
	data, err := os.ReadFile(filepath.Join(repo, ".claude", "skills", "pdf", "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "edited")
}

func TestSync_IgnoresModifiedSkillNotInManifest(t *testing.T) {
	// A skill the manifest no longer lists is removed by sync; it must not
	// also be reported as modified.
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf", "old"})
	lockInstalledHashes(t, repo, "pdf", "old")
	editSkill(t, repo, "old")
	_, err := newWithProvider(&fakeProvider{}).Sync(repo, true, false, false)
	var diffErr *SyncDiffError
	require.ErrorAs(t, err, &diffErr)
	assert.Empty(t, diffErr.Modified)
	assert.Equal(t, []string{"old"}, diffErr.Extra)
}
