package skell

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyCmd_ReviewsSkillsTheBaselineAdds(t *testing.T) {
	src := t.TempDir()
	dir := filepath.Join(src, "skills", "demo")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "scripts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: demo\ndescription: demo skill. Use when testing.\n---\nv1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts", "run.sh"), []byte("echo hi\n"), 0o755))

	baseline := filepath.Join(t.TempDir(), "baseline.toml")
	require.NoError(t, os.WriteFile(baseline, []byte(fmt.Sprintf(
		"target = \"claude\"\n[registries]\nsrc = %q\n[skills.demo]\nregistry = \"src\"\n", filepath.ToSlash(src))), 0o644))

	repo := t.TempDir()
	_, err := executeCmd(t, "init", "--repo", repo, "--target", "claude", "--yes")
	require.NoError(t, err)

	// Without a terminal a skill that ships scripts is refused, and nothing changes.
	out, err := executeCmd(t, "apply", baseline, "--repo", repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--yes")
	assert.Contains(t, out, "scripts/run.sh")
	m, merr := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, merr)
	assert.Empty(t, m.Skills)
	assert.Empty(t, m.Registries)
	assert.NoDirExists(t, filepath.Join(repo, ".claude", "skills", "demo"))

	// --yes approves it.
	_, err = executeCmd(t, "apply", baseline, "--repo", repo, "--yes")
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(repo, ".claude", "skills", "demo", "scripts", "run.sh"))
}
