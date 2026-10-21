package skell

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newBundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Named sets of skills (a role or team starter pack) you install in one step",
		Long: `A bundle is a named list of skills stored in the project's skell.toml:

  [bundles.frontend]
  description = "Everything a UI developer needs"
  skills = ["react", "css-review", "a11y-check"]

Install one with 'skell bundle install frontend' (or 'skell install @frontend').
Bundles travel with 'skell apply <baseline.toml>', so a team can publish its
starter packs once and every project picks them up. Dependencies declared by
the skills (metadata.requires) are installed along with them.`,
	}
	cmd.AddCommand(newBundleListCmd(), newBundleCreateCmd(), newBundleInstallCmd(), newBundleDeleteCmd())
	return cmd
}

func newBundleListCmd() *cobra.Command {
	var repo string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show the bundles defined in this project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			list, err := engine.New(defaultCacheRoot()).ListBundles(root)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(w).Encode(list)
			}
			if len(list) == 0 {
				_, _ = fmt.Fprintln(w, "  no bundles yet — create one with 'skell bundle create <name> <skill>...'")
				return nil
			}
			for _, b := range list {
				_, _ = fmt.Fprintf(w, "\n  @%s  %d/%d installed\n", b.Name, b.Installed, len(b.Skills))
				if b.Description != "" {
					_, _ = fmt.Fprintf(w, "     %s\n", b.Description)
				}
				_, _ = fmt.Fprintf(w, "     %s\n", strings.Join(b.Skills, ", "))
			}
			_, _ = fmt.Fprintln(w)
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Project (defaults to the current directory)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newBundleCreateCmd() *cobra.Command {
	var repo, description string
	cmd := &cobra.Command{
		Use:   "create <name> <skill>...",
		Short: "Create (or replace) a bundle",
		Example: `  skell bundle create frontend react css-review a11y-check --description "UI developer pack"
  skell bundle install frontend`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			if err := engine.New(defaultCacheRoot()).SaveBundle(root, args[0], description, args[1:]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s  bundle @%s saved (%d skills) — install it with 'skell bundle install %s'\n",
				markOK, args[0], len(args)-1, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Project (defaults to the current directory)")
	cmd.Flags().StringVar(&description, "description", "", "What the bundle is for")
	return cmd
}

func newBundleDeleteCmd() *cobra.Command {
	var repo string
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a bundle (its skills stay installed)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			if err := engine.New(defaultCacheRoot()).DeleteBundle(root, strings.TrimPrefix(args[0], "@")); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s  bundle %s deleted\n", markOK, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Project (defaults to the current directory)")
	return cmd
}

func newBundleInstallCmd() *cobra.Command {
	var f repoFlags
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "install <name>...",
		Short: "Install every skill in one or more bundles",
		Long:  "Same as 'skell install @name', with all of install's checks: dependencies are added, and skills that ship scripts or broad tool access are reviewed first.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append([]string{}, args...)
			for i, a := range args {
				if !strings.HasPrefix(a, "@") {
					args[i] = "@" + a
				}
			}
			install := newInstallCmd()
			install.SetOut(cmd.OutOrStdout())
			install.SetErr(cmd.ErrOrStderr())
			var flags []string
			for _, r := range f.repo {
				flags = append(flags, "--repo", r)
			}
			if f.allRepos != "" {
				flags = append(flags, "--all-repos", f.allRepos)
			}
			if f.dryRun {
				flags = append(flags, "--dry-run")
			}
			if f.jsonOut {
				flags = append(flags, "--json")
			}
			if f.user {
				flags = append(flags, "--user")
			}
			if assumeYes {
				flags = append(flags, "--yes")
			}
			install.SetArgs(append(flags, args...))
			return install.Execute()
		},
	}
	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Don't ask for confirmation before installing")
	return cmd
}
