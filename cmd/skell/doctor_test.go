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

func makeDoctorCmdRepo(t *testing.T, withLock bool, skills []string) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	skillEntries := make(map[string]manifest.SkillEntry)
	var locked []model.InstalledSkill
	for _, name := range skills {
		skillEntries[name] = manifest.SkillEntry{Version: "1.0.0", Registry: "default"}
		locked = append(locked, model.InstalledSkill{Name: name, Version: "1.0.0", Registry: "default"})
		skillDir := filepath.Join(claudeDir, "skills", name)
		require.NoError(t, os.MkdirAll(skillDir, 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(skillDir, "SKILL.md"),
			[]byte("---\nname: "+name+"\n---\n"),
			0644,
		))
	}
	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     skillEntries,
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	if withLock {
		lf := &lockfile.LockFile{SkellVersion: "0.1.0", Skills: locked}
		require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	}
	return repo
}

func TestDoctorCmd_CleanRepo_PrintsOk(t *testing.T) {
	repo := makeDoctorCmdRepo(t, true, []string{"pdf"})
	out, err := executeCmd(t, "doctor", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "ok")
}

func TestDoctorCmd_NoManifest_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "doctor", "--repo", repo)
	assert.Error(t, err)
}

func TestDoctorCmd_UntrackedSkill_ReturnsError(t *testing.T) {
	repo := makeDoctorCmdRepo(t, true, []string{"pdf"})
	// Add untracked skill
	extraDir := filepath.Join(repo, ".claude", "skills", "extra")
	require.NoError(t, os.MkdirAll(extraDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(extraDir, "SKILL.md"), []byte("---\nname: extra\n---\n"), 0644))

	out, err := executeCmd(t, "doctor", "--repo", repo)
	assert.Error(t, err)
	assert.Contains(t, out, "extra")
}
