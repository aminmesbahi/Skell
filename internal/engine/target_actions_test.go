package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/aminmesbahi/skell/internal/validator"
	"github.com/stretchr/testify/require"
)

func TestExplicitTargetActionsDoNotChangeOtherInstallation(t *testing.T) {
	repo := t.TempDir()
	for _, id := range []string{"claude", "cursor"} {
		agent, err := target.Lookup(id)
		require.NoError(t, err)
		dir := filepath.Join(agent.SkillsDir(repo), "example")
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: example\ndescription: Test skill\n---\nUse this skill for testing.\n"), 0644))
		require.NoError(t, manifest.Write(manifest.LocalPathFor(repo, agent), &manifest.Manifest{Registries: map[string]string{}, Skills: map[string]manifest.SkillEntry{"example": {Version: "1.0.0"}}}))
		require.NoError(t, lockfile.Write(lockfile.PathFor(repo, agent), &lockfile.LockFile{Skills: []model.InstalledSkill{{Name: "example", Version: "1.0.0", InstalledPath: filepath.Join(agent.Dir, "skills", "example")}}}))
	}
	eng := newWithProvider(nil)
	cursor, err := target.Lookup("cursor")
	require.NoError(t, err)
	claude, err := target.Lookup("claude")
	require.NoError(t, err)
	require.NoError(t, eng.PinFor(repo, "example", "", "cursor"))
	cursorLock, err := lockfile.Read(lockfile.PathFor(repo, cursor))
	require.NoError(t, err)
	claudeLock, err := lockfile.Read(lockfile.PathFor(repo, claude))
	require.NoError(t, err)
	require.True(t, cursorLock.FindSkill("example").Pinned)
	require.False(t, claudeLock.FindSkill("example").Pinned)
	report, err := eng.UpgradeFor(repo, "example", "cursor", false, false)
	require.NoError(t, err)
	require.Equal(t, []string{"example (pinned)"}, report.Skipped)
	require.NoError(t, eng.UnpinFor(repo, "example", "cursor"))
	results, err := eng.ValidateAll(context.Background(), repo, validator.Options{})
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.NotEqual(t, results[0].Target, results[1].Target)
	require.NoError(t, eng.RemoveFor(repo, "example", "cursor", false))
	require.NoDirExists(t, filepath.Join(cursor.SkillsDir(repo), "example"))
	require.DirExists(t, filepath.Join(claude.SkillsDir(repo), "example"))
	m, err := manifest.Read(manifest.LocalPathFor(repo, claude))
	require.NoError(t, err)
	require.Contains(t, m.Skills, "example")
}
