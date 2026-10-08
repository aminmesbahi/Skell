package skell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/mcp"
	"github.com/aminmesbahi/skell/internal/version"
	"github.com/spf13/cobra"
)

const mcpInstructions = `Skell manages Agent Skills (SKILL.md folders) for this project.
Typical flow: skell_catalog or skell_search to find skills, skell_review to inspect one,
skell_install to add it, skell_status to check for updates. Install returns
needs_confirmation=true for skills that ship scripts or pre-approve broad tools: show
the review to the user and call again with confirm=true only if they agree.
All tools default to the current working directory as the project ("repo").`

func newMCPCmd() *cobra.Command {
	var repo string
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run Skell as an MCP server so AI agents can find and install skills",
		Long: `Starts a Model Context Protocol server on stdio. Add it to your agent so it
can browse, review, install and update skills itself ("find me a skill for PDF
forms and install it").

Claude Code:   claude mcp add skell -- skell mcp
Other clients: command "skell", args ["mcp"]

The server works on the project in the current directory (or --repo). Installs
of skills that ship scripts or pre-approve broad tool access are not performed
until the agent calls again with confirm=true, so you see the review first.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			srv := mcp.NewServer("skell", version.Version, mcpInstructions, mcpTools(root))
			return srv.Serve(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Project the server operates on (defaults to current directory)")
	return cmd
}

// mcpArgs is the union of all tool arguments; each tool reads what it needs.
type mcpArgs struct {
	Repo     string   `json:"repo"`
	Query    string   `json:"query"`
	Name     string   `json:"name"`
	Skills   []string `json:"skills"`
	Registry string   `json:"registry"`
	Source   string   `json:"source"`
	Alias    string   `json:"alias"`
	Confirm  bool     `json:"confirm"`
	Refresh  bool     `json:"refresh"`
}

func mcpTools(defaultRepo string) []mcp.Tool {
	eng := func() *engine.Engine { return engine.New(defaultCacheRoot()) }
	parse := func(raw json.RawMessage) (mcpArgs, string, error) {
		var a mcpArgs
		if err := json.Unmarshal(raw, &a); err != nil {
			return a, "", fmt.Errorf("invalid arguments: %w", err)
		}
		repo := a.Repo
		if repo == "" {
			repo = defaultRepo
		}
		return a, repo, nil
	}
	repoProp := mcp.Str("Project directory (defaults to the server's working directory)")
	ro := &mcp.Annotations{ReadOnlyHint: true, OpenWorldHint: true}

	return []mcp.Tool{
		{
			Name: "skell_catalog", Title: "Browse skill sources",
			Description: "List well-known skill sources (git repositories of skills) that can be added with skell_add_source.",
			InputSchema: mcp.Object(map[string]any{"query": mcp.Str("Optional keyword filter")}),
			Annotations: ro,
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, _, err := parse(raw)
				if err != nil {
					return nil, err
				}
				return loadCatalog(false).Search(a.Query), nil
			},
		},
		{
			Name: "skell_search", Title: "Search skills",
			Description: "Search skills offered by the project's configured sources (and global sources) by keyword in name, description and tags.",
			InputSchema: mcp.Object(map[string]any{"query": mcp.Str("Keyword; empty lists everything"), "repo": repoProp}),
			Annotations: ro,
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				skills, err := eng().SearchMerged(repo, a.Query, "", "", "")
				if err != nil {
					return nil, err
				}
				if len(skills) == 0 {
					return "No skills found. Add a source first (skell_catalog, then skell_add_source).", nil
				}
				return skills, nil
			},
		},
		{
			Name: "skell_review", Title: "Review a skill before installing",
			Description: "Show what installing a skill would add: files, scripts, allowed tools, external links and warnings.",
			InputSchema: mcp.Object(map[string]any{"name": mcp.Str("Skill name"), "registry": mcp.Str("Source alias (optional; found automatically)"), "repo": repoProp}, "name"),
			Annotations: ro,
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				return eng().ReviewSkill(repo, a.Name, a.Registry, "")
			},
		},
		{
			Name: "skell_list_installed", Title: "List installed skills",
			Description: "List skills installed in the project.",
			InputSchema: mcp.Object(map[string]any{"repo": repoProp}),
			Annotations: &mcp.Annotations{ReadOnlyHint: true},
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				_, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				return eng().ListFor(repo, "")
			},
		},
		{
			Name: "skell_status", Title: "Check for updates",
			Description: "Compare installed skills with their sources: up-to-date, outdated (newer version or changed content), pinned, locally-modified…",
			InputSchema: mcp.Object(map[string]any{"refresh": mcp.Bool("Fetch the latest from every source first"), "repo": repoProp}),
			Annotations: ro,
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				e := eng()
				if a.Refresh {
					if m, err := manifest.Resolve(repo); err == nil {
						if err := e.CacheRefresh(m); err != nil {
							return nil, err
						}
					}
				}
				return e.StatusFor(repo, "")
			},
		},
		{
			Name: "skell_diff", Title: "Diff an installed skill",
			Description: "Unified diff from the installed copy of a skill to the current source version.",
			InputSchema: mcp.Object(map[string]any{"name": mcp.Str("Installed skill name"), "repo": repoProp}, "name"),
			Annotations: ro,
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				return eng().Diff(repo, a.Name, "", false)
			},
		},
		{
			Name: "skell_add_source", Title: "Add a skill source",
			Description: "Add a skill source to the project: a catalog id (e.g. \"anthropic\"), GitHub owner/repo, or git URL. If it points at a single skill, that skill is installed.",
			InputSchema: mcp.Object(map[string]any{"source": mcp.Str("Catalog id, owner/repo, owner/repo/path/to/skill or URL"), "alias": mcp.Str("Optional alias"), "repo": repoProp}, "source"),
			Annotations: &mcp.Annotations{OpenWorldHint: true, IdempotentHint: true},
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				expanded, fromCatalog := engine.ExpandAddArg(a.Source, catalogLookup)
				alias := a.Alias
				if fromCatalog && alias == "" {
					alias = a.Source
				}
				e := eng()
				parsed, err := e.ResolveAdd(repo, expanded, alias)
				if err != nil {
					return nil, err
				}
				if parsed.SkillName != "" && !a.Confirm {
					if rev, err := e.ReviewSkill(repo, parsed.SkillName, parsed.Alias, parsed.GitURL); err == nil && rev.NeedsConfirmation() {
						return map[string]any{"needs_confirmation": true, "review": rev,
							"message": "This skill ships scripts or pre-approves broad tool access. Show the review to the user and call again with confirm=true if they agree."}, nil
					}
				}
				return e.AddFromURLWith(repo, expanded, engine.AddOptions{Alias: parsed.Alias})
			},
		},
		{
			Name: "skell_install", Title: "Install skills",
			Description: "Install one or more skills into the project from its configured sources. Skills with scripts or broad tool permissions require confirm=true after the user has seen skell_review.",
			InputSchema: mcp.Object(map[string]any{
				"skills":   mcp.StrList("Skill names to install"),
				"registry": mcp.Str("Source alias (optional; found automatically)"),
				"confirm":  mcp.Bool("Set true after the user approved the review of a risky skill"),
				"repo":     repoProp,
			}, "skills"),
			Annotations: &mcp.Annotations{OpenWorldHint: true},
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				if len(a.Skills) == 0 {
					return nil, errors.New("skills is required")
				}
				e := eng()
				var pending []*engine.Review
				for _, s := range a.Skills {
					rev, err := e.ReviewSkill(repo, s, a.Registry, "")
					if err != nil {
						return nil, err
					}
					if rev.NeedsConfirmation() && !a.Confirm {
						pending = append(pending, rev)
					}
				}
				if len(pending) > 0 {
					return map[string]any{"needs_confirmation": true, "reviews": pending,
						"message": "Nothing installed. These skills ship scripts or pre-approve broad tool access. Show the reviews to the user and call again with confirm=true if they agree."}, nil
				}
				var installed []string
				for _, s := range a.Skills {
					if err := e.InstallTo(repo, s, a.Registry, "", "", false); err != nil {
						return nil, fmt.Errorf("installed %v, then failed on %s: %w", installed, s, err)
					}
					installed = append(installed, s)
				}
				return map[string]any{"installed": installed, "message": "Installed " + strings.Join(installed, ", ") + ". The agent may need to reload to pick up new skills."}, nil
			},
		},
		{
			Name: "skell_upgrade", Title: "Upgrade skills",
			Description: "Upgrade one skill (name) or all non-pinned skills to the latest source version. Locally modified skills are not overwritten.",
			InputSchema: mcp.Object(map[string]any{"name": mcp.Str("Skill to upgrade; omit for all"), "repo": repoProp}),
			Annotations: &mcp.Annotations{OpenWorldHint: true},
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				return eng().UpgradeFor(repo, a.Name, "", false, false)
			},
		},
		{
			Name: "skell_remove", Title: "Remove a skill",
			Description: "Remove an installed skill from the project.",
			InputSchema: mcp.Object(map[string]any{"name": mcp.Str("Skill name"), "repo": repoProp}, "name"),
			Annotations: &mcp.Annotations{DestructiveHint: true},
			Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
				a, repo, err := parse(raw)
				if err != nil {
					return nil, err
				}
				if err := eng().RemoveFor(repo, a.Name, "", false); err != nil {
					return nil, err
				}
				return "Removed " + a.Name, nil
			},
		},
	}
}
