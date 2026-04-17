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

func makeLockFile(t *testing.T, repo string, skills ...model.InstalledSkill) {
	t.Helper()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	lf := &lockfile.LockFile{SkellVersion: "0.1.0", Skills: skills}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
}

func TestListCmd_PrintsInstalledSkills(t *testing.T) {
	repo := t.TempDir()
	makeLockFile(t, repo,
		model.InstalledSkill{Name: "pdf-processing", Version: "1.2.0", Registry: "default"},
		model.InstalledSkill{Name: "code-review", Version: "2.0.0", Registry: "default"},
	)

	out, err := executeCmd(t, "list", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "pdf-processing")
	assert.Contains(t, out, "code-review")
}

func TestListCmd_EmptyRepo_PrintsNoSkills(t *testing.T) {
	repo := t.TempDir()

	out, err := executeCmd(t, "list", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "no skills installed")
}

func TestListCmd_PinnedSkill_ShowsPinnedLabel(t *testing.T) {
	repo := t.TempDir()
	makeLockFile(t, repo,
		model.InstalledSkill{Name: "code-review", Version: "2.0.0", Registry: "default", Pinned: true},
	)

	out, err := executeCmd(t, "list", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "[pinned]")
}
