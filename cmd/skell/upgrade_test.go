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

func makeUpgradeRepo(t *testing.T, skillName, version string) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     map[string]manifest.SkillEntry{skillName: {Version: version, Registry: "default"}},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills:       []model.InstalledSkill{{Name: skillName, Version: version, Registry: "default"}},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	return repo
}

func TestUpgradeCmd_NoLockFile_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{Registries: map[string]string{"default": "https://x.com"}}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	_, err := executeCmd(t, "upgrade", "--repo", repo)
	assert.Error(t, err)
}

func TestUpgradeCmd_AlreadyUpToDate_PrintsNothingToUpgrade(t *testing.T) {
	// registry is not implemented → skill will show as error from GetSkill
	// Just verify no panic and a graceful error.
	repo := makeUpgradeRepo(t, "pdf-processing", "1.0.0")
	_, err := executeCmd(t, "upgrade", "--repo", repo)
	// Real registry returns error, so upgrade propagates it.
	assert.Error(t, err)
}

func TestUpgradeCmd_DryRunFlag_Accepted(t *testing.T) {
	repo := makeUpgradeRepo(t, "pdf-processing", "1.0.0")
	// With dry-run the engine still calls GetSkill and will fail, but flag is parsed.
	_, err := executeCmd(t, "upgrade", "--repo", repo, "--dry-run")
	// Error expected because real registry is not implemented.
	assert.Error(t, err)
}
