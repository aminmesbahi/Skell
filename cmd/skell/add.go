package skell

import (
	"encoding/json"
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var f repoFlags

	cmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add a skill (or skill registry) directly from a GitHub URL",
		Long: `Parses a GitHub tree URL and either:

  • Installs a specific skill when the URL points to a skill directory
    (e.g. https://github.com/owner/repo/tree/main/skills/my-skill)

  • Registers a skill registry when the URL points to a skills root
    (e.g. https://github.com/owner/repo/tree/main/skills)

  • Registers a plain repository as a registry
    (e.g. https://github.com/owner/repo)

The registry alias is derived from the repository name.
Updates skell.toml (and skell.lock for skill installs).`,
		Example: `  # Install a specific skill
  skell add https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills/akka-testing-patterns

  # Register a skill registry
  skell add https://github.com/Aaronontheweb/dotnet-skills/tree/master/skills

  # Register a plain repo as a registry
  skell add https://github.com/davidfowl/dotnet-skillz

  # Dry-run (preview without writing)
  skell add https://github.com/owner/repo/tree/main/skills/my-skill --dry-run

  # Operate on a specific repo
  skell add https://github.com/owner/repo/tree/main/skills/my-skill --repo /path/to/project`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawURL := args[0]

			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())

			type jsonResult struct {
				Repo       string `json:"repo"`
				Alias      string `json:"alias"`
				SkillName  string `json:"skill_name,omitempty"`
				Registered bool   `json:"registered"`
				Installed  bool   `json:"installed"`
				DryRun     bool   `json:"dry_run"`
			}

			var results []jsonResult

			for _, repo := range repos {
				res, err := eng.AddFromURL(repo, rawURL, f.dryRun)
				if err != nil {
					return fmt.Errorf("%s: %w", repo, err)
				}

				if f.jsonOut {
					results = append(results, jsonResult{
						Repo:       repo,
						Alias:      res.Alias,
						SkillName:  res.SkillName,
						Registered: res.Registered,
						Installed:  res.Installed,
						DryRun:     f.dryRun,
					})
					continue
				}

				switch {
				case res.Installed:
					fmt.Fprintf(cmd.OutOrStdout(), "installed skill %q from registry %q into %s\n",
						res.SkillName, res.Alias, repo)
				case res.Registered:
					fmt.Fprintf(cmd.OutOrStdout(), "registered registry %q in %s\n", res.Alias, repo)
				case res.SkillName != "":
					fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would install skill %q from registry %q into %s\n",
						res.SkillName, res.Alias, repo)
				default:
					fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would register registry %q in %s\n", res.Alias, repo)
				}
			}

			if f.jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	return cmd
}
