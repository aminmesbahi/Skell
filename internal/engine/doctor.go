package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
)

// Doctor runs all diagnostic checks on a repository.
func (e *Engine) Doctor(repoRoot string) ([]DiagnosticIssue, error) {
	var issues []DiagnosticIssue

	// 1. Manifest
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		issues = append(issues, DiagnosticIssue{
			Severity: SeverityError,
			Code:     "no-manifest",
			Message:  "no manifest (skell.toml) found",
			Hint:     "run 'skell init' to create one",
		})
		return issues, nil
	}

	// 2. Lock file
	lockPath := lockfile.PathFor(repoRoot, *t)
	lf, err := lockfile.Read(lockPath)
	if err != nil {
		issues = append(issues, DiagnosticIssue{
			Severity: SeverityWarning,
			Code:     "no-lockfile",
			Message:  "no lock file found (skell.lock)",
			Hint:     "run 'skell sync' to create one",
		})
		lf = &lockfile.LockFile{}
	}

	// 3. Per-skill checks
	skillsDir := t.SkillsDir(repoRoot)
	for _, locked := range lf.Skills {
		skillDir := filepath.Join(skillsDir, locked.Name)

		// Directory present?
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			issues = append(issues, DiagnosticIssue{
				Severity: SeverityError,
				Code:     "missing-dir",
				Message:  fmt.Sprintf("skill %q is in lock file but directory is missing", locked.Name),
				Hint:     fmt.Sprintf("run 'skell install %s' to reinstall", locked.Name),
			})
			continue
		}

		// SKILL.md valid per the Agent Skills spec? This goes beyond "parseable"
		// — it enforces required frontmatter, directory layout, token budgets,
		// and code-fence integrity via the validator.
		issues = append(issues, validationIssues(skillDir, locked.Name)...)

		// Content hash matches?
		if locked.ContentHash != "" {
			ok, err := hasher.Verify(skillDir, locked.ContentHash)
			if err != nil {
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning,
					Code:     "hash-error",
					Message:  fmt.Sprintf("skill %q: could not verify content hash: %v", locked.Name, err),
				})
			} else if !ok {
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning,
					Code:     "locally-modified",
					Message:  fmt.Sprintf("skill %q: content hash mismatch (locally modified)", locked.Name),
					Hint:     fmt.Sprintf("review your edits with 'skell diff %s'; discard them with 'skell upgrade %s --force'", locked.Name, locked.Name),
				})
			}
		}
	}

	// 4. Installed skills not in manifest
	if _, err := os.Stat(skillsDir); err == nil {
		entries, _ := os.ReadDir(skillsDir)
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".skell-") {
				continue
			}
			if _, ok := m.Skills[entry.Name()]; !ok {
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning,
					Code:     "untracked-skill",
					Message:  fmt.Sprintf("skill %q is installed but not in manifest", entry.Name()),
					Hint:     "run 'skell sync' to reconcile",
				})
			}
		}
	}

	// 5. Mirror targets in sync with the primary install?
	for _, s := range staleMirrors(repoRoot, m, *t, lf) {
		issues = append(issues, DiagnosticIssue{
			Severity: SeverityWarning,
			Code:     "stale-mirror",
			Message:  fmt.Sprintf("mirror copy %s is missing or out of date", s),
			Hint:     "run 'skell sync' (or 'skell doctor --fix') to refresh mirrors",
		})
	}

	return issues, nil
}
