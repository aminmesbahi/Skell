package registry_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
	assert.Error(t, err)
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
