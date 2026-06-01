package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/validator"
)

// NamedValidation pairs a skill name with its validation result.
type NamedValidation struct {
	Name   string            `json:"name"`
	Result *validator.Result `json:"result"`
}

// ValidateSkill validates a single installed skill directory against the Agent
// Skills spec (plus any opt-in checks selected in opts).
func (e *Engine) ValidateSkill(ctx context.Context, repoRoot, skillName string, opts validator.Options) (*validator.Result, error) {
	if err := ValidateSkillName(skillName); err != nil {
		return nil, err
	}
	t := ResolveTarget(repoRoot)
	skillDir := filepath.Join(t.SkillsDir(repoRoot), skillName)
	if _, err := os.Stat(skillDir); err != nil {
		return nil, fmt.Errorf("skill %q is not installed in %s", skillName, repoRoot)
	}
	return validator.ValidateDir(ctx, skillDir, opts), nil
}

// ValidateAll validates every installed skill in the repository, returning one
// result per skill in stable name order.
func (e *Engine) ValidateAll(ctx context.Context, repoRoot string, opts validator.Options) ([]NamedValidation, error) {
	installed, err := e.List(repoRoot)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(installed))
	for _, s := range installed {
		names = append(names, s.Name)
	}
	sort.Strings(names)

	t := ResolveTarget(repoRoot)
	var out []NamedValidation
	for _, name := range names {
		skillDir := filepath.Join(t.SkillsDir(repoRoot), name)
		out = append(out, NamedValidation{Name: name, Result: validator.ValidateDir(ctx, skillDir, opts)})
	}
	return out, nil
}

// copySkill copies a skill from reg into destDir. When validation is required
// (policy require-validation, unless overridden), the skill is first copied to a
// temporary staging directory and validated there; an invalid skill is rejected
// without ever overwriting destDir, so a good install is never clobbered by a
// broken upgrade.
func (e *Engine) copySkill(reg registry.Registry, name, version, destDir string) error {
	if !e.requireValidation {
		return e.provider.CopySkillTo(reg, name, version, destDir)
	}

	parent := filepath.Dir(destDir)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(parent, ".skell-validate-")
	if err != nil {
		return fmt.Errorf("failed to create validation staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	stage := filepath.Join(staging, name)
	if err := e.provider.CopySkillTo(reg, name, version, stage); err != nil {
		return err
	}

	res := validator.ValidateDir(context.Background(), stage, validator.Options{})
	if res.HasErrors() {
		return fmt.Errorf(
			"skill %q failed validation (%d error(s)): %s\n  run 'skell validate %s' for details, or pass --no-validate to override",
			name, res.Errors, summarizeFindings(res), name,
		)
	}

	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("failed to clear destination %q: %w", destDir, err)
	}
	if err := os.Rename(stage, destDir); err != nil {
		return fmt.Errorf("failed to move validated skill into place: %w", err)
	}
	return nil
}

// validationIssues runs offline spec validation on a skill directory and maps
// any findings to doctor DiagnosticIssue entries.
func validationIssues(skillDir, name string) []DiagnosticIssue {
	res := validator.ValidateDir(context.Background(), skillDir, validator.Options{})
	var issues []DiagnosticIssue
	for _, f := range res.Findings {
		issues = append(issues, DiagnosticIssue{
			Severity: validationSeverity(f.Severity),
			Code:     "validation:" + f.Category,
			Message:  fmt.Sprintf("skill %q: %s", name, f.Message),
			Hint:     "run 'skell validate " + name + "' for the full report",
		})
	}
	return issues
}

func validationSeverity(s validator.Severity) DiagnosticSeverity {
	switch s {
	case validator.SeverityError:
		return SeverityError
	case validator.SeverityWarning:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// summarizeFindings renders up to three error findings for an error message.
func summarizeFindings(res *validator.Result) string {
	var msgs []string
	for _, f := range res.Findings {
		if f.Severity != validator.SeverityError {
			continue
		}
		msgs = append(msgs, f.Message)
		if len(msgs) == 3 {
			break
		}
	}
	return strings.Join(msgs, "; ")
}
