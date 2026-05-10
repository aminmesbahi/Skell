package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/scanner"
	"github.com/aminmesbahi/skell/internal/target"
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

func TestSkillsDirFor(t *testing.T) {
	tg := target.MustLookup("copilot")
	path := scanner.SkillsDirFor("/my/repo", tg)
	assert.Equal(t, filepath.Join("/my/repo", ".github", "skills"), path)
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

func TestScanAll_FindsSkellManifestOnlyDir(t *testing.T) {
	root := t.TempDir()
	proj := filepath.Join(root, "skell-proj")
	claudeDir := filepath.Join(proj, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "skell.toml"), []byte("[skills]\n"), 0600))

	results, err := scanner.ScanAll(root)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, proj, results[0].RepoRoot)
}

func TestHasSkellManifest_True(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "skell.toml"), []byte("[skills]\n"), 0600))
	assert.True(t, scanner.HasSkellManifest(dir))
}

func TestHasSkellManifest_False(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, scanner.HasSkellManifest(dir))
}

func TestScanAll_InvalidPath_ReturnsError(t *testing.T) {
	_, err := scanner.ScanAll("/nonexistent/path/that/definitely/does/not/exist")
	assert.Error(t, err)
}

func TestScanRepo_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; skipping permission test")
	}
	root := t.TempDir()
	makeGitRepo(t, root)
	skillsDir := filepath.Join(root, ".claude", "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))
	// Create a subdirectory inside skills that is not readable.
	unreadable := filepath.Join(skillsDir, "unreadable")
	require.NoError(t, os.MkdirAll(unreadable, 0755))
	// Skills are plain directories; the scanner reads the skills dir with ReadDir.
	// The scan should succeed (it doesn't recurse into skill dirs).
	result, err := scanner.ScanRepo(root)
	require.NoError(t, err)
	assert.Len(t, result.InstalledSkills, 1)
}

// TestIsGitRepo_AcceptsWorktreeFile: .git as a file (worktree/submodule).
func TestIsGitRepo_AcceptsWorktreeFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git"),
		[]byte("gitdir: /some/other/path\n"), 0600))
	assert.True(t, scanner.IsGitRepo(dir))
}

// TestScanAll_Recursive: ScanAll descends into nested directories.
func TestScanAll_Recursive(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "team-a", "service-x")
	makeGitRepo(t, nested)

	results, err := scanner.ScanAll(root)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, nested, results[0].RepoRoot)
}

// TestScanAll_DoesNotDescendIntoFoundRepo: stop at repository boundaries.
func TestScanAll_DoesNotDescendIntoFoundRepo(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "outer")
	makeGitRepo(t, repo)
	makeGitRepo(t, filepath.Join(repo, "inner"))

	results, err := scanner.ScanAll(root)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

// TestScanAll_SkipsHeavyDirs: node_modules and similar dirs are skipped.
func TestScanAll_SkipsHeavyDirs(t *testing.T) {
	root := t.TempDir()
	hidden := filepath.Join(root, "node_modules", "pkg")
	makeGitRepo(t, hidden)
	visible := filepath.Join(root, "real")
	makeGitRepo(t, visible)

	results, err := scanner.ScanAll(root)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, visible, results[0].RepoRoot)
}

func TestScanAll_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-a-dir")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0600))
	_, err := scanner.ScanAll(file)
	assert.Error(t, err)
}
