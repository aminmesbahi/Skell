package skell

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/validator"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var f repoFlags
	var full, links, strict bool

	cmd := &cobra.Command{
		Use:   "validate [skill-name]",
		Short: "Validate installed skills against the Agent Skills spec",
		Long: `Validates one or all installed skills using the Agent Skills validator.

By default only offline structure checks run (required frontmatter, directory
layout, token budgets, code-fence integrity, internal links). Use --full to add
offline content-quality and contamination analysis, and --links to additionally
check external links over the network.

Exits non-zero if any skill has errors (or, with --strict, any warnings).`,
		Example: `  # Validate every installed skill in the current repo
  skell validate

  # Validate a single skill
  skell validate pdf-processing

  # Full offline analysis, failing on warnings too
  skell validate --full --strict

  # Also verify external links resolve
  skell validate --links

  # JSON output for CI
  skell validate --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			var skillName string
			if len(args) == 1 {
				skillName = args[0]
			}

			eng := engine.New(defaultCacheRoot())
			opts := validator.Options{Content: full, Contamination: full, Links: links}
			w := cmd.OutOrStdout()

			anyError, anyWarning := false, false
			for _, repo := range repos {
				results, err := collectValidations(cmd.Context(), eng, repo, skillName, opts)
				if err != nil {
					return err
				}
				for _, nv := range results {
					if nv.Result.HasErrors() {
						anyError = true
					}
					if nv.Result.HasWarnings() {
						anyWarning = true
					}
				}
				if f.jsonOut {
					out, _ := json.Marshal(results)
					_, _ = fmt.Fprintf(w, "%s\n", out)
					continue
				}
				printValidations(w, repo, results)
			}

			if anyError || (strict && anyWarning) {
				return fmt.Errorf("validation failed")
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&full, "full", false, "Also run offline content-quality and contamination analysis")
	cmd.Flags().BoolVar(&links, "links", false, "Also validate external links (network access)")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat warnings as failures (non-zero exit)")
	return cmd
}

func collectValidations(ctx context.Context, eng *engine.Engine, repo, skillName string, opts validator.Options) ([]engine.NamedValidation, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if skillName != "" {
		res, err := eng.ValidateSkill(ctx, repo, skillName, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", repo, err)
		}
		return []engine.NamedValidation{{Name: skillName, Result: res}}, nil
	}
	results, err := eng.ValidateAll(ctx, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", repo, err)
	}
	return results, nil
}

func printValidations(w interface{ Write([]byte) (int, error) }, repo string, results []engine.NamedValidation) {
	_, _ = fmt.Fprintf(w, "  validate  %s\n", repo)
	if len(results) == 0 {
		_, _ = fmt.Fprintln(w, "  (no installed skills)")
		return
	}
	for _, nv := range results {
		r := nv.Result
		switch {
		case r.HasErrors():
			_, _ = fmt.Fprintf(w, "  ✗ %s — %d error(s), %d warning(s)\n", nv.Name, r.Errors, r.Warnings)
		case r.HasWarnings():
			_, _ = fmt.Fprintf(w, "  ! %s — %d warning(s)\n", nv.Name, r.Warnings)
		default:
			_, _ = fmt.Fprintf(w, "  ✓ %s — passed\n", nv.Name)
		}
		for _, finding := range r.Findings {
			loc := finding.File
			if loc != "" && finding.Line > 0 {
				loc = fmt.Sprintf("%s:%d", finding.File, finding.Line)
			}
			if loc != "" {
				loc = " (" + loc + ")"
			}
			_, _ = fmt.Fprintf(w, "      %-7s [%s] %s%s\n", finding.Severity, finding.Category, finding.Message, loc)
		}
	}
}
