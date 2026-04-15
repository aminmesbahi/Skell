package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeGitRepo(t *testing.T, root string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0755))
}

func makeSkill(t *testing.T, repoRoot, skillName string) {
	t.Helper()
	dir := filepath.Join(repoRoot, ".claude", "skills", skillName)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+skillName+"\n---\n"), 0600))
}

func TestIsGitRepo_True(t *testing.T) {
	dir := t.TempDir()
	makeGitRepo(t, dir)
	assert.True(t, scanner.IsGitRepo(dir))
}

func TestIsGitRepo_False(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, scanner.IsGitRepo(dir))
}

func TestSkillsDir(t *testing.T) {
	path := scanner.SkillsDir("/my/repo")
	assert.Equal(t, filepath.Join("/my/repo", ".claude", "skills"), path)
}

func TestScanRepo_NoSkills(t *testing.T) {
	dir := t.TempDir()
	makeGitRepo(t, dir)

	result, err := scanner.ScanRepo(dir)
	require.NoError(t, err)
	assert.Empty(t, result.InstalledSkills)
}

func TestScanRepo_DetectsInstalledSkills(t *testing.T) {
	dir := t.TempDir()
	makeGitRepo(t, dir)
	makeSkill(t, dir, "pdf-processing")
	makeSkill(t, dir, "code-review")

	result, err := scanner.ScanRepo(dir)
	require.NoError(t, err)
	assert.Len(t, result.InstalledSkills, 2)
}

func TestScanRepo_DetectsManifestPresence(t *testing.T) {
	dir := t.TempDir()
	makeGitRepo(t, dir)
	claudeDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "skell.toml"), []byte("[skills]\n"), 0600))

	result, err := scanner.ScanRepo(dir)
	require.NoError(t, err)
	assert.True(t, result.HasManifest)
	assert.False(t, result.HasLockFile)
}

func TestScanAll_FindsMultipleRepos(t *testing.T) {
	root := t.TempDir()
	repo1 := filepath.Join(root, "proj-a")
	repo2 := filepath.Join(root, "proj-b")
	makeGitRepo(t, repo1)
	makeGitRepo(t, repo2)

	results, err := scanner.ScanAll(root)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestScanAll_IgnoresNonGitDirs(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "not-a-repo"), 0755))
	makeGitRepo(t, filepath.Join(root, "real-repo"))

	results, err := scanner.ScanAll(root)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}
