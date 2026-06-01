// Package validator wraps the external Agent Skills validator
// (github.com/agent-ecosystem/skill-validator) and adapts its results into
// Skell's own types, so the rest of the codebase does not depend on the
// validator's API surface directly.
//
// By default only the offline "structure" checks run (spec conformance:
// required frontmatter, directory layout, token budgets, code-fence integrity,
// internal link resolution). Content-quality and contamination analysis are
// offline but opt-in; external link checking performs network I/O and is also
// opt-in.
package validator

import (
	"context"
	"strings"

	"github.com/agent-ecosystem/skill-validator/orchestrate"
	"github.com/agent-ecosystem/skill-validator/structure"
	svtypes "github.com/agent-ecosystem/skill-validator/types"
)

// Severity mirrors the validator's finding levels.
type Severity string

// Severity levels, ordered from most to least serious.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Finding is a single validation result for a skill.
type Finding struct {
	Severity Severity `json:"severity"`
	Category string   `json:"category"`
	Message  string   `json:"message"`
	File     string   `json:"file,omitempty"`
	Line     int      `json:"line,omitempty"`
}

// TokenCount reports the token size of a single file in the skill.
type TokenCount struct {
	File   string `json:"file"`
	Tokens int    `json:"tokens"`
}

// Analysis holds the offline quality metrics produced when the content and/or
// contamination groups run. Nil on a structure-only validation.
type Analysis struct {
	// Content quality (from the SKILL.md content analyzer).
	WordCount              int     `json:"word_count"`
	CodeBlockRatio         float64 `json:"code_block_ratio"`
	ImperativeRatio        float64 `json:"imperative_ratio"`
	InformationDensity     float64 `json:"information_density"`
	InstructionSpecificity float64 `json:"instruction_specificity"`
	SectionCount           int     `json:"section_count"`
	ListItemCount          int     `json:"list_item_count"`
	HasContent             bool    `json:"has_content"`

	// Cross-language contamination.
	ContaminationScore   float64  `json:"contamination_score"`
	ContaminationLevel   string   `json:"contamination_level"`
	CodeLanguages        []string `json:"code_languages,omitempty"`
	MismatchedCategories []string `json:"mismatched_categories,omitempty"`
	LanguageMismatch     bool     `json:"language_mismatch"`
	HasContamination     bool     `json:"has_contamination"`

	// Token budget.
	TotalTokens int `json:"total_tokens"`
	SkillTokens int `json:"skill_tokens"`
}

// Result is the adapted outcome of validating one skill directory.
type Result struct {
	SkillDir    string       `json:"skill_dir"`
	Findings    []Finding    `json:"findings"`
	TokenCounts []TokenCount `json:"token_counts,omitempty"`
	Analysis    *Analysis    `json:"analysis,omitempty"`
	Errors      int          `json:"errors"`
	Warnings    int          `json:"warnings"`
}

// HasErrors reports whether the skill has any error-level findings.
func (r *Result) HasErrors() bool { return r != nil && r.Errors > 0 }

// HasWarnings reports whether the skill has any warning-level findings.
func (r *Result) HasWarnings() bool { return r != nil && r.Warnings > 0 }

// Options selects which check groups run. The zero value runs structure checks
// only (offline, no network).
type Options struct {
	// Content enables offline content-quality metrics.
	Content bool
	// Contamination enables offline cross-language contamination analysis.
	Contamination bool
	// Links enables external HTTP/HTTPS link validation (network I/O).
	Links bool
	// SkipOrphans disables the orphan-file reachability check.
	SkipOrphans bool
	// AllowExtraFrontmatter tolerates frontmatter keys outside the spec.
	AllowExtraFrontmatter bool
}

// ValidateDir validates the skill directory at dir and returns the adapted
// result. The context cancels network operations when Links is enabled.
func ValidateDir(ctx context.Context, dir string, opts Options) *Result {
	enabled := map[orchestrate.CheckGroup]bool{
		orchestrate.GroupStructure:     true,
		orchestrate.GroupContent:       opts.Content,
		orchestrate.GroupContamination: opts.Contamination,
		orchestrate.GroupLinks:         opts.Links,
	}
	rep := orchestrate.RunAllChecks(ctx, dir, orchestrate.Options{
		Enabled: enabled,
		StructOpts: structure.Options{
			SkipOrphans:           opts.SkipOrphans,
			AllowExtraFrontmatter: opts.AllowExtraFrontmatter,
		},
	})
	return fromReport(rep)
}

// fromReport converts the validator's report into a Skell Result, dropping
// Pass-level results (which carry no actionable signal).
func fromReport(rep *svtypes.Report) *Result {
	if rep == nil {
		return &Result{}
	}
	out := &Result{
		SkillDir: rep.SkillDir,
		Errors:   rep.Errors,
		Warnings: rep.Warnings,
	}
	for _, r := range rep.Results {
		if r.Level == svtypes.Pass {
			continue
		}
		out.Findings = append(out.Findings, Finding{
			Severity: mapLevel(r.Level),
			Category: r.Category,
			Message:  r.Message,
			File:     r.File,
			Line:     r.Line,
		})
	}
	skillTokens := 0
	total := 0
	for _, tc := range rep.TokenCounts {
		out.TokenCounts = append(out.TokenCounts, TokenCount{File: tc.File, Tokens: tc.Tokens})
		total += tc.Tokens
		if strings.HasPrefix(tc.File, "SKILL.md") {
			skillTokens = tc.Tokens
		}
	}

	out.Analysis = buildAnalysis(rep, total, skillTokens)
	return out
}

// buildAnalysis assembles the offline quality metrics, or returns nil when
// neither content nor contamination analysis ran (structure-only validation).
func buildAnalysis(rep *svtypes.Report, totalTokens, skillTokens int) *Analysis {
	if rep.ContentReport == nil && rep.ContaminationReport == nil {
		if totalTokens == 0 {
			return nil
		}
		return &Analysis{TotalTokens: totalTokens, SkillTokens: skillTokens}
	}

	a := &Analysis{TotalTokens: totalTokens, SkillTokens: skillTokens}
	if c := rep.ContentReport; c != nil {
		a.HasContent = true
		a.WordCount = c.WordCount
		a.CodeBlockRatio = c.CodeBlockRatio
		a.ImperativeRatio = c.ImperativeRatio
		a.InformationDensity = c.InformationDensity
		a.InstructionSpecificity = c.InstructionSpecificity
		a.SectionCount = c.SectionCount
		a.ListItemCount = c.ListItemCount
	}
	if cz := rep.ContaminationReport; cz != nil {
		a.HasContamination = true
		a.ContaminationScore = cz.ContaminationScore
		a.ContaminationLevel = cz.ContaminationLevel
		a.CodeLanguages = cz.CodeLanguages
		a.MismatchedCategories = cz.MismatchedCategories
		a.LanguageMismatch = cz.LanguageMismatch
	}
	return a
}

func mapLevel(l svtypes.Level) Severity {
	switch l {
	case svtypes.Error:
		return SeverityError
	case svtypes.Warning:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}
