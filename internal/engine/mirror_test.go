package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMirrors_InstallRemoveSyncDoctor(t *testing.T) {
	repo := t.TempDir()
	eng := newWithProvider(&fakeProvider{skill: &model.RegistrySkill{Name: "demo"}})
	makeManifestWithRegistry(t, repo, "default", "https://example.invalid/r.git")

	rep, err := eng.SetMirrors(repo, []string{"copilot", "cursor"}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"copilot", "cursor"}, rep.Targets)

	require.NoError(t, eng.Install(repo, "demo", "default", "", false))
	for _, dir := range []string{".claude", ".github", ".cursor"} {
		assert.FileExists(t, filepath.Join(repo, dir, "skills", "demo", "SKILL.md"), dir)
	}

	// Mirrors are not separate targets: the skill is listed once.
	skills, err := eng.ListFor(repo, "")
	require.NoError(t, err)
	assert.Len(t, skills, 1)

	// Doctor notices a stale mirror and sync repairs it.
	require.NoError(t, os.RemoveAll(filepath.Join(repo, ".github", "skills", "demo")))
	issues, err := eng.Doctor(repo)
	require.NoError(t, err)
	assert.True(t, hasIssue(issues, "stale-mirror"))
	_, err = eng.Sync(repo, true, false, false)
	require.Error(t, err)
	_, err = eng.Sync(repo, false, false, false)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(repo, ".github", "skills", "demo", "SKILL.md"))

	// Remove cleans up mirrors too.
	require.NoError(t, eng.Remove(repo, "demo", false))
	assert.NoDirExists(t, filepath.Join(repo, ".cursor", "skills", "demo"))

	// Dropping a mirror deletes what Skell placed there.
	require.NoError(t, eng.Install(repo, "demo", "default", "", false))
	_, err = eng.SetMirrors(repo, nil, []string{"cursor"})
	require.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(repo, ".cursor", "skills", "demo"))
	assert.DirExists(t, filepath.Join(repo, ".github", "skills", "demo"))
}

func TestSetMirrors_RejectsPrimaryAndUnknown(t *testing.T) {
	repo := t.TempDir()
	eng := newWithProvider(&fakeProvider{})
	makeManifestWithRegistry(t, repo, "default", "https://example.invalid/r.git")
	_, err := eng.SetMirrors(repo, []string{"claude"}, nil)
	assert.Error(t, err)
	_, err = eng.SetMirrors(repo, []string{"nope"}, nil)
	assert.Error(t, err)
}

