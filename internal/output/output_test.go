package output_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func newBuf() (*output.Printer, *bytes.Buffer) {
	var buf bytes.Buffer
	return output.NewPrinterTo(&buf, false), &buf
}

func newJSONBuf() (*output.Printer, *bytes.Buffer) {
	var buf bytes.Buffer
	return output.NewPrinterTo(&buf, true), &buf
}

func newStdout() *output.Printer {
	return output.NewPrinter(false)
}

// ── NewPrinter / NewPrinterTo ─────────────────────────────────────────────

func TestNewPrinter_NotNil(t *testing.T) {
	p := newStdout()
	assert.NotNil(t, p)
}

func TestNewPrinterTo_NotNil(t *testing.T) {
	p, _ := newBuf()
	assert.NotNil(t, p)
}

// ── PrintStatusTable ─────────────────────────────────────────────────────────

func TestPrintStatusTable_Text(t *testing.T) {
	p, buf := newBuf()
	entries := []model.StatusEntry{
		{Name: "skill-a", Installed: "1.0.0", Latest: "1.1.0", Status: model.StatusOutdated},
		{Name: "skill-b", Installed: "2.0.0", Latest: "2.0.0", Status: model.StatusUpToDate},
	}
	p.PrintStatusTable(entries)
	out := buf.String()
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "skill-a")
	assert.Contains(t, out, "outdated")
	assert.Contains(t, out, "skill-b")
}

func TestPrintStatusTable_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	entries := []model.StatusEntry{
		{Name: "skill-a", Installed: "1.0.0", Latest: "1.1.0", Status: model.StatusOutdated},
	}
	p.PrintStatusTable(entries)
	var out []model.StatusEntry
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "skill-a", out[0].Name)
}

// ── PrintSkillList ────────────────────────────────────────────────────────────

func TestPrintSkillList_Text_Pinned(t *testing.T) {
	p, buf := newBuf()
	skills := []model.InstalledSkill{
		{Name: "foo", Version: "1.0.0", Registry: "reg1", Pinned: true},
		{Name: "bar", Version: "2.0.0", Registry: "reg2", Pinned: false},
	}
	p.PrintSkillList(skills)
	out := buf.String()
	assert.Contains(t, out, "foo")
	assert.Contains(t, out, "yes")
	assert.Contains(t, out, "bar")
}

func TestPrintSkillList_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	skills := []model.InstalledSkill{{Name: "foo", Registry: "reg1"}}
	p.PrintSkillList(skills)
	var out []model.InstalledSkill
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "foo", out[0].Name)
}

// ── PrintRegistrySkillList ────────────────────────────────────────────────────

func TestPrintRegistrySkillList_Text(t *testing.T) {
	p, buf := newBuf()
	skills := []model.RegistrySkill{
		{Name: "baz", Description: "desc", Metadata: model.SkillMetadata{Version: "1.0", Lifecycle: "stable", Owner: "team"}},
	}
	p.PrintRegistrySkillList(skills)
	out := buf.String()
	assert.Contains(t, out, "baz")
	assert.Contains(t, out, "1.0")
}

func TestPrintRegistrySkillList_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	skills := []model.RegistrySkill{{Name: "baz"}}
	p.PrintRegistrySkillList(skills)
	var out []model.RegistrySkill
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "baz", out[0].Name)
}

// ── PrintInfoResult ───────────────────────────────────────────────────────────

func TestPrintInfoResult_FullSkill_Text(t *testing.T) {
	p, buf := newBuf()
	r := &model.InfoResult{
		Skill: &model.RegistrySkill{
			Name:        "my-skill",
			Description: "does stuff",
			License:     "MIT",
			Metadata: model.SkillMetadata{
				Version:    "1.2.3",
				Owner:      "team",
				Lifecycle:  "stable",
				Scope:      "repo",
				Tags:       "go,test",
				SourceRepo: "https://github.com/foo/bar",
			},
		},
		Lock: &model.InstalledSkill{
			Version:     "1.2.3",
			InstalledAt: "2024-01-01T00:00:00Z",
			ContentHash: "sha256:abc",
			Pinned:      true,
		},
		Status: model.StatusUpToDate,
	}
	p.PrintInfoResult("my-skill", r)
	out := buf.String()
	assert.Contains(t, out, "my-skill")
	assert.Contains(t, out, "1.2.3")
	assert.Contains(t, out, "does stuff")
	assert.Contains(t, out, "MIT")
	assert.Contains(t, out, "stable")
	assert.Contains(t, out, "repo")
	assert.Contains(t, out, "go,test")
	assert.Contains(t, out, "https://github.com/foo/bar")
	assert.Contains(t, out, "sha256:abc")
	assert.Contains(t, out, "pinned")
	assert.Contains(t, out, "up-to-date")
}

