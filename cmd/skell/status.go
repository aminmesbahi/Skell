package skell

import (
	"fmt"
	"io"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var f repoFlags
	var only string
	var targetID string
	var refresh bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show comparison between registry and local installs",
		Long: `Compares every installed skill against its source to show whether each is
up-to-date, outdated, pinned, linked, locally-modified, or has missing metadata.

A skill is "outdated" when its source has a newer version — or when the source
files changed even though the version did not (many skills have no version).
Status compares against the locally cached copy of each source; pass
--refresh to fetch the latest first.`,
		Example: `  # Check status of all skills in the current repo
  skell status

  # Fetch the latest from every source first
  skell status --refresh

  # Show only outdated skills
  skell status --only outdated

  # Status across multiple repos
  skell status --repo ./service-a --repo ./service-b

  # Machine-readable JSON output
  skell status --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			w := cmd.OutOrStdout()
			for _, repo := range repos {
				if refresh {
					if m, err := manifest.Resolve(repo); err == nil {
						if err := eng.CacheRefresh(m); err != nil {
							return err
						}
					}
				}
				entries, err := eng.StatusFor(repo, targetID)
				if err != nil {
					return err
				}
				if len(repos) > 1 && !f.jsonOut {
					_, _ = fmt.Fprintf(w, "\n  %s\n", repo)
				}
				filtered := filterStatus(entries, only)
				if len(filtered) == 0 {
					if !f.jsonOut {
						_, _ = fmt.Fprintln(w, "  no skills to show")
					}
					continue
				}
				p.PrintStatusTable(filtered)
				if !f.jsonOut {
					printStatusHints(w, filtered)
				}
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&only, "only", "", "Filter by status (e.g. outdated, locally-modified)")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform to check status for: claude | codex | copilot | cursor | windsurf | opencode | cline | grok")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Fetch the latest from every source before comparing")
	return cmd
}

// printStatusHints suggests the next command for anything that needs action.
func printStatusHints(w io.Writer, entries []model.StatusEntry) {
	counts := map[model.SkillStatus]int{}
	first := map[model.SkillStatus]string{}
	for _, e := range entries {
		counts[e.Status]++
		if first[e.Status] == "" {
			first[e.Status] = e.Name
		}
	}
	var lines []string
	if n := counts[model.StatusOutdated]; n > 0 {
		lines = append(lines, fmt.Sprintf("%d outdated — review with 'skell diff %s', apply with 'skell upgrade'", n, first[model.StatusOutdated]))
	}
	if n := counts[model.StatusLocallyModified]; n > 0 {
		lines = append(lines, fmt.Sprintf("%d locally modified — see your edits with 'skell diff %s'", n, first[model.StatusLocallyModified]))
	}
	if n := counts[model.StatusDeprecated] + counts[model.StatusArchived]; n > 0 {
		lines = append(lines, fmt.Sprintf("%d deprecated/archived — consider 'skell remove' or a replacement ('skell search')", n))
	}
	if n := counts[model.StatusUnknown]; n > 0 {
		lines = append(lines, fmt.Sprintf("%d unknown — run 'skell doctor' for details", n))
	}
	if len(lines) == 0 {
		_, _ = fmt.Fprintln(w, "\n  ✓  everything is up to date")
		return
	}
	_, _ = fmt.Fprintln(w)
	for _, l := range lines {
		_, _ = fmt.Fprintf(w, "  →  %s\n", l)
	}
}

func filterStatus(entries []model.StatusEntry, only string) []model.StatusEntry {
	if only == "" {
		return entries
	}
	var out []model.StatusEntry
	for _, e := range entries {
		if string(e.Status) == only {
			out = append(out, e)
		}
	}
	return out
}