func TestLink_LiveSkillIsKeptBySyncAndRemovable(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git", "info"), 0o755))
	eng := newWithProvider(&fakeProvider{})
	makeManifestWithRegistry(t, repo, "default", "https://example.invalid/r.git")

	src := filepath.Join(t.TempDir(), "my-skill")
	require.NoError(t, os.MkdirAll(src, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "SKILL.md"),
		[]byte("---\nname: my-skill\ndescription: d\n---\nv1\n"), 0o644))

	res, err := eng.Link(repo, src, "")
	require.NoError(t, err)
	assert.Equal(t, "my-skill", res.Name)

	// Edits in the source are visible through the link.
	require.NoError(t, os.WriteFile(filepath.Join(src, "SKILL.md"),
		[]byte("---\nname: my-skill\ndescription: d\n---\nv2\n"), 0o644))
	data, err := os.ReadFile(filepath.Join(repo, ".claude", "skills", "my-skill", "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "v2")

	lf, err := lockfile.Read(lockfile.PathFor(repo, target.MustLookup("claude")))
	require.NoError(t, err)
	require.NotNil(t, lf.FindSkill("my-skill"))
	assert.True(t, lf.FindSkill("my-skill").Linked)

	exclude, err := os.ReadFile(filepath.Join(repo, ".git", "info", "exclude"))
	require.NoError(t, err)
	assert.Contains(t, string(exclude), "/.claude/skills/my-skill")

	// Sync must not prune a linked skill even though it's not in the manifest.
	rep, err := eng.Sync(repo, false, false, false)
	require.NoError(t, err)
	assert.Empty(t, rep.Removed)

	st, err := eng.StatusFor(repo, "claude")
	require.NoError(t, err)
	require.Len(t, st, 1)
	assert.Equal(t, model.StatusLinked, st[0].Status)

	require.NoError(t, eng.Remove(repo, "my-skill", false))
	assert.FileExists(t, filepath.Join(src, "SKILL.md"), "removing the link must not touch the source")
}

func hasIssue(issues []DiagnosticIssue, code string) bool {
	for _, i := range issues {
		if i.Code == code {
			return true
		}
	}
	return false
}

func TestReviewSkill_FlagsScriptsAndBroadTools(t *testing.T) {
	f := newGitRegistryFixture(t)
	dir := filepath.Join(f.upstream, "skills", "demo")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "scripts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts", "run.sh"), []byte("echo hi\n"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: demo\ndescription: d\nallowed-tools: Bash Read\n---\nSee https://example.com/docs.\n"), 0o644))
	git(t, f.upstream, "add", ".")
	git(t, f.upstream, "commit", "-qm", "scripts")
	git(t, f.cacheClone, "fetch", "-q", "origin")
	git(t, f.cacheClone, "reset", "-q", "--hard", "FETCH_HEAD")

	rev, err := f.eng.ReviewSkill(f.repo, "demo", "demo", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"SKILL.md", "scripts/run.sh"}, rev.Files)
	assert.Equal(t, []string{"scripts/run.sh"}, rev.Scripts)
	assert.Equal(t, []string{"https://example.com/docs"}, rev.Links)
	assert.True(t, rev.NeedsConfirmation())
	assert.Len(t, rev.Warnings, 2)
}

func TestInstall_FindsRegistryAutomatically(t *testing.T) {
	repo := t.TempDir()
	eng := newWithProvider(&fakeProvider{skill: &model.RegistrySkill{Name: "pdf"}})
	makeManifestWithRegistry(t, repo, "anthropic", "https://example.invalid/a.git")
	require.NoError(t, eng.Install(repo, "pdf", "", "", false))
	lf, err := lockfile.Read(lockfile.PathFor(repo, target.MustLookup("claude")))
	require.NoError(t, err)
	assert.Equal(t, "anthropic", lf.FindSkill("pdf").Registry)

	// Two sources both providing the skill → the user must choose.
	repo2 := t.TempDir()
	makeManifestWithRegistry(t, repo2, "a", "https://example.invalid/a.git")
	mp := manifest.LocalPath(repo2)
	m, err := manifest.Read(mp)
	require.NoError(t, err)
	m.Registries["b"] = "https://example.invalid/b.git"
	require.NoError(t, manifest.Write(mp, m))
	err = eng.Install(repo2, "pdf", "", "", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "several sources")

	// No sources at all → point the user at 'skell add'.
	repo3 := t.TempDir()
	require.NoError(t, eng.InitFor(repo3, target.MustLookup("claude")))
	err = eng.Install(repo3, "pdf", "", "", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skell add")
}

func TestApplyBaseline_AddsMissingKeepsOwnRejectsRedirect(t *testing.T) {
	repo := t.TempDir()
	eng := newWithProvider(&fakeProvider{})
	makeManifestWithRegistry(t, repo, "team", "https://example.invalid/team.git")

	base := &manifest.Manifest{
		Mirrors:    []string{"copilot", "claude"},
		Registries: map[string]string{"team": "https://example.invalid/team.git", "extra": "https://example.invalid/extra.git"},
		Skills:     map[string]manifest.SkillEntry{"pdf": {Registry: "extra"}},
	}
	rep, err := eng.ApplyBaseline(repo, base, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"extra"}, rep.AddedRegistries)
	assert.Equal(t, []string{"pdf"}, rep.AddedSkills)
	assert.Equal(t, []string{"copilot"}, rep.AddedMirrors)

	base.Registries["team"] = "https://evil.invalid/team.git"
	_, err = eng.ApplyBaseline(repo, base, false)
	assert.Error(t, err)
}
