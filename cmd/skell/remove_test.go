package skell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRemoveRepo(t *testing.T, skillName string) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	skillDir := filepath.Join(claudeDir, "skills", skillName)
	require.NoError(t, os.MkdirAll(skillDir, 0755))

	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     map[string]manifest.SkillEntry{skillName: {Version: "1.0.0", Registry: "default"}},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills:       []model.InstalledSkill{{Name: skillName, Version: "1.0.0", Registry: "default"}},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	mdFile := filepath.Join(skillDir, skillName+".md")
	require.NoError(t, os.WriteFile(mdFile, []byte("---\nname: "+skillName+"\n---\n"), 0644))
	return repo
}

func TestRemoveCmd_RemovesInstalledSkill(t *testing.T) {
	repo := makeRemoveRepo(t, "pdf-processing")
	_, err := executeCmd(t, "remove", "pdf-processing", "--repo", repo)
	require.NoError(t, err)
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "pdf-processing"))
	assert.True(t, os.IsNotExist(statErr), "skill directory should be removed")
}

func TestRemoveCmd_NotInstalled_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{Registries: map[string]string{}, Skills: map[string]manifest.SkillEntry{}}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	_, err := executeCmd(t, "remove", "nonexistent", "--repo", repo)
	assert.Error(t, err)
}

func TestRemoveCmd_DryRun_PrintsWouldRemove(t *testing.T) {
	repo := makeRemoveRepo(t, "pdf-processing")
	out, err := executeCmd(t, "remove", "pdf-processing", "--repo", repo, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "pdf-processing")
	// dry-run: directory still exists
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "pdf-processing"))
	assert.NoError(t, statErr, "dry-run should not delete files")
}
