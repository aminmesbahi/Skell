package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSkillURL_SpecificSkill(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills/akka-testing-patterns")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/Aaronontheweb/dotnet-skills", u.GitURL)
	assert.Equal(t, "dotnet-skills", u.Alias)
	assert.Equal(t, "akka-testing-patterns", u.SkillName)
	assert.Equal(t, "skills/akka-testing-patterns", u.SubPath)
}

func TestParseSkillURL_RegistryRoot(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/Aaronontheweb/dotnet-skills", u.GitURL)
	assert.Equal(t, "dotnet-skills", u.Alias)
	assert.Equal(t, "", u.SkillName)
	assert.Equal(t, "skills", u.SubPath)
}

func TestParseSkillURL_SubdirRegistryRoot(t *testing.T) {
	// e.g. /tree/main/ai/claude — 2 segments, but "ai/claude" is a skills root, not a skill named "claude"
	// ParseSkillURL can't know this at parse time; it sets SkillName = "claude" and SubPath = "ai/claude".
	// The fallback in AddFromURL resolves the ambiguity at runtime.
	u, err := ParseSkillURL("https://github.com/owner/myrepo/tree/main/ai/claude")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/myrepo", u.GitURL)
	assert.Equal(t, "myrepo", u.Alias)
	assert.Equal(t, "claude", u.SkillName)
	assert.Equal(t, "ai/claude", u.SubPath)
}

func TestParseSkillURL_DeepSkillPath(t *testing.T) {
	// 3 segments: ai/claude/my-skill → skill name = my-skill, subpath = ai/claude/my-skill
	u, err := ParseSkillURL("https://github.com/owner/myrepo/tree/main/ai/claude/my-skill")
	require.NoError(t, err)
	assert.Equal(t, "my-skill", u.SkillName)
	assert.Equal(t, "ai/claude/my-skill", u.SubPath)
}

func TestParseSkillURL_PlainGitURL(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/davidfowl/dotnet-skillz")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/davidfowl/dotnet-skillz", u.GitURL)
	assert.Equal(t, "dotnet-skillz", u.Alias)
	assert.Equal(t, "", u.SkillName)
}

func TestParseSkillURL_BranchOnlyNoSubpath(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/owner/repo/tree/main")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/repo", u.GitURL)
	assert.Equal(t, "repo", u.Alias)
	assert.Equal(t, "", u.SkillName)
}

func TestParseSkillURL_DeepSubpath(t *testing.T) {
	// e.g. packages/frontend/skills/my-skill → skill name = my-skill
	u, err := ParseSkillURL("https://github.com/owner/repo/tree/main/packages/skills/my-skill")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/repo", u.GitURL)
	assert.Equal(t, "my-skill", u.SkillName)
}

func TestParseSkillURL_TrailingSlash(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/owner/repo/tree/main/skills/my-skill/")
	require.NoError(t, err)
	assert.Equal(t, "akka-testing-patterns", func() string {
		u2, _ := ParseSkillURL("https://github.com/A/dotnet-skills/tree/master/skills/akka-testing-patterns/")
		return u2.SkillName
	}())
	assert.Equal(t, "my-skill", u.SkillName)
}

func TestParseSkillURL_AliasLowercased(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/Owner/MySkills/tree/main/skills/tool")
	require.NoError(t, err)
	assert.Equal(t, "myskills", u.Alias)
}

func TestParseSkillURL_InvalidURL(t *testing.T) {
	_, err := ParseSkillURL("not-a-url")
	assert.Error(t, err)
}

func TestParseSkillURL_RelativeURL(t *testing.T) {
	// Clearly relative (no leading /) should still be rejected
	_, err := ParseSkillURL("relative/path")
	assert.Error(t, err)
}

func TestParseSkillURL_GitSuffix(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/owner/my-skills.git")
	require.NoError(t, err)
	assert.Equal(t, "my-skills", u.Alias)
	assert.Equal(t, "https://github.com/owner/my-skills.git", u.GitURL)
}

func TestParseSkillURL_BlobLinkToSkillMD(t *testing.T) {
	// People often copy the link to SKILL.md itself
	u, err := ParseSkillURL("https://github.com/owner/repo/blob/main/skills/pdf-processing/SKILL.md")
	require.NoError(t, err)
	assert.Equal(t, "pdf-processing", u.SkillName)
	assert.Equal(t, "skills/pdf-processing", u.SubPath)
}

func TestParseSkillURL_BlobDirectory(t *testing.T) {
	u, err := ParseSkillURL("https://github.com/owner/repo/blob/main/skills")
	require.NoError(t, err)
	assert.Equal(t, "", u.SkillName)
	assert.Equal(t, "skills", u.SubPath)
}

func TestParseSkillURL_SupportsBothTreeAndBlob(t *testing.T) {
	for _, url := range []string{
		"https://github.com/o/r/tree/main/foo/bar",
		"https://github.com/o/r/blob/main/foo/bar",
	} {
		u, err := ParseSkillURL(url)
		require.NoError(t, err)
		assert.Equal(t, "bar", u.SkillName)
	}
}

func TestParseSkillURL_BranchWithSlash(t *testing.T) {
	// Encoded branch names with /
	u, err := ParseSkillURL("https://github.com/owner/repo/tree/feature%2Fwip/skills/tool")
	require.NoError(t, err)
	assert.Equal(t, "tool", u.SkillName)
}

func TestParseSkillURL_LocalPath_SingleSkillFolder_SetsSkillName(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "tesla-design-system")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: tesla-design-system\n---\n"), 0o600))

	u, err := ParseSkillURL(skillDir)
	require.NoError(t, err)
	assert.Equal(t, skillDir, u.GitURL)
	assert.Equal(t, "tesla-design-system", u.Alias)
	assert.Equal(t, "tesla-design-system", u.SkillName,
		"a folder with SKILL.md at root should be installed as a single skill")
}

func TestParseSkillURL_LocalPath_RegistryRoot_LeavesSkillNameEmpty(t *testing.T) {
	dir := t.TempDir()
	// No SKILL.md at the root — just nested skill folders.
	nested := filepath.Join(dir, "skill-a")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "SKILL.md"), []byte("---\nname: skill-a\n---\n"), 0o600))

	u, err := ParseSkillURL(dir)
	require.NoError(t, err)
	assert.Equal(t, "", u.SkillName,
		"a folder without a top-level SKILL.md should be registered as a registry root")
}

func TestParseSkillURL_LocalPath_NonExistent_LeavesSkillNameEmpty(t *testing.T) {
	// Path doesn't exist — we shouldn't crash, and we shouldn't claim it's a
	// single-skill folder.
	u, err := ParseSkillURL("/definitely/does/not/exist/anywhere")
	require.NoError(t, err)
	assert.Equal(t, "", u.SkillName)
}

func TestIsLocalPathForAdd_WindowsDriveGuard(t *testing.T) {
	assert.False(t, isLocalPathForAdd("C:"))
	assert.True(t, isLocalPathForAdd("C:\\skills"))
	assert.True(t, isLocalPathForAdd("C:/skills"))
}
