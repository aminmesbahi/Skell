package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// publishRisky commits a new upstream version of "demo" that ships a script
// and pre-approves Bash.
func (f *gitRegistryFixture) publishRisky(t *testing.T) {
	t.Helper()
	dir := filepath.Join(f.upstream, "skills", "demo")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "scripts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts", "run.sh"), []byte("echo hi\n"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: demo\ndescription: demo skill\nallowed-tools: Bash\n---\nv2\n"), 0o644))
	git(t, f.upstream, "add", ".")
	git(t, f.upstream, "commit", "-qm", "risky")
	git(t, f.cacheClone, "fetch", "-q", "origin")
	git(t, f.cacheClone, "reset", "-q", "--hard", "FETCH_HEAD")
}

func TestUpgrade_GateDeclinesNewScriptsAndTools(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	f.publishRisky(t)

	var seen []UpgradeRisk
	f.eng.SetUpgradeGate(func(r UpgradeRisk) bool { seen = append(seen, r); return false })

	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	assert.Empty(t, report.Upgraded)
	require.Len(t, report.Skipped, 1)
	assert.Contains(t, report.Skipped[0], "needs review")
	require.Len(t, report.NeedsReview, 1)
	require.Len(t, seen, 1)
	assert.Equal(t, []string{"scripts/run.sh"}, seen[0].NewScripts)
	assert.NotEmpty(t, seen[0].NewToolWarnings)
	assert.Contains(t, f.installedBody(t), "v1", "a declined upgrade must leave the installed copy alone")
}

func TestUpgrade_GateApprovalApplies(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	f.publishRisky(t)
	f.eng.SetUpgradeGate(func(UpgradeRisk) bool { return true })

	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	require.Len(t, report.Upgraded, 1)
	assert.Contains(t, f.installedBody(t), "v2")
	assert.FileExists(t, filepath.Join(f.repo, ".claude", "skills", "demo", "scripts", "run.sh"))
}

func TestUpgrade_NoGateStillUpgradesAndReports(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	f.publishRisky(t)

	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	assert.Len(t, report.Upgraded, 1)
	assert.Len(t, report.NeedsReview, 1)
}

func TestUpgrade_DryRunReportsRiskWithoutGate(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	f.publishRisky(t)
	called := false
	f.eng.SetUpgradeGate(func(UpgradeRisk) bool { called = true; return false })

	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, true)
	require.NoError(t, err)
	assert.False(t, called, "dry runs never prompt")
	assert.Len(t, report.Upgraded, 1)
	assert.Len(t, report.NeedsReview, 1)
}

func TestUpgrade_NoNewRiskDoesNotConsultGate(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	f.publish(t, "v2")
	f.eng.SetUpgradeGate(func(UpgradeRisk) bool { t.Fatal("gate must not be consulted"); return false })

	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	assert.Len(t, report.Upgraded, 1)
	assert.Empty(t, report.NeedsReview)
}

func TestUpgrade_AlreadyRiskySkillIsNotReflagged(t *testing.T) {
	f := newGitRegistryFixture(t)
	f.publishRisky(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	// Same script and tools, new prose.
	require.NoError(t, os.WriteFile(filepath.Join(f.upstream, "skills", "demo", "SKILL.md"),
		[]byte("---\nname: demo\ndescription: demo skill\nallowed-tools: Bash\n---\nv3\n"), 0o644))
	git(t, f.upstream, "commit", "-qam", "prose")
	git(t, f.cacheClone, "fetch", "-q", "origin")
	git(t, f.cacheClone, "reset", "-q", "--hard", "FETCH_HEAD")
	f.eng.SetUpgradeGate(func(UpgradeRisk) bool { t.Fatal("gate must not be consulted"); return false })

	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	assert.Len(t, report.Upgraded, 1)
	assert.Empty(t, report.NeedsReview)
}

func TestUpgrade_ForceRestoresLocalEditsWhenUpstreamUnchanged(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	skill := filepath.Join(f.repo, ".claude", "skills", "demo", "SKILL.md")
	orig, err := os.ReadFile(skill)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(skill, append(orig, []byte("my edit\n")...), 0o644))

	// Without --force the edit is kept and the message says why.
	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	require.Len(t, report.Skipped, 1)
	assert.Contains(t, report.Skipped[0], "local edits kept")
	assert.Contains(t, f.installedBody(t), "my edit")

	// --force discards it.
	report, err = f.eng.UpgradeFor(f.repo, "demo", "", true, false)
	require.NoError(t, err)
	assert.Len(t, report.Upgraded, 1)
	assert.NotContains(t, f.installedBody(t), "my edit")
}
