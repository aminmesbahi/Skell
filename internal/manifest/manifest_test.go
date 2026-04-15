package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
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
