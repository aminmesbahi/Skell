package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var f repoFlags

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check for manifest, lock file, and install problems",
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			hasIssues := false
			for _, repo := range repos {
				issues, err := eng.Doctor(repo)
				if err != nil {
					return err
				}
				var entries []output.DiagnosticEntry
				for _, issue := range issues {
					entries = append(entries, output.DiagnosticEntry{
						Severity: string(issue.Severity),
						Code:     issue.Code,
						Message:  issue.Message,
						Hint:     issue.Hint,
					})
					hasIssues = true
				}
				if len(entries) == 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  ok  %s — no issues found\n", repo)
					continue
				}
				p.PrintDiagnostics(entries)
			}
			if hasIssues {
				return fmt.Errorf("doctor found issues — see above")
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	return cmd
}
