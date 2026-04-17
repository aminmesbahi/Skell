package skell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeInstalledSkillCmd(t *testing.T, repoRoot, name, content string) {
	t.Helper()
	skillDir := filepath.Join(repoRoot, ".claude", "skills", name)
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0600))
}

func TestInfoCmd_DisplaysSkillMetadata(t *testing.T) {
	repo := t.TempDir()
	makeInstalledSkillCmd(t, repo, "pdf-processing",
		"---\nname: pdf-processing\ndescription: Extract PDFs.\nmetadata:\n  version: \"1.2.0\"\n  owner: platform-team\n  lifecycle: stable\n---\n")

	// Change working directory so the command finds the repo
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	out, err := executeCmd(t, "info", "pdf-processing")
	require.NoError(t, err)

	assert.Contains(t, out, "pdf-processing")
	assert.Contains(t, out, "1.2.0")
	assert.Contains(t, out, "platform-team")
	assert.Contains(t, out, "stable")
}

func TestInfoCmd_ShowsLockfileData(t *testing.T) {
	repo := t.TempDir()
	makeInstalledSkillCmd(t, repo, "code-review",
		"---\nname: code-review\nmetadata:\n  version: \"2.0.0\"\n---\n")

	claudeDir := filepath.Join(repo, ".claude")
	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "code-review", Version: "2.0.0", InstalledAt: "2026-04-12T10:00:00Z"},
		},
	}
	require.NoError(t, lockfile.Write(filepath.Join(claudeDir, "skell.lock"), lf))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	out, err := executeCmd(t, "info", "code-review")
	require.NoError(t, err)

	assert.Contains(t, out, "2.0.0")
	assert.Contains(t, out, "2026-04-12T10:00:00Z")
}

func TestInfoCmd_UnknownSkill_ReturnsError(t *testing.T) {
	repo := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	_, err = executeCmd(t, "info", "does-not-exist")
	assert.Error(t, err)
}

func TestInfoCmd_RequiresSkillNameArg(t *testing.T) {
	_, err := executeCmd(t, "info")
	assert.Error(t, err)
}
