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

func makeStatusRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     map[string]manifest.SkillEntry{},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	lf := &lockfile.LockFile{
		SkellVersion: "0.1.0",
		Skills: []model.InstalledSkill{
			{Name: "pdf-processing", Version: "1.0.0", Registry: "default"},
		},
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	return repo
}

func TestStatusCmd_PrintsTable(t *testing.T) {
	repo := makeStatusRepo(t)
	out, err := executeCmd(t, "status", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "pdf-processing")
	assert.Contains(t, out, "1.0.0")
}

func TestStatusCmd_OnlyFilter_HidesNonMatching(t *testing.T) {
	repo := makeStatusRepo(t)
	out, err := executeCmd(t, "status", "--repo", repo, "--only", "outdated")
	require.NoError(t, err)
	// skill is unknown (registry not implemented), not outdated
	assert.NotContains(t, out, "pdf-processing")
}

func TestStatusCmd_NoLockFile_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{Registries: map[string]string{}}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	_, err := executeCmd(t, "status", "--repo", repo)
	assert.Error(t, err)
}
