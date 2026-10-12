package skell

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// riskyUpgradeFixture returns a project with "demo" installed from a local
// folder source, and the source folder. riskify() then publishes a v2 that
// ships a script and pre-approves Bash.
func riskyUpgradeFixture(t *testing.T) (repo string, riskify func()) {
	t.Helper()
	src := t.TempDir()
	dir := filepath.Join(src, "skills", "demo")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	write := func(rel, content string, mode os.FileMode) {
		p := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), mode))
	}
	write("SKILL.md", "---\nname: demo\ndescription: demo skill. Use when testing.\n---\nv1\n", 0o644)

	repo = t.TempDir()
	_, err := executeCmd(t, "init", "--repo", repo, "--target", "claude", "--yes")
	require.NoError(t, err)
	_, err = executeCmd(t, "add", src, "--alias", "src", "--repo", repo)
	require.NoError(t, err)
	_, err = executeCmd(t, "install", "demo", "--repo", repo, "--yes")
	require.NoError(t, err)

	return repo, func() {
		write("SKILL.md", "---\nname: demo\ndescription: demo skill. Use when testing.\nallowed-tools: Bash\n---\nv2\n", 0o644)
		write("scripts/run.sh", "echo hi\n", 0o755)
	}
}

func TestUpgradeCmd_ShowsWhatAnUpdateAdds(t *testing.T) {
	repo, riskify := riskyUpgradeFixture(t)
	riskify()

	out, err := executeCmd(t, "upgrade", "--repo", repo, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "this update adds")
	assert.Contains(t, out, "scripts/run.sh")

	// Without a terminal (like here) the risky update is refused, not applied.
	out, err = executeCmd(t, "upgrade", "--repo", repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--yes")
	assert.Contains(t, out, "adds more than the installed copy has")
	assert.NoFileExists(t, filepath.Join(repo, ".claude", "skills", "demo", "scripts", "run.sh"))

	// --yes approves it.
	out, err = executeCmd(t, "upgrade", "--repo", repo, "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "1 skill(s) upgraded")
	assert.FileExists(t, filepath.Join(repo, ".claude", "skills", "demo", "scripts", "run.sh"))
}

func TestMCPUpgrade_RequiresConfirmationForNewRisk(t *testing.T) {
	repo, riskify := riskyUpgradeFixture(t)
	riskify()

	var upgrade func(args string) (any, error)
	for _, tool := range mcpTools(repo) {
		if tool.Name == "skell_upgrade" {
			upgrade = func(args string) (any, error) {
				return tool.Handler(context.Background(), json.RawMessage(args))
			}
		}
	}
	require.NotNil(t, upgrade)

	got, err := upgrade(`{}`)
	require.NoError(t, err)
	res, ok := got.(map[string]any)
	require.True(t, ok, "expected a needs_confirmation result, got %T", got)
	assert.Equal(t, true, res["needs_confirmation"])
	assert.NoFileExists(t, filepath.Join(repo, ".claude", "skills", "demo", "scripts", "run.sh"))

	_, err = upgrade(`{"confirm": true}`)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(repo, ".claude", "skills", "demo", "scripts", "run.sh"))
}
