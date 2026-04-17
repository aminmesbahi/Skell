package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
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
			w := cmd.OutOrStdout()
			hasIssues := false
			for _, repo := range repos {
				issues, err := eng.Doctor(repo)
				if err != nil {
					return err
				}
				if len(issues) == 0 {
					_, _ = fmt.Fprintf(w, "  ok  %s — no issues found\n", repo)
					continue
				}
				for _, issue := range issues {
					hasIssues = true
					_, _ = fmt.Fprintf(w, "  [%s]  %s\n", issue.Severity, issue.Message)
					if issue.Hint != "" {
						_, _ = fmt.Fprintf(w, "         hint: %s\n", issue.Hint)
					}
				}
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
