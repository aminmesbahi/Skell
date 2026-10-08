package skell

import (
	"fmt"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/scaffold"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	var dir, description, owner, repo string
	var asRegistry, withScripts, link bool

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new skill (or a whole skill-source repository)",
		Long: `Scaffolds a spec-compliant skill folder with a starter SKILL.md.

By default the skill is created in a standalone folder (--path, default the
current directory). Pass --link to also link it into the current project so
your agent can use it while you write it.

With --registry, creates a skill-source repository instead: a README, an
example skill under skills/, and a GitHub Actions workflow that validates
every skill on pull requests.`,
		Example: `  # A new skill in ./release-notes
  skell new release-notes --description "Write release notes from merged PRs. Use when preparing a release."

  # Create it in a skills folder and use it in this project right away
  skell new release-notes --path ~/src/my-skills --link

  # Start a team skill repository
  skell new team-skills --registry`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			if asRegistry {
				target := filepath.Join(dir, args[0])
				if err := scaffold.Registry(target, scaffold.RegistryOptions{Name: args[0]}); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(w, "  ✓  created skill repository at %s\n", target)
				_, _ = fmt.Fprintln(w, "     next: git init, push it, then teammates run 'skell add <owner>/<repo>'")
				return nil
			}
			path, err := scaffold.Skill(dir, scaffold.SkillOptions{
				Name: args[0], Description: description, Owner: owner, WithScripts: withScripts,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(w, "  ✓  created %s\n", filepath.Join(path, "SKILL.md"))
			if link {
				root, err := resolveRepo(repo)
				if err != nil {
					return err
				}
				res, err := engine.New(defaultCacheRoot()).Link(root, path, "")
				if err != nil {
					return fmt.Errorf("skill created, but linking failed: %w", err)
				}
				_, _ = fmt.Fprintf(w, "  ✓  linked into %s (edits are live)\n", res.Path)
			}
			_, _ = fmt.Fprintf(w, "     check it with: skell validate --path %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "path", ".", "Parent folder to create the skill (or repository) in")
	cmd.Flags().StringVar(&description, "description", "", "What the skill does and when to use it")
	cmd.Flags().StringVar(&owner, "owner", "", "Owning team or person (metadata.owner)")
	cmd.Flags().BoolVar(&withScripts, "scripts", false, "Also create scripts/ and references/ folders")
	cmd.Flags().BoolVar(&link, "link", false, "Link the new skill into the current project (see 'skell link')")
	cmd.Flags().StringVar(&repo, "repo", "", "Project to link into with --link (defaults to current directory)")
	cmd.Flags().BoolVar(&asRegistry, "registry", false, "Create a skill-source repository instead of a single skill")
	return cmd
}
