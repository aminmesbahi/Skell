package frontmatter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSkillMD = `---
name: pdf-processing
description: Extract PDF text, fill forms, merge files.
license: Apache-2.0
metadata:
  version: "1.2.0"
  owner: platform-team
  lifecycle: stable
  scope: shared
  tags: documents, extraction
  source_repo: https://github.com/mycompany/skills-registry
---

# PDF Processing

Use this skill when working with PDF files.
`

func TestParse_ValidFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(validSkillMD), 0600))

	skill, err := frontmatter.Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "pdf-processing", skill.Name)
	assert.Equal(t, "1.2.0", skill.Metadata.Version)
	assert.Equal(t, model.LifecycleStable, skill.Metadata.Lifecycle)
	assert.Equal(t, "platform-team", skill.Metadata.Owner)
}

func TestParse_MissingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte("# No frontmatter"), 0600))

	_, err := frontmatter.Parse(path)
	assert.Error(t, err)
}

func TestParse_MissingVersionField(t *testing.T) {
	content := `---
name: no-version-skill
description: A skill without a version.
metadata:
  owner: someone
---
`
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	skill, err := frontmatter.Parse(path)
	require.NoError(t, err)
	assert.Empty(t, skill.Metadata.Version)
}

func TestParse_NonExistentFile(t *testing.T) {
	_, err := frontmatter.Parse("/nonexistent/SKILL.md")
	assert.Error(t, err)
}

func TestParseDir_FindsSkillMD(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(validSkillMD), 0600))

	skill, err := frontmatter.ParseDir(dir)
	require.NoError(t, err)
	assert.Equal(t, "pdf-processing", skill.Name)
}

func TestParseDir_MissingSkillMD(t *testing.T) {
	dir := t.TempDir()
	_, err := frontmatter.ParseDir(dir)
	assert.Error(t, err)
}

func TestParse_MergesTopLevelMetadataFields(t *testing.T) {
	content := `---
name: merged-skill
description: merged metadata fields
license: Apache-2.0
paths: src/**
disable_model_invocation: true
compatibility: cursor
metadata:
  version: "2.0.0"
---
`
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	skill, err := frontmatter.Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "src/**", skill.Metadata.Paths)
	assert.True(t, skill.Metadata.DisableModelInvocation)
	assert.Equal(t, "cursor", skill.Metadata.Compatibility)
	assert.Equal(t, "Apache-2.0", skill.Metadata.License)
}

func TestParse_SanitizesControlCharactersInFields(t *testing.T) {
	// A malicious registry could embed a newline plus fake CLI output (or an
	// ANSI escape sequence) in a frontmatter field to spoof what the user
	// sees in `skell list`/`search`/`info`. Parse must neutralize that.
	content := "---\n" +
		"name: \"evil\\n  \\u2713  47 skills installed\"\n" +
		"description: \"line one\\nline two\"\n" +
		"metadata:\n" +
		"  owner: \"team\\u001b[31mRED\\u001b[0m\"\n" +
		"---\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	skill, err := frontmatter.Parse(path)
	require.NoError(t, err)
	assert.NotContains(t, skill.Name, "\n")
	assert.NotContains(t, skill.Description, "\n")
	assert.NotContains(t, skill.Metadata.Owner, "\x1b")
	assert.Equal(t, "line one line two", skill.Description)
}

func TestParse_RejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")

	huge := make([]byte, 0, 6<<20)
	huge = append(huge, "---\nname: huge\n---\n"...)
	for len(huge) < 6<<20 {
		huge = append(huge, 'x')
	}
	require.NoError(t, os.WriteFile(path, huge, 0600))

	_, err := frontmatter.Parse(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds")
}

func TestParse_DoesNotOverrideMetadataFields(t *testing.T) {
	content := `---
name: keep-metadata
license: Apache-2.0
paths: src/**
disable_model_invocation: true
compatibility: cursor
metadata:
  version: "2.0.0"
  paths: docs/**
  disable_model_invocation: false
  compatibility: copilot
  license: MIT
---
`
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	skill, err := frontmatter.Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "docs/**", skill.Metadata.Paths)
	assert.True(t, skill.Metadata.DisableModelInvocation)
	assert.Equal(t, "copilot", skill.Metadata.Compatibility)
	assert.Equal(t, "MIT", skill.Metadata.License)
}
