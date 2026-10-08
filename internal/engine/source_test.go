package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

func writeDemoSkill(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, "skills", "demo")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: demo\ndescription: demo skill\n---\n"+body+"\n"), 0o644))
}

// gitRegistryFixture builds an upstream repo, a pre-populated cache clone (so
// the "remote" registry needs no network) and an initialised project repo.
type gitRegistryFixture struct {
	upstream, cacheClone, repo string
	eng                        *Engine
}

func newGitRegistryFixture(t *testing.T) *gitRegistryFixture {
	t.Helper()
	home := t.TempDir()
	up := t.TempDir()
	git(t, up, "init", "-q")
	git(t, up, "config", "user.email", "t@t")
	git(t, up, "config", "user.name", "t")
	git(t, up, "config", "core.autocrlf", "false")
	writeDemoSkill(t, up, "v1")
	git(t, up, "add", ".")
	git(t, up, "commit", "-qm", "v1")

	cacheRoot := filepath.Join(home, "cache")
	require.NoError(t, os.MkdirAll(cacheRoot, 0o755))
	clone := filepath.Join(cacheRoot, "demo")
	git(t, cacheRoot, "clone", "-q", "--config", "core.autocrlf=false", "file://"+filepath.ToSlash(up), clone)

	repo := t.TempDir()
	eng := New(cacheRoot)
	require.NoError(t, eng.InitFor(repo, target.MustLookup("claude")))
	mp := manifest.LocalPathFor(repo, target.MustLookup("claude"))
	m, err := manifest.Read(mp)
	require.NoError(t, err)
	m.Registries["demo"] = "https://example.invalid/demo.git"
	require.NoError(t, manifest.Write(mp, m))
	return &gitRegistryFixture{upstream: up, cacheClone: clone, repo: repo, eng: eng}
}

// publish commits a new upstream version and refreshes the cache clone.
func (f *gitRegistryFixture) publish(t *testing.T, body string) string {
	t.Helper()
	writeDemoSkill(t, f.upstream, body)
	git(t, f.upstream, "commit", "-qam", body)
	git(t, f.cacheClone, "fetch", "-q", "origin")
	git(t, f.cacheClone, "reset", "-q", "--hard", "FETCH_HEAD")
	return git(t, f.upstream, "rev-parse", "HEAD")
}

func (f *gitRegistryFixture) lockEntry(t *testing.T) model.InstalledSkill {
	t.Helper()
	lf, err := lockfile.Read(lockfile.PathFor(f.repo, target.MustLookup("claude")))
	require.NoError(t, err)
	e := lf.FindSkill("demo")
	require.NotNil(t, e)
	return *e
}

func (f *gitRegistryFixture) installedBody(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(f.repo, ".claude", "skills", "demo", "SKILL.md"))
	require.NoError(t, err)
	return string(data)
}

func TestGitRegistry_CommitLockStatusDiffSyncUpgrade(t *testing.T) {
	f := newGitRegistryFixture(t)
	c1 := git(t, f.upstream, "rev-parse", "HEAD")

	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	locked := f.lockEntry(t)
	assert.Equal(t, c1, locked.Commit)
	assert.Equal(t, "skills/demo", locked.SourcePath)

	// Unversioned but unchanged content → up-to-date (not "unversioned").
	st, err := f.eng.StatusFor(f.repo, "claude")
	require.NoError(t, err)
	require.Len(t, st, 1)
	assert.Equal(t, model.StatusUpToDate, st[0].Status)

	// Upstream edits without a version bump are detected.
	c2 := f.publish(t, "v2")
	st, err = f.eng.StatusFor(f.repo, "claude")
	require.NoError(t, err)
	assert.Equal(t, model.StatusOutdated, st[0].Status)
	assert.True(t, st[0].Changed)

	d, err := f.eng.Diff(f.repo, "demo", "", false)
	require.NoError(t, err)
	assert.Contains(t, d.Patch, "-v1")
	assert.Contains(t, d.Patch, "+v2")
	assert.Equal(t, c2, d.LatestCommit)

	// Sync restores the *locked* revision, not the newest one.
	require.NoError(t, os.RemoveAll(filepath.Join(f.repo, ".claude", "skills", "demo")))
	rep, err := f.eng.Sync(f.repo, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"demo"}, rep.Installed)
	assert.Contains(t, f.installedBody(t), "v1")

	// Upgrade moves to the new commit.
	report, err := f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	require.Len(t, report.Upgraded, 1)
	assert.Contains(t, f.installedBody(t), "v2")
	assert.Equal(t, c2, f.lockEntry(t).Commit)

	// Nothing changed upstream → upgrade is a no-op.
	report, err = f.eng.UpgradeFor(f.repo, "demo", "", false, false)
	require.NoError(t, err)
	assert.Empty(t, report.Upgraded)
}

func TestUpgrade_PreservesPinAndSourceRepo(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))

	lp := lockfile.PathFor(f.repo, target.MustLookup("claude"))
	lf, err := lockfile.Read(lp)
	require.NoError(t, err)
	lf.Skills[0].Pinned = true
	lf.Skills[0].SourceRepo = "https://github.com/x/y/tree/main/skills/demo"
	require.NoError(t, lockfile.Write(lp, lf))

	f.publish(t, "v2")
	_, err = f.eng.UpgradeFor(f.repo, "demo", "", true, false)
	require.NoError(t, err)
	e := f.lockEntry(t)
	assert.True(t, e.Pinned, "forced upgrade must keep the pin")
	assert.Equal(t, "https://github.com/x/y/tree/main/skills/demo", e.SourceRepo)
}

func TestExpandAddArg(t *testing.T) {
	lookup := func(id string) (string, bool) {
		if id == "anthropic" {
			return "https://github.com/anthropics/skills", true
		}
		return "", false
	}
	cases := map[string]string{
		"anthropic":              "https://github.com/anthropics/skills",
		"owner/repo":             "https://github.com/owner/repo",
		"owner/repo.git":         "https://github.com/owner/repo",
		"owner/repo/skills/pdf":  "https://github.com/owner/repo/tree/HEAD/skills/pdf",
		"https://gitlab.com/a/b": "https://gitlab.com/a/b",
		"git@github.com:a/b.git": "git@github.com:a/b.git",
		"not-a-catalog-id":       "not-a-catalog-id",
	}
	for in, want := range cases {
		got, _ := ExpandAddArg(in, lookup)
		assert.Equal(t, want, got, in)
	}
	_, isCatalog := ExpandAddArg("anthropic", lookup)
	assert.True(t, isCatalog)

	dir := t.TempDir()
	got, _ := ExpandAddArg(dir, lookup)
	assert.Equal(t, dir, got)
}

func TestAddFromURL_DisambiguatesAlias(t *testing.T) {
	repo := t.TempDir()
	eng := newWithProvider(&fakeProvider{})
	require.NoError(t, eng.InitFor(repo, target.MustLookup("claude")))

	r1, err := eng.AddFromURL(repo, "https://github.com/anthropics/skills", false)
	require.NoError(t, err)
	assert.Equal(t, "skills", r1.Alias)
	r2, err := eng.AddFromURL(repo, "https://github.com/openai/skills", false)
	require.NoError(t, err)
	assert.Equal(t, "openai-skills", r2.Alias)

	r3, err := eng.AddFromURLWith(repo, "https://github.com/openai/skills", AddOptions{Alias: "openai"})
	require.NoError(t, err)
	assert.Equal(t, "openai", r3.Alias)
}
