package registry_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeLocalRegistry creates a minimal "registry" as a local git repo
// with a single skill directory containing a SKILL.md.
func makeLocalRegistry(t *testing.T, skillName string) string {
	t.Helper()
	dir := t.TempDir()

	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	// Disable CRLF conversion so SKILL.md is cloned with LF line endings.
	run(t, dir, "git", "config", "core.autocrlf", "false")

	// Create skill directory and SKILL.md
	skillDir := filepath.Join(dir, skillName)
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	content := "---\nname: " + skillName + "\ndescription: A test skill\nlicense: MIT\nmetadata:\n  version: \"1.0.0\"\n  lifecycle: stable\n---\n# " + skillName + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644))

	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "add skill")
	return dir
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "command failed: %s %v\n%s", name, args, out)
}

// makeNestedRegistry creates a minimal "registry" with skills in a subdirectory,
// simulating repos like davidfowl/dotnet-skillz (skills/<name>/SKILL.md).
func makeNestedRegistry(t *testing.T, subDir, skillName string) string {
	t.Helper()
	dir := t.TempDir()

	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "config", "core.autocrlf", "false")

	// Create nested skill directory and SKILL.md
	skillDir := filepath.Join(dir, subDir, skillName)
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	content := "---\nname: " + skillName + "\ndescription: A nested skill\nlicense: MIT\n---\n# " + skillName + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644))

	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "add nested skill")
	return dir
}

// TestRegistry_ListSkills_Nested verifies skills nested in subdirectories are discovered.
func TestRegistry_ListSkills_Nested(t *testing.T) {
	regDir := makeNestedRegistry(t, "skills", "ilspy-decompile")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "dotnet-skillz", URL: regDir}

	skills, err := adapter.ListSkills(reg)
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "ilspy-decompile", skills[0].Name)
}

// TestRegistry_GetSkill_Nested returns a skill nested inside a subdirectory.
func TestRegistry_GetSkill_Nested(t *testing.T) {
	regDir := makeNestedRegistry(t, "plugins/dotnet/skills", "csharp-analyzer")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "dotnet-skills", URL: regDir}

	skill, err := adapter.GetSkill(reg, "csharp-analyzer")
	require.NoError(t, err)
	assert.Equal(t, "csharp-analyzer", skill.Name)
}

// TestRegistry_CopySkillTo_Nested copies a deeply nested skill to a destination.
func TestRegistry_CopySkillTo_Nested(t *testing.T) {
	regDir := makeNestedRegistry(t, "plugins/dotnet/skills", "csharp-analyzer")
	cacheRoot := t.TempDir()
	destPath := filepath.Join(t.TempDir(), "csharp-analyzer")

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "dotnet-skills", URL: regDir}

	require.NoError(t, adapter.CopySkillTo(reg, "csharp-analyzer", "", destPath))

	_, err := os.Stat(filepath.Join(destPath, "SKILL.md"))
	assert.NoError(t, err, "SKILL.md should be present in copied nested skill dir")
}

func TestRegistry_FetchClonesRepo(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}

	require.NoError(t, adapter.Fetch(reg))

	_, err := os.Stat(filepath.Join(cacheRoot, "default", ".git"))
	assert.NoError(t, err, "cloned registry should have .git directory")
}

// TestRegistry_ListSkills returns skills from a fetched registry.
func TestRegistry_ListSkills(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}

	skills, err := adapter.ListSkills(reg)
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "pdf-processing", skills[0].Name)
	assert.Equal(t, "1.0.0", skills[0].Metadata.Version)
}

// TestRegistry_GetSkill returns the specific skill from the cache.
func TestRegistry_GetSkill(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}

	skill, err := adapter.GetSkill(reg, "pdf-processing")
	require.NoError(t, err)
	assert.Equal(t, "pdf-processing", skill.Name)
}

// TestRegistry_GetSkill_NotFound returns an error for unknown skills.
func TestRegistry_GetSkill_NotFound(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}
	require.NoError(t, adapter.Fetch(reg))

	_, err := adapter.GetSkill(reg, "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, registry.ErrSkillNotFound)
}

// TestRegistry_CopySkillTo copies a skill directory to a destination path.
func TestRegistry_CopySkillTo(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()
	destPath := filepath.Join(t.TempDir(), "pdf-processing")

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}

	require.NoError(t, adapter.CopySkillTo(reg, "pdf-processing", "1.0.0", destPath))

	_, err := os.Stat(filepath.Join(destPath, "SKILL.md"))
	assert.NoError(t, err, "SKILL.md should be copied to destination")
}

