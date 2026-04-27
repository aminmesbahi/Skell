package engine

import (
	"errors"
	"testing"

	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeUpgradeFixture(t *testing.T, skillName, installedVersion string) string {
	t.Helper()
	repo := makeRepo(t)
	makePinnableSkill(t, repo, skillName, installedVersion)
	return repo
}

func TestUpgrade_NoLockFile_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	eng := newWithProvider(&fakeProvider{})
	_, err := eng.Upgrade(repo, "", false, false)
	assert.Error(t, err)
}

func TestUpgrade_SkillNotInstalled_ReturnsError(t *testing.T) {
	repo := makeUpgradeFixture(t, "pdf", "1.0.0")
	eng := newWithProvider(&fakeProvider{})
	_, err := eng.Upgrade(repo, "missing-skill", false, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestUpgrade_PinnedSkill_IsSkipped(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "pdf", "1.0.0")

	// Pin it manually.
	eng := newWithProvider(&fakeProvider{})
	require.NoError(t, eng.Pin(repo, "pdf", ""))

	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "1.1.0"},
	}}
	report, err := newWithProvider(fp).Upgrade(repo, "pdf", false, false)
	require.NoError(t, err)
	assert.Empty(t, report.Upgraded)
	assert.Len(t, report.Skipped, 1)
	assert.Contains(t, report.Skipped[0], "pinned")
}

func TestUpgrade_AlreadyUpToDate_IsSkipped(t *testing.T) {
	repo := makeUpgradeFixture(t, "pdf", "1.0.0")
	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "1.0.0"},
	}}
	report, err := newWithProvider(fp).Upgrade(repo, "pdf", false, false)
	require.NoError(t, err)
	assert.Empty(t, report.Upgraded)
	assert.Contains(t, report.Skipped[0], "already up-to-date")
}

func TestUpgrade_DryRun_ReportsWithoutWriting(t *testing.T) {
	repo := makeUpgradeFixture(t, "pdf", "1.0.0")
	fp := &fakeProvider{skill: &model.RegistrySkill{
		Name:     "pdf",
		Metadata: model.SkillMetadata{Version: "2.0.0"},
	}}
	report, err := newWithProvider(fp).Upgrade(repo, "pdf", false, true)
	require.NoError(t, err)
	assert.Len(t, report.Upgraded, 1)
	assert.Contains(t, report.Upgraded[0], "1.0.0")
	assert.Contains(t, report.Upgraded[0], "2.0.0")
	// No files should be written.
	assert.Equal(t, 0, fp.copyCalls)
}

func TestUpgrade_RegistryError_ReturnsError(t *testing.T) {
	repo := makeUpgradeFixture(t, "pdf", "1.0.0")
	fp := &fakeProvider{getErr: errors.New("registry unavailable")}
	_, err := newWithProvider(fp).Upgrade(repo, "pdf", false, false)
	assert.Error(t, err)
}

// TestUpgrade_BothUnversioned_DoesNotShortCircuit: when both the locked and
// the registry skill have empty versions, upgrade must proceed (re-copy)
// rather than report "already up-to-date".
func TestUpgrade_BothUnversioned_DoesNotShortCircuit(t *testing.T) {
	repo := makeUpgradeFixture(t, "rolling-skill", "")
	fp := &fakeProvider{skill: &model.RegistrySkill{Name: "rolling-skill"}}
	report, err := newWithProvider(fp).Upgrade(repo, "rolling-skill", false, true)
	require.NoError(t, err)
	assert.Empty(t, report.Skipped)
	assert.Len(t, report.Upgraded, 1)
}
