package main

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSkillMetadataFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    SkillMetadataFields
	}{
		{
			name: "full metadata block",
			content: `---
name: pdf-processing
description: Extract text from PDFs
metadata:
  version: "2.1.0"
  owner: platform-team
  lifecycle: stable
  tags: pdf, extraction
---
Body here`,
			want: SkillMetadataFields{
				Description: "Extract text from PDFs",
				Version:     "2.1.0",
				Owner:       "platform-team",
				Lifecycle:   "stable",
				Tags:        "pdf, extraction",
			},
		},
		{
			name: "version only under metadata",
			content: `---
name: test
metadata:
  version: 1.0.0-beta
---
`,
			want: SkillMetadataFields{Version: "1.0.0-beta"},
		},
		{
			name: "no metadata block",
			content: `---
name: simple
description: Just a skill
---`,
			want: SkillMetadataFields{Description: "Just a skill"},
		},
		{
			name:    "empty",
			content: "",
			want:    SkillMetadataFields{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseSkillMetadataFields(tc.content)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestApplyFrontmatterEdits_Version(t *testing.T) {
	original := `---
name: my-skill
description: old desc
metadata:
  version: "0.9.0"
  owner: old-owner
---
Content
`

	fields := SkillMetadataFields{
		Description: "new desc",
		Version:     "1.2.3",
		Owner:       "new-owner",
	}

	updated := applyFrontmatterEdits(original, fields)

	assert.Contains(t, updated, "description: new desc")
	assert.Contains(t, updated, "version: 1.2.3")
	assert.Contains(t, updated, "owner: new-owner")
	// old version should be gone
	assert.NotContains(t, updated, "version: \"0.9.0\"")
}

func TestApplyFrontmatterEdits_InsertsVersionWhenMissing(t *testing.T) {
	original := `---
name: fresh-skill
description: brand new
---
`

	fields := SkillMetadataFields{Version: "0.1.0"}

	updated := applyFrontmatterEdits(original, fields)

	assert.Contains(t, updated, "metadata:")
	assert.Contains(t, updated, "version: 0.1.0")
}

func TestContributeMetadata_GhMissing(t *testing.T) {
	// Only run this test if gh is actually missing on the machine.
	// If gh exists in the test env, we skip so we don't break real contribution tests.
	if _, err := exec.LookPath("gh"); err == nil {
		t.Skip("gh CLI is present in this environment; skipping missing-gh test")
	}

	app := NewApp()
	res := app.ContributeMetadata(ContributeParams{
		SourceRepo: "https://github.com/example/repo",
		SkillName:  "test",
	})

	require.False(t, res.Success)
	assert.Contains(t, strings.ToLower(res.Error), "github cli")
	assert.Contains(t, res.Error, "cli.github.com")
}