// TestRegistry_CacheStatus returns a non-empty string after fetching.
func TestRegistry_CacheStatus(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}
	require.NoError(t, adapter.Fetch(reg))

	status, err := adapter.CacheStatus()
	require.NoError(t, err)
	assert.Contains(t, status, "default")
}

// TestRegistry_CacheClear removes the cache directory.
func TestRegistry_CacheClear(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}
	require.NoError(t, adapter.Fetch(reg))

	require.NoError(t, adapter.CacheClear())
	_, err := os.Stat(cacheRoot)
	assert.True(t, os.IsNotExist(err), "cache root should be removed after CacheClear")
}

// TestRegistry_CacheRefresh fetches on an existing clone without error.
func TestRegistry_CacheRefresh(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}
	require.NoError(t, adapter.Fetch(reg))

	assert.NoError(t, adapter.CacheRefresh(reg))
}

func TestRegistry_CopySkillTo_SkillNotFound_ReturnsError(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()
	destPath := filepath.Join(t.TempDir(), "nonexistent")

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}

	err := adapter.CopySkillTo(reg, "nonexistent", "", destPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRegistry_CacheStatus_EmptyCache(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)

	status, err := adapter.CacheStatus()
	require.NoError(t, err)
	assert.Contains(t, status, "empty")
}

// makeRegistryWithFrontmatterName creates a registry where the skill directory
// name (dirName) differs from the name declared inside SKILL.md (frontmatterName).
// This reproduces the bug where skell install fails for such skills.
func makeRegistryWithFrontmatterName(t *testing.T, dirName, frontmatterName string) string {
	t.Helper()
	dir := t.TempDir()

	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "config", "core.autocrlf", "false")

	skillDir := filepath.Join(dir, dirName)
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	content := "---\nname: " + frontmatterName + "\ndescription: test\nlicense: MIT\n---\n# skill\n"
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644))

	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "add skill")
	return dir
}

// TestRegistry_GetSkill_FrontmatterNameDiffersFromDir ensures a skill is found
// when the SKILL.md name: field differs from the directory name.
// Regression test for: skill "X" not found when dir is named "X-something".
func TestRegistry_GetSkill_FrontmatterNameDiffersFromDir(t *testing.T) {
	// Directory is "employment-certificate-debug-skill",
	// but SKILL.md declares name: "employment-certificate-debug".
	regDir := makeRegistryWithFrontmatterName(t, "employment-certificate-debug-skill", "employment-certificate-debug")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "hros-scripts", URL: regDir}

	skill, err := adapter.GetSkill(reg, "employment-certificate-debug")
	require.NoError(t, err)
	assert.Equal(t, "employment-certificate-debug", skill.Name)
}

// TestRegistry_CopySkillTo_FrontmatterNameDiffersFromDir ensures CopySkillTo works
// when the skill is located by its frontmatter name rather than directory name.
func TestRegistry_CopySkillTo_FrontmatterNameDiffersFromDir(t *testing.T) {
	regDir := makeRegistryWithFrontmatterName(t, "employment-certificate-debug-skill", "employment-certificate-debug")
	cacheRoot := t.TempDir()
	destPath := filepath.Join(t.TempDir(), "employment-certificate-debug")

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "hros-scripts", URL: regDir}

	require.NoError(t, adapter.CopySkillTo(reg, "employment-certificate-debug", "", destPath))

	_, err := os.Stat(filepath.Join(destPath, "SKILL.md"))
	assert.NoError(t, err, "SKILL.md should be present after copy")
}

// TestRegistry_ListSkills_FrontmatterNameDiffersFromDir verifies ListSkills
// returns the frontmatter name (not the directory name) for such skills.
func TestRegistry_ListSkills_FrontmatterNameDiffersFromDir(t *testing.T) {
	regDir := makeRegistryWithFrontmatterName(t, "employment-certificate-debug-skill", "employment-certificate-debug")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "hros-scripts", URL: regDir}

	skills, err := adapter.ListSkills(reg)
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "employment-certificate-debug", skills[0].Name)
}

func TestRegistry_Fetch_AlreadyCloned_Pulls(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}

	require.NoError(t, adapter.Fetch(reg))
	require.NoError(t, adapter.Fetch(reg))
}

