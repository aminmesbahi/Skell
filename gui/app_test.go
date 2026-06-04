package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestResolveToolBinary_UsesEnvOverride(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, toolFilename("skell"))
	require.NoError(t, os.WriteFile(bin, []byte(""), 0600))

	oldEnv := os.Getenv("SKELL_BIN")
	require.NoError(t, os.Setenv("SKELL_BIN", bin))
	t.Cleanup(func() { _ = os.Setenv("SKELL_BIN", oldEnv) })

	resolved, err := resolveToolBinary("skell", "SKELL_BIN", "install it")
	require.NoError(t, err)
	assert.Equal(t, bin, resolved)
}

func TestResolveToolBinary_UsesBundledCandidate(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, toolFilename("skell"))
	require.NoError(t, os.WriteFile(bin, []byte(""), 0600))

	oldExec := currentExecutable
	oldLookPath := lookPath
	currentExecutable = func() (string, error) { return filepath.Join(dir, "skell-gui.exe"), nil }
	lookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	t.Cleanup(func() {
		currentExecutable = oldExec
		lookPath = oldLookPath
	})

	resolved, err := resolveToolBinary("skell", "SKELL_BIN", "install it")
	require.NoError(t, err)
	assert.Equal(t, bin, resolved)
}

func TestResolveToolBinary_SkipsCurrentExecutableCandidate(t *testing.T) {
	dir := t.TempDir()
	selfBin := filepath.Join(dir, toolFilename("skell"))
	cliBin := filepath.Join(dir, "bin", toolFilename("skell"))
	require.NoError(t, os.MkdirAll(filepath.Dir(cliBin), 0755))
	require.NoError(t, os.WriteFile(selfBin, []byte(""), 0600))
	require.NoError(t, os.WriteFile(cliBin, []byte(""), 0600))

	oldExec := currentExecutable
	oldLookPath := lookPath
	oldDirs := extraToolSearchDirs
	currentExecutable = func() (string, error) { return selfBin, nil }
	lookPath = func(string) (string, error) { return cliBin, nil }
	extraToolSearchDirs = func() []string { return nil } // isolate from system-installed skell
	t.Cleanup(func() {
		currentExecutable = oldExec
		lookPath = oldLookPath
		extraToolSearchDirs = oldDirs
	})

	resolved, err := resolveToolBinary("skell", "SKELL_BIN", "install it")
	require.NoError(t, err)
	assert.Equal(t, cliBin, resolved)
}

func TestResolveToolBinary_RejectsCurrentExecutableFromPath(t *testing.T) {
	dir := t.TempDir()
	selfBin := filepath.Join(dir, toolFilename("skell"))
	require.NoError(t, os.WriteFile(selfBin, []byte(""), 0600))

	oldExec := currentExecutable
	oldLookPath := lookPath
	oldDirs := extraToolSearchDirs
	currentExecutable = func() (string, error) { return selfBin, nil }
	lookPath = func(string) (string, error) { return selfBin, nil }
	extraToolSearchDirs = func() []string { return nil } // isolate from system-installed skell
	t.Cleanup(func() {
		currentExecutable = oldExec
		lookPath = oldLookPath
		extraToolSearchDirs = oldDirs
	})

	resolved, err := resolveToolBinary("skell", "SKELL_BIN", "install it")
	require.Error(t, err)
	assert.Empty(t, resolved)
	assert.Contains(t, err.Error(), "running GUI executable")
}

func TestParseValidationOutput_NullFindings(t *testing.T) {
	out := `[{"name":"clean","result":{"findings":null,"errors":0,"warnings":0}}]`
	results, err := parseValidationOutput(out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Findings == nil {
		t.Fatal("expected empty findings slice, got nil")
	}
	if len(results[0].Findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(results[0].Findings))
	}
}

func TestParseValidationOutput(t *testing.T) {
	out := `[{"name":"good","result":{"skill_dir":"/x/good","findings":[],"errors":0,"warnings":0}},{"name":"bad","result":{"errors":1,"warnings":1,"findings":[{"severity":"error","category":"Frontmatter","message":"description is required","file":"SKILL.md","line":0}]}}]`
	results, err := parseValidationOutput(out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "good" || results[0].Errors != 0 {
		t.Errorf("unexpected first result: %+v", results[0])
	}
	if results[1].Errors != 1 || len(results[1].Findings) != 1 || results[1].Findings[0].Category != "Frontmatter" {
		t.Errorf("unexpected second result: %+v", results[1])
	}
}
