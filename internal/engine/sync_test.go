package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeSyncFixture creates a repo with a manifest declaring one skill and one skill installed on disk.
// manifestSkill and installedSkill may differ to test sync behaviour.
func makeSyncFixture(t *testing.T, manifestSkills, installedSkills []string) string {
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
	return repo
}

func TestSync_AlreadyInSync_ReturnsEmptyReport(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, false)
	require.NoError(t, err)
	assert.Empty(t, report.Installed)
	assert.Empty(t, report.Removed)
}

func TestSync_DryRun_ReportsDiffWithoutWriting(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf", "code-review"}, []string{"pdf", "old-skill"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, true)
	require.NoError(t, err)
	assert.Contains(t, report.Installed, "code-review")
	assert.Contains(t, report.Removed, "old-skill")
	// dry-run: no files written (old-skill dir still present)
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "old-skill"))
	assert.NoError(t, statErr, "dry-run should not delete files")
}

func TestSync_CheckOnly_ReturnsDiffError(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf", "extra-skill"})
	fp := &fakeProvider{}
	_, err := newWithProvider(fp).Sync(repo, true, false)
	require.Error(t, err)
	var diffErr *SyncDiffError
	assert.ErrorAs(t, err, &diffErr)
	assert.Contains(t, diffErr.Extra, "extra-skill")
}

func TestSync_CheckOnly_NoChange_ReturnsNil(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, true, false)
	require.NoError(t, err)
	assert.Empty(t, report.Installed)
	assert.Empty(t, report.Removed)
}

func TestSync_RemovesExtraSkills(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{"pdf", "old-skill"})
	fp := &fakeProvider{}
	report, err := newWithProvider(fp).Sync(repo, false, false)
	require.NoError(t, err)
	assert.Contains(t, report.Removed, "old-skill")
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "old-skill"))
	assert.True(t, os.IsNotExist(statErr), "old-skill directory should be removed")
}

func TestSync_InstallMissingFails_WhenRegistryUnavailable(t *testing.T) {
	repo := makeSyncFixture(t, []string{"pdf"}, []string{})
	fp := &fakeProvider{getErr: errors.New("registry unavailable")}
	_, err := newWithProvider(fp).Sync(repo, false, false)
	assert.Error(t, err)
}
