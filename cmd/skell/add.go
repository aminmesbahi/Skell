package skell

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var f repoFlags
	var alias string
	var assumeYes bool

	cmd := &cobra.Command{
		Use:   "add <source>...",
		Short: "Add a skill source (or a single skill) by catalog id, owner/repo, URL or path",
		Long: `Adds a skill source to skell.toml, or installs a single skill from it.

<source> can be:

  • a catalog id                 skell add anthropic          (see 'skell catalog')
  • a GitHub owner/repo          skell add davidfowl/dotnet-skillz
  • one skill in a GitHub repo   skell add anthropics/skills/skills/pdf
  • a full git or GitHub URL     skell add https://github.com/owner/repo/tree/main/skills/my-skill
  • a local folder               skell add ./my-skills

When the source points at a single skill it is installed (after a short review
if it ships scripts or asks for broad tool access); otherwise the source is
registered so 'skell search' and 'skell install' can use it.

The registry alias is the repository name (or the catalog id); if that name is
already taken by another source, the owner is prefixed (e.g. "openai-skills").
Use --alias to choose it yourself.`,
		Example: `  # Add a well-known source and browse it
  skell add anthropic
  skell search pdf

  # Add a GitHub repo by owner/repo
  skell add davidfowl/dotnet-skillz

  # Install one skill directly
  skell add anthropics/skills/skills/pdf

  # Full URL, explicit alias, preview only
  skell add https://github.com/owner/repo --alias team --dry-run`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completeCatalogIDs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if alias != "" && len(args) > 1 {
				return errors.New("--alias can only be used with a single source")
			}
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			w := cmd.OutOrStdout()

			type jsonResult struct {
				Repo       string `json:"repo"`
				Source     string `json:"source"`
				Alias      string `json:"alias"`
				URL        string `json:"url"`
				SkillName  string `json:"skill_name,omitempty"`
				Registered bool   `json:"registered"`
				Installed  bool   `json:"installed"`
				Skipped    bool   `json:"skipped,omitempty"`
				DryRun     bool   `json:"dry_run"`
			}
			var results []jsonResult
			registered := false

			for _, arg := range args {
				expanded, fromCatalog := engine.ExpandAddArg(arg, catalogLookup)
				useAlias := alias
				if fromCatalog && useAlias == "" {
					useAlias = arg
				}
				for _, repo := range repos {
					parsed, err := eng.ResolveAdd(repo, expanded, useAlias)
					if err != nil {
						return fmt.Errorf("%s: %w", repo, err)
					}
					if parsed.SkillName != "" && !f.jsonOut {
						// A failed review (e.g. the URL is a skills folder, not
						// one skill) falls through to AddFromURLWith, which
						// reports the authoritative result.
						rev, err := eng.ReviewSkill(repo, parsed.SkillName, parsed.Alias, parsed.GitURL)
						if err == nil && !reviewAndConfirm(cmd, rev, assumeYes || f.dryRun, f.dryRun) {
							_, _ = fmt.Fprintf(w, "  skipped  %s\n", parsed.SkillName)
							results = append(results, jsonResult{Repo: repo, Source: arg, Alias: parsed.Alias, SkillName: parsed.SkillName, Skipped: true})
							continue
						}
					}
					res, err := eng.AddFromURLWith(repo, expanded, engine.AddOptions{Alias: parsed.Alias, DryRun: f.dryRun})
					if err != nil {
						return fmt.Errorf("%s: %w", repo, err)
					}
					results = append(results, jsonResult{
						Repo: repo, Source: arg, Alias: res.Alias, URL: res.URL, SkillName: res.SkillName,
						Registered: res.Registered, Installed: res.Installed, DryRun: f.dryRun,
					})
					if f.jsonOut {
						continue
					}
					switch {
					case res.Installed:
						_, _ = fmt.Fprintf(w, "  ✓  installed %q from %q into %s\n", res.SkillName, res.Alias, repo)
					case res.SkillName != "" && f.dryRun:
						_, _ = fmt.Fprintf(w, "  dry-run  would install %q from %q into %s\n", res.SkillName, res.Alias, repo)
					case f.dryRun:
						_, _ = fmt.Fprintf(w, "  dry-run  would add source %q (%s) to %s\n", res.Alias, res.URL, repo)
					case res.Registered:
						registered = true
						_, _ = fmt.Fprintf(w, "  ✓  added source %q (%s) to %s\n", res.Alias, res.URL, repo)
					default:
						_, _ = fmt.Fprintf(w, "  ok  source %q is already configured in %s\n", res.Alias, repo)
					}
				}
			}

			if f.jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}
			if registered {
				_, _ = fmt.Fprintln(w, "\n  next: 'skell search' to browse its skills, then 'skell install <skill>'")
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&alias, "alias", "", "Registry alias to use instead of the derived one")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Don't ask for confirmation before installing")
	return cmd
}
