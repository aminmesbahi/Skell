package validator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSkill(t *testing.T, name, content string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0600))
	return dir
}

func TestValidateDir_ValidSkill_NoErrors(t *testing.T) {
	dir := writeSkill(t, "good-skill", "---\nname: good-skill\ndescription: A clear, useful skill. Use when testing the validator adapter end to end.\n---\n\n# Good Skill\n\nDo the thing. Run the command.\n")
	res := ValidateDir(context.Background(), dir, Options{})
	require.NotNil(t, res)
	assert.False(t, res.HasErrors(), "valid skill should have no errors: %+v", res.Findings)
}

func TestValidateDir_MissingDescription_ReportsError(t *testing.T) {
	// The Agent Skills spec requires both name and description.
	dir := writeSkill(t, "bad-skill", "---\nname: bad-skill\n---\n\n# Bad Skill\n")
	res := ValidateDir(context.Background(), dir, Options{})
	require.NotNil(t, res)
	assert.True(t, res.HasErrors(), "skill missing description should error")
	assert.NotEmpty(t, res.Findings)
}

func TestValidateDir_NoFrontmatter_ReportsError(t *testing.T) {
	dir := writeSkill(t, "raw-skill", "# Just markdown, no frontmatter\n")
	res := ValidateDir(context.Background(), dir, Options{})
	require.NotNil(t, res)
	assert.True(t, res.HasErrors())
}

func TestValidateDir_StructureOnly_NoContentAnalysis(t *testing.T) {
	dir := writeSkill(t, "good-skill", "---\nname: good-skill\ndescription: A clear, useful skill. Use when testing analysis surfacing.\n---\n\n# Good Skill\n\nDo the thing. Run the command.\n")
	res := ValidateDir(context.Background(), dir, Options{})
	require.NotNil(t, res)
	// Structure-only: no content/contamination metrics. (Token counts may still
	// populate an Analysis with totals, but HasContent must be false.)
	if res.Analysis != nil {
		assert.False(t, res.Analysis.HasContent)
		assert.False(t, res.Analysis.HasContamination)
	}
}

func TestValidateDir_Full_PopulatesAnalysis(t *testing.T) {
	dir := writeSkill(t, "good-skill", "---\nname: good-skill\ndescription: A clear, useful skill. Use when testing analysis surfacing end to end.\n---\n\n# Good Skill\n\nDo the thing. Run the command when needed.\n\n## Steps\n\n- First do this.\n- Then do that.\n\n```python\nprint(\"hello\")\n```\n")
	res := ValidateDir(context.Background(), dir, Options{Content: true, Contamination: true})
	require.NotNil(t, res)
	require.NotNil(t, res.Analysis, "full analysis should populate Analysis")
	assert.True(t, res.Analysis.HasContent)
	assert.True(t, res.Analysis.HasContamination)
	assert.Greater(t, res.Analysis.WordCount, 0)
	assert.Greater(t, res.Analysis.TotalTokens, 0)
}
