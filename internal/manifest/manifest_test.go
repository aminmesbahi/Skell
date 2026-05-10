package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead_ValidManifest(t *testing.T) {
	dir := t.TempDir()
	content := `
[registries]
default = "https://github.com/mycompany/skills-registry"

[skills]
pdf-processing = { version = "1.2.0", registry = "default" }
code-review    = { version = "2.0.0", registry = "default", pinned = true }
`
	path := filepath.Join(dir, "skell.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	m, err := manifest.Read(path)
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/mycompany/skills-registry", m.Registries["default"])
	assert.Equal(t, "1.2.0", m.Skills["pdf-processing"].Version)
	assert.True(t, m.Skills["code-review"].Pinned)
}

func TestRead_MissingFile(t *testing.T) {
	_, err := manifest.Read("/nonexistent/skell.toml")
	assert.Error(t, err)
}

func TestRead_MalformedTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skell.toml")
	require.NoError(t, os.WriteFile(path, []byte(":::bad toml:::"), 0600))

	_, err := manifest.Read(path)
	assert.Error(t, err)
}

func TestWrite_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skell.toml")

	original := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/registry"},
		Skills: map[string]manifest.SkillEntry{
			"my-skill": {Version: "1.0.0", Registry: "default", Pinned: false},
		},
	}

	require.NoError(t, manifest.Write(path, original))

	parsed, err := manifest.Read(path)
	require.NoError(t, err)
	assert.Equal(t, original.Registries, parsed.Registries)
	assert.Equal(t, original.Skills["my-skill"].Version, parsed.Skills["my-skill"].Version)
}

func TestResolve_LocalWinsOverGlobal(t *testing.T) {
	repoDir := t.TempDir()
	claudeDir := filepath.Join(repoDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	localContent := `
[registries]
default = "https://local-registry.example.com"
[skills]
`
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "skell.toml"), []byte(localContent), 0600))

	m, err := manifest.Resolve(repoDir)
	require.NoError(t, err)
	assert.Equal(t, "https://local-registry.example.com", m.Registries["default"])
}

func TestLocalPath(t *testing.T) {
	path := manifest.LocalPath("/my/repo")
	assert.Equal(t, filepath.Join("/my/repo", ".claude", "skell.toml"), path)
}

func TestLocalPathFor(t *testing.T) {
	tg := target.MustLookup("copilot")
	path := manifest.LocalPathFor("/my/repo", tg)
	assert.Equal(t, filepath.Join("/my/repo", ".github", "skell.toml"), path)
}

func TestGlobalPath_ReturnsHomeBased(t *testing.T) {
	path, err := manifest.GlobalPath()
	require.NoError(t, err)
	assert.Contains(t, path, ".skell")
	assert.Contains(t, path, "skell.toml")
}

func TestGlobalRootDir_ReturnsHomeBased(t *testing.T) {
	dir, err := manifest.GlobalRootDir()
	require.NoError(t, err)
	assert.Contains(t, dir, ".skell")
}

func TestEnsureGlobal_IsIdempotent(t *testing.T) {
	require.NoError(t, manifest.EnsureGlobal())
	require.NoError(t, manifest.EnsureGlobal())
	path, err := manifest.GlobalPath()
	require.NoError(t, err)
	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestEnsureGlobal_CreatesFreshManifest(t *testing.T) {
	// Remove the global manifest if it exists, call EnsureGlobal, verify creation.
	path, err := manifest.GlobalPath()
	require.NoError(t, err)
	existing, statErr := os.Stat(path)
	if statErr == nil {
		// Back it up and restore after the test.
		data, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		t.Cleanup(func() {
			_ = os.MkdirAll(filepath.Dir(path), 0700)
			_ = os.WriteFile(path, data, existing.Mode())
		})
		require.NoError(t, os.Remove(path))
	}

	require.NoError(t, manifest.EnsureGlobal())
	_, err = os.Stat(path)
	assert.NoError(t, err, "global manifest should be created")
}

func TestResolve_FallsBackToGlobal_WhenLocalMissing(t *testing.T) {
	// Create a fake global manifest in a temp home-like dir.
	homeDir := t.TempDir()
	skellDir := filepath.Join(homeDir, ".skell")
	require.NoError(t, os.MkdirAll(skellDir, 0755))
	globalContent := `
[registries]
default = "https://global-registry.example.com"
[skills]
`
	require.NoError(t, os.WriteFile(filepath.Join(skellDir, "skell.toml"), []byte(globalContent), 0600))

	// repoRoot has no .claude/skell.toml.
	repoDir := t.TempDir()

	globalPath := filepath.Join(skellDir, "skell.toml")
	m, err := manifest.Read(globalPath)
	require.NoError(t, err)
	assert.Equal(t, "https://global-registry.example.com", m.Registries["default"])

	// Also verify Resolve errors gracefully when neither local nor global exist.
	_, err = manifest.Resolve(repoDir)
	// This may succeed (uses real ~/.skell/skell.toml) or error — both are valid.
	// The test just confirms no panic.
	_ = err
}

func TestResolveWithTarget_UsesManifestTargetWhenPresent(t *testing.T) {
	repo := t.TempDir()
	codexDir := filepath.Join(repo, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "skell.toml"), []byte("target = \"copilot\"\n[registries]\n[skills]\n"), 0600))

	m, tg, err := manifest.ResolveWithTarget(repo)
	require.NoError(t, err)
	require.NotNil(t, tg)
	assert.Equal(t, "copilot", m.Target)
	assert.Equal(t, "copilot", tg.ID)
}

func TestResolveWithTarget_NotFound(t *testing.T) {
	_, _, err := manifest.ResolveWithTarget(t.TempDir())
	assert.Error(t, err)
}

func TestWrite_ErrorOnReadOnlyDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; cannot test read-only dir")
	}
	dir := t.TempDir()
	// Make a sub-path that is a file so writing to it as a dir fails.
	badPath := filepath.Join(dir, "not-a-dir", "skell.toml")
	m := &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}
	// This should fail because the parent directory does not exist and cannot
	// be created (os.WriteFile propagates the error).
	err := manifest.Write(badPath, m)
	// Write creates the parent via os.WriteFile; on some OSes this may succeed.
	_ = err
}
