package registry

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

func writeSkill(t *testing.T, root, rel, body string) {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(dir, 0o755))
	content := "---\nname: " + filepath.Base(dir) + "\ndescription: test\n---\n" + body + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644))
}

// newUpstream creates a git repo with two commits of skills/demo and returns
// the repo path plus both commit SHAs.
func newUpstream(t *testing.T) (string, string, string) {
	t.Helper()
	up := t.TempDir()
	gitRun(t, up, "init", "-q")
	gitRun(t, up, "config", "user.email", "t@t")
	gitRun(t, up, "config", "user.name", "t")
	gitRun(t, up, "config", "core.autocrlf", "false")
	writeSkill(t, up, "skills/demo", "v1")
	gitRun(t, up, "add", ".")
	gitRun(t, up, "commit", "-qm", "v1")
	c1 := gitRun(t, up, "rev-parse", "HEAD")
	writeSkill(t, up, "skills/demo", "v2")
	gitRun(t, up, "commit", "-qam", "v2")
	c2 := gitRun(t, up, "rev-parse", "HEAD")
	return up, c1, c2
}

// remoteReg returns a registry whose URL looks remote but whose cache was
// pre-populated by cloning the local upstream (so no network is used).
func remoteReg(t *testing.T, up string, shallow bool) (*Adapter, Registry) {
	t.Helper()
	cache := t.TempDir()
	args := []string{"clone", "-q"}
	if shallow {
		args = append(args, "--depth=1")
	}
	args = append(args, "file://"+filepath.ToSlash(up), filepath.Join(cache, "demo-reg"))
	gitRun(t, cache, args...)
	return NewAdapter(cache), Registry{Alias: "demo-reg", URL: "https://example.invalid/demo-reg.git"}
}

func TestSource_ReportsRelPathAndCommit(t *testing.T) {
	up, _, c2 := newUpstream(t)
	a, reg := remoteReg(t, up, false)
	src, err := a.Source(reg, "demo")
	require.NoError(t, err)
	assert.Equal(t, "skills/demo", src.RelPath)
	assert.Equal(t, c2, src.Commit)
}

func TestCopySkillAt_InstallsOlderCommit(t *testing.T) {
	up, c1, _ := newUpstream(t)
	for _, shallow := range []bool{false, true} {
		a, reg := remoteReg(t, up, shallow)
		dest := filepath.Join(t.TempDir(), "skills", "demo")
		require.NoError(t, a.CopySkillAt(reg, "demo", c1, "skills/demo", dest), "shallow=%v", shallow)
		data, err := os.ReadFile(filepath.Join(dest, "SKILL.md"))
		require.NoError(t, err)
		assert.Contains(t, string(data), "v1", "shallow=%v", shallow)
	}
}

func TestCopySkillAt_RejectsBadInput(t *testing.T) {
	up, _, _ := newUpstream(t)
	a, reg := remoteReg(t, up, false)
	dest := filepath.Join(t.TempDir(), "demo")
	assert.Error(t, a.CopySkillAt(reg, "demo", "--upload-pack=evil", "", dest))
	assert.Error(t, a.CopySkillAt(reg, "demo", "0123456789abcdef0123456789abcdef01234567", "../escape", dest))
}

func TestDiffDirs(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(a, "SKILL.md"), []byte("one\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(b, "SKILL.md"), []byte("two\n"), 0o644))
	out, err := DiffDirs(a, b, false)
	require.NoError(t, err)
	assert.Contains(t, out, "-one")
	assert.Contains(t, out, "+two")

	same, err := DiffDirs(a, a, false)
	require.NoError(t, err)
	assert.Empty(t, same)
}