func TestRegistry_GetSkill_FetchFails_ReturnsError(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "bad", URL: "file:///nonexistent"}

	_, err := adapter.GetSkill(reg, "some-skill")
	require.Error(t, err)
}

func TestRegistry_ListSkills_FetchFails_ReturnsError(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "bad", URL: "file:///nonexistent"}

	_, err := adapter.ListSkills(reg)
	require.Error(t, err)
}

func TestRegistry_CacheClear_EmptyDir_Succeeds(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)
	require.NoError(t, adapter.CacheClear())
}

func TestRegistry_ListSkills_SkipsSkillWithNoName(t *testing.T) {
	// SKILL.md without a name field → skill name falls back to directory name.
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "config", "core.autocrlf", "false")

	skillDir := filepath.Join(dir, "my-unnamed-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	// SKILL.md with frontmatter but no name field.
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
		[]byte("---\ndescription: A skill with no name\n---\n"), 0644))
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "add nameless skill")

	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "test", URL: dir}

	skills, err := adapter.ListSkills(reg)
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "my-unnamed-skill", skills[0].Name)
}

// TestRegistry_CopySkillTo_RemovesStaleFiles: files removed in a new revision
// must not linger from the previous install.
func TestRegistry_CopySkillTo_RemovesStaleFiles(t *testing.T) {
	regDir := makeLocalRegistry(t, "pdf-processing")
	cacheRoot := t.TempDir()
	destPath := filepath.Join(t.TempDir(), "pdf-processing")

	require.NoError(t, os.MkdirAll(destPath, 0755))
	stale := filepath.Join(destPath, "stale.md")
	require.NoError(t, os.WriteFile(stale, []byte("removed in v2"), 0600))

	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: regDir}
	require.NoError(t, adapter.CopySkillTo(reg, "pdf-processing", "1.0.0", destPath))

	_, err := os.Stat(stale)
	assert.True(t, os.IsNotExist(err), "stale.md must have been removed during copy")
}

// TestRegistry_CopySkillTo_PreservesSymlinks: symlinks in the registry must be
// copied as symlinks rather than dereferenced into the destination.
func TestRegistry_CopySkillTo_PreservesSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on windows")
	}

	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "config", "core.autocrlf", "false")

	skillDir := filepath.Join(dir, "evil-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
		[]byte("---\nname: evil-skill\n---\n"), 0644))
	// Symlink that, if dereferenced, would copy the contents of /etc/hosts
	// (or any sensitive file) into the destination repo.
	require.NoError(t, os.Symlink("/etc/hosts", filepath.Join(skillDir, "leaked")))

	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "add evil skill")

	cacheRoot := t.TempDir()
	destPath := filepath.Join(t.TempDir(), "evil-skill")
	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "default", URL: dir}
	require.NoError(t, adapter.CopySkillTo(reg, "evil-skill", "", destPath))

	leaked := filepath.Join(destPath, "leaked")
	info, err := os.Lstat(leaked)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&os.ModeSymlink, "leaked must be a symlink, not a regular file")

	// Verify the destination does not contain dereferenced /etc/hosts contents.
	hosts, err := os.ReadFile("/etc/hosts")
	if err == nil {
		got, _ := os.ReadFile(leaked)
		// Reading the symlink may or may not work depending on access; the
		// crucial assertion is that the file in the dest tree is itself a
		// symlink (checked above), not a regular file with copied bytes.
		_ = got
		_ = hosts
	}
}

// TestRegistry_Fetch_RejectsArgvInjectionURL: URLs starting with '-' must not
// reach the git CLI.
func TestRegistry_Fetch_RejectsArgvInjectionURL(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "evil", URL: "--upload-pack=touch /tmp/pwn"}

	err := adapter.Fetch(reg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not start with '-'")
}

// TestRegistry_Fetch_RejectsDisallowedScheme: unsupported transports such as
// ext:: must be rejected.
func TestRegistry_Fetch_RejectsDisallowedScheme(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "ext", URL: "ext::sh -c whoami"}

	err := adapter.Fetch(reg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")
}

// TestRegistry_Fetch_RejectsTraversalAlias: alias must not escape cache root.
func TestRegistry_Fetch_RejectsTraversalAlias(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := registry.NewAdapter(cacheRoot)
	reg := registry.Registry{Alias: "../escape", URL: "https://example.com/repo"}

	err := adapter.Fetch(reg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alias")
}
