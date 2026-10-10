package skell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeRiskySource writes a local skill source with one harmless skill ("calm")
// and one that ships a script and pre-approves Bash ("scripty").
func makeRiskySource(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(src, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	write("skills/calm/SKILL.md", "---\nname: calm\ndescription: A harmless skill. Use when calm.\n---\n# Calm\n")
	write("skills/scripty/SKILL.md", "---\nname: scripty\ndescription: Runs a script. Use when asked.\nallowed-tools: Bash\n---\n# Scripty\n")
	write("skills/scripty/scripts/run.sh", "echo hi\n")
	return src
}

func initRepoWithSource(t *testing.T, src string) string {
	t.Helper()
	repo := t.TempDir()
	_, err := executeCmd(t, "init", "--yes", "--target", "claude", "--repo", repo)
	require.NoError(t, err)
	_, err = executeCmd(t, "add", src, "--alias", "src", "--repo", repo)
	require.NoError(t, err)
	return repo
}

func TestInstallCmd_RiskySkillWithoutTerminalNeedsYes(t *testing.T) {
	src := makeRiskySource(t)
	repo := initRepoWithSource(t, src)

	out, err := executeCmd(t, "install", "scripty", "--repo", repo)
	require.Error(t, err, "tests run without a terminal, so nobody can approve")
	assert.Contains(t, err.Error(), "--yes")
	assert.Contains(t, out, "scripts")
	assert.NoDirExists(t, filepath.Join(repo, ".claude", "skills", "scripty"))

	_, err = executeCmd(t, "install", "scripty", "--repo", repo, "--yes")
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(repo, ".claude", "skills", "scripty"))
}

func TestInstallCmd_RiskySkillWithJSONNeedsYes(t *testing.T) {
	src := makeRiskySource(t)
	repo := initRepoWithSource(t, src)

	_, err := executeCmd(t, "install", "scripty", "--repo", repo, "--json")
	require.Error(t, err, "--json has no one to ask either")
	assert.NoDirExists(t, filepath.Join(repo, ".claude", "skills", "scripty"))
}

func TestInstallCmd_HarmlessSkillInstallsWithoutYes(t *testing.T) {
	src := makeRiskySource(t)
	repo := initRepoWithSource(t, src)

	_, err := executeCmd(t, "install", "calm", "--repo", repo)
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(repo, ".claude", "skills", "calm"))
}

func TestInstallCmd_UnknownRegistryBeatsAlreadyInstalled(t *testing.T) {
	src := makeRiskySource(t)
	repo := initRepoWithSource(t, src)
	_, err := executeCmd(t, "install", "calm", "--repo", repo)
	require.NoError(t, err)

	_, err = executeCmd(t, "install", "calm", "--repo", repo, "--registry", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `registry "nope" not configured`)
}

func TestSyncCmd_CheckFailsOnEditedSkillAndForceRestoresIt(t *testing.T) {
	src := makeRiskySource(t)
	repo := initRepoWithSource(t, src)
	_, err := executeCmd(t, "install", "calm", "--repo", repo)
	require.NoError(t, err)

	out, err := executeCmd(t, "sync", "--check", "--repo", repo)
	require.NoError(t, err, out)

	skillFile := filepath.Join(repo, ".claude", "skills", "calm", "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte("tampered\n"), 0o644))

	out, err = executeCmd(t, "sync", "--check", "--repo", repo)
	require.Error(t, err)
	assert.Contains(t, out, "modified calm")

	out, err = executeCmd(t, "sync", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "modified calm")
	data, _ := os.ReadFile(skillFile)
	assert.Equal(t, "tampered\n", string(data), "a plain sync keeps local edits")

	out, err = executeCmd(t, "sync", "--force", "--repo", repo)
	require.NoError(t, err, out)
	assert.Contains(t, out, "restore")
	data, _ = os.ReadFile(skillFile)
	assert.Contains(t, string(data), "A harmless skill")

	_, err = executeCmd(t, "sync", "--check", "--repo", repo)
	require.NoError(t, err)
}

func TestStatusCmd_NoManifest_IsFriendly(t *testing.T) {
	_, err := executeCmd(t, "status", "--repo", t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skell init")
	assert.NotContains(t, err.Error(), "no such file")
}