func TestPrintInfoResult_NoSkillNoLock_Text(t *testing.T) {
	p, buf := newBuf()
	r := &model.InfoResult{}
	p.PrintInfoResult("empty-skill", r)
	out := buf.String()
	assert.Contains(t, out, "empty-skill")
}

func TestPrintInfoResult_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	r := &model.InfoResult{
		Skill:  &model.RegistrySkill{Name: "my-skill"},
		Status: model.StatusUpToDate,
	}
	p.PrintInfoResult("my-skill", r)
	assert.True(t, json.Valid(buf.Bytes()))
}

// ── PrintAction ───────────────────────────────────────────────────────────────

func TestPrintAction_Text_Normal(t *testing.T) {
	p, buf := newBuf()
	p.PrintAction(output.ActionEvent{Action: "install", Skill: "foo", Repo: "/repo"})
	assert.Contains(t, buf.String(), "install")
	assert.Contains(t, buf.String(), "foo")
}

func TestPrintAction_Text_DryRun(t *testing.T) {
	p, buf := newBuf()
	p.PrintAction(output.ActionEvent{Action: "install", Skill: "foo", DryRun: true})
	assert.Contains(t, buf.String(), "dry-run")
	assert.Contains(t, buf.String(), "would install foo")
}

func TestPrintAction_Text_WithMessage(t *testing.T) {
	p, buf := newBuf()
	p.PrintAction(output.ActionEvent{Action: "skip", Skill: "foo", Message: "already up-to-date"})
	assert.Contains(t, buf.String(), "already up-to-date")
}

func TestPrintAction_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	p.PrintAction(output.ActionEvent{Action: "install", Skill: "foo"})
	var out output.ActionEvent
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "install", out.Action)
}

// ── PrintDiagnostics ──────────────────────────────────────────────────────────

func TestPrintDiagnostics_NoIssues_Text(t *testing.T) {
	p, buf := newBuf()
	p.PrintDiagnostics(nil)
	assert.Contains(t, buf.String(), "no issues found")
}

func TestPrintDiagnostics_WithHint_Text(t *testing.T) {
	p, buf := newBuf()
	issues := []output.DiagnosticEntry{
		{Severity: "error", Code: "no-manifest", Message: "missing manifest", Hint: "run skell init"},
		{Severity: "warning", Code: "drift", Message: "skill drift detected", Hint: ""},
	}
	p.PrintDiagnostics(issues)
	out := buf.String()
	assert.Contains(t, out, "error")
	assert.Contains(t, out, "missing manifest")
	assert.Contains(t, out, "run skell init")
	assert.Contains(t, out, "drift")
}

func TestPrintDiagnostics_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	p.PrintDiagnostics([]output.DiagnosticEntry{{Severity: "error", Message: "bad"}})
	assert.True(t, json.Valid(buf.Bytes()))
}

// ── Success / Error ───────────────────────────────────────────────────────────

func TestSuccess_Text(t *testing.T) {
	p, buf := newBuf()
	p.Success("all done")
	assert.Contains(t, buf.String(), "all done")
}

func TestSuccess_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	p.Success("all done")
	var out output.ActionEvent
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "success", out.Action)
	assert.Equal(t, "all done", out.Message)
}

func TestError_Text_WithHint(t *testing.T) {
	p, buf := newBuf()
	p.Error("something broke", "try again")
	out := buf.String()
	assert.Contains(t, out, "something broke")
	assert.Contains(t, out, "try again")
}

func TestError_Text_NoHint(t *testing.T) {
	p, buf := newBuf()
	p.Error("oops", "")
	assert.Contains(t, buf.String(), "oops")
	assert.False(t, strings.Contains(buf.String(), "hint"))
}

func TestError_JSON(t *testing.T) {
	p, buf := newJSONBuf()
	p.Error("oops", "fix it")
	var out output.ActionEvent
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "error", out.Action)
}

// errorWriter always returns an error on Write.
type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

func TestPrintInfoResult_WriterError_NoPanic(t *testing.T) {
	// A writer that errors should not cause a panic — the fmtLineWriter
	// records the first error and silently skips subsequent writes.
	p := output.NewPrinterTo(errorWriter{}, false)
	r := &model.InfoResult{
		Skill: &model.RegistrySkill{
			Name:    "my-skill",
			License: "MIT",
			Metadata: model.SkillMetadata{
				Version:    "1.0",
				Owner:      "team",
				Lifecycle:  "stable",
				Scope:      "repo",
				Tags:       "go",
				SourceRepo: "https://github.com/foo/bar",
			},
		},
		Lock: &model.InstalledSkill{Version: "1.0", Pinned: true},
	}
	// Should not panic even though writes fail.
	assert.NotPanics(t, func() {
		p.PrintInfoResult("my-skill", r)
	})
}
