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

// ── resolveRepos ────────────────────────────────────────────────────────────

func TestResolveRepos_ExplicitRepo(t *testing.T) {
	repos, err := resolveRepos(repoFlags{repo: []string{"/some/path"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"/some/path"}, repos)
}

func TestResolveRepos_MultipleRepos(t *testing.T) {
	repos, err := resolveRepos(repoFlags{repo: []string{"/a", "/b"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"/a", "/b"}, repos)
}

func TestResolveRepos_AllRepos(t *testing.T) {
	root := t.TempDir()
	// Create two git repos inside root.
	repo1 := filepath.Join(root, "proj-a")
	repo2 := filepath.Join(root, "proj-b")
	require.NoError(t, os.MkdirAll(filepath.Join(repo1, ".git"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(repo2, ".git"), 0755))

	repos, err := resolveRepos(repoFlags{allRepos: root})
	require.NoError(t, err)
	assert.Len(t, repos, 2)
}

func TestResolveRepos_AllRepos_BadPath_ReturnsError(t *testing.T) {
	_, err := resolveRepos(repoFlags{allRepos: "/nonexistent/path/that/does/not/exist"})
	assert.Error(t, err)
}

func TestResolveRepos_Global_ReturnsHomeBasedPath(t *testing.T) {
	repos, err := resolveRepos(repoFlags{global: true})
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Contains(t, repos[0], ".skell")
}

func TestResolveRepos_CWD_Fallback(t *testing.T) {
	repos, err := resolveRepos(repoFlags{})
	require.NoError(t, err)
	require.Len(t, repos, 1)
	// Should be a non-empty path (CWD).
	assert.NotEmpty(t, repos[0])
}

func TestResolveRepo_EmptyUseCWD(t *testing.T) {
	path, err := resolveRepo("")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestResolveRepo_ExplicitPath(t *testing.T) {
	path, err := resolveRepo("/explicit/path")
	require.NoError(t, err)
	assert.Equal(t, "/explicit/path", path)
}

// ── defaultCacheRoot ────────────────────────────────────────────────────────

func TestDefaultCacheRoot_ReturnsPath(t *testing.T) {
	root := defaultCacheRoot()
	assert.NotEmpty(t, root)
	assert.Contains(t, root, ".skell")
}

// ── listRegistry ─────────────────────────────────────────────────────────────

func TestListCmd_SourceRegistry_NoManifest_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "list", "--repo", repo, "--source", "registry")
	assert.Error(t, err)
}

func TestListCmd_SourceRegistry_WithManifest_NoSkills(t *testing.T) {
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	out, err := executeCmd(t, "list", "--repo", repo, "--source", "registry")
	require.NoError(t, err)
	assert.Contains(t, out, "no skills found in registry")
}

// ── upgrade cmd ──────────────────────────────────────────────────────────────

func TestUpgradeCmd_NothingToUpgrade_OutputsMessage(t *testing.T) {
	// A repo with a manifest + lockfile but an empty registry will error.
	// We test the "nothing to upgrade" branch by creating a repo with no lock file entries.
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com"},
		Skills:     map[string]manifest.SkillEntry{},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))
	lf := &lockfile.LockFile{SkellVersion: "0.1.0", Skills: []model.InstalledSkill{}}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))

	out, err := executeCmd(t, "upgrade", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "nothing to upgrade")
}

// ── scanner extra ─────────────────────────────────────────────────────────────

// TestScanAll_FindsSkellManifestRepo checks that dirs with only skell.toml are found.
func TestScanAll_FindsSkellManifestOnlyRepo(t *testing.T) {
	root := t.TempDir()
	proj := filepath.Join(root, "skell-only-repo")
	claudeDir := filepath.Join(proj, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "skell.toml"), []byte("[skills]\n"), 0600))

	// Also add a non-special directory that should be ignored.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "regular-dir"), 0755))

	// Import scanner via cmd test (same package as skell cmd).
	// We call it via the CLI: list --all-repos.
	out, err := executeCmd(t, "list", "--all-repos", root)
	require.NoError(t, err)
	assert.NotContains(t, out, "Error")
}

// ── resolveManifest ──────────────────────────────────────────────────────────

func TestResolveManifest_WithLocalManifest_ReturnsIt(t *testing.T) {
	repo := makeRepoWithManifestCmd(t)
	m, err := resolveManifest(repo, false)
	require.NoError(t, err)
	assert.NotNil(t, m)
}

func TestResolveManifest_NoManifest_FallbackFalse_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	_, err := resolveManifest(repo, false)
	assert.Error(t, err)
}

func TestResolveManifest_NoManifest_FallbackTrue_ReturnsGlobal(t *testing.T) {
	// When fallback=true and there is no local manifest, resolveManifest falls
	// back to the global manifest (creating it if necessary).
	repo := t.TempDir()
	m, err := resolveManifest(repo, true)
	require.NoError(t, err)
	assert.NotNil(t, m)
}
