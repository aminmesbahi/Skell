package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeRepo creates a minimal fake repository layout in a temp dir.
func makeRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0755))
	return dir
}

// makeInstalledSkill creates a .claude/skills/<name>/SKILL.md inside repoRoot.
func makeInstalledSkill(t *testing.T, repoRoot, name, content string) {
	t.Helper()
	skillDir := filepath.Join(repoRoot, ".claude", "skills", name)
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0600))
}

func TestInit_CreatesManifestFromInstalledSkills(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "pdf-processing", "---\nname: pdf-processing\ndescription: PDF tool.\nmetadata:\n  version: \"1.2.0\"\n---\n")
	makeInstalledSkill(t, repo, "code-review", "---\nname: code-review\ndescription: Review tool.\nmetadata:\n  version: \"2.0.0\"\n---\n")

	eng := newWithProvider(nil)
	require.NoError(t, eng.Init(repo))

	manifestPath := manifest.LocalPath(repo)
	m, err := manifest.Read(manifestPath)
	require.NoError(t, err)

	assert.Len(t, m.Skills, 2)
	assert.Equal(t, "1.2.0", m.Skills["pdf-processing"].Version)
	assert.Empty(t, m.Skills["pdf-processing"].Registry, "Init must not stamp a default registry alias")
	assert.Equal(t, "2.0.0", m.Skills["code-review"].Version)
}

func TestInit_EmptyRepo_CreatesEmptyManifest(t *testing.T) {
	repo := makeRepo(t)

	eng := newWithProvider(nil)
	require.NoError(t, eng.Init(repo))

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Empty(t, m.Skills)
}

func TestInit_SkillWithNoFrontmatter_EmptyVersion(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "bare-skill", "# No frontmatter at all")

	eng := newWithProvider(nil)
	require.NoError(t, eng.Init(repo))

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)

	entry, ok := m.Skills["bare-skill"]
	require.True(t, ok, "bare-skill should appear in manifest")
	assert.Empty(t, entry.Version, "version should be empty when frontmatter is absent")
	assert.Empty(t, entry.Registry)
}

func TestInit_FailsIfManifestAlreadyExists(t *testing.T) {
	repo := makeRepo(t)
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "skell.toml"), []byte("[skills]\n"), 0600))

	eng := newWithProvider(nil)
	err := eng.Init(repo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "skell.toml already exists")
}

func TestInit_CreatesClaudeDirIfAbsent(t *testing.T) {
	repo := makeRepo(t)
	// .claude does not exist yet
	eng := newWithProvider(nil)
	require.NoError(t, eng.Init(repo))

	_, err := os.Stat(filepath.Join(repo, ".claude"))
	assert.NoError(t, err, ".claude directory should be created")
}
