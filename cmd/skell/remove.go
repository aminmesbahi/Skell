package skell

import (
	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var f repoFlags

	cmd := &cobra.Command{
		Use:   "remove <skill-name>",
		Short: "Remove a skill from one or more repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			for _, repo := range repos {
				if err := eng.Remove(repo, args[0], f.dryRun); err != nil {
					return err
				}
				p.PrintAction(output.ActionEvent{
					Action: "remove", Skill: args[0], Repo: repo, DryRun: f.dryRun,
				})
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	return cmd
}
