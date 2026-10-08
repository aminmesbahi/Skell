package skell

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	var f repoFlags
	var noSync bool

	cmd := &cobra.Command{
		Use:   "apply <baseline>",
		Short: "Adopt a shared team baseline (sources + skills) in one step",
		Long: `Merges a shared skell.toml — your team's or organisation's standard set of
skill sources, skills and mirror targets — into each project, then syncs.

<baseline> is a path to a skell.toml or an https:// URL to one (for example a
raw file in a team repository). Entries the project already has are kept; a
source alias that points somewhere else in the baseline is an error, so a
baseline can never silently swap where a project's skills come from.`,
		Example: `  # Adopt the team baseline
  skell apply https://raw.githubusercontent.com/acme/skills/main/baseline.toml

  # Preview, then apply to every repo under ~/work
  skell apply ./baseline.toml --dry-run
  skell apply ./baseline.toml --all-repos ~/work`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := loadBaseline(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			w := cmd.OutOrStdout()
			type result struct {
				Repo   string              `json:"repo"`
				Report *engine.ApplyReport `json:"report"`
				Synced *engine.SyncReport  `json:"synced,omitempty"`
			}
			var results []result
			for _, repo := range repos {
				rep, err := eng.ApplyBaseline(repo, base, f.dryRun)
				if err != nil {
					return fmt.Errorf("%s: %w", repo, err)
				}
				r := result{Repo: repo, Report: rep}
				if !f.dryRun && !noSync {
					if r.Synced, err = eng.Sync(repo, false, false, false); err != nil {
						return fmt.Errorf("%s: %w", repo, err)
					}
				}
				results = append(results, r)
				if f.jsonOut {
					continue
				}
				verb := "added"
				if f.dryRun {
					verb = "would add"
				}
				_, _ = fmt.Fprintf(w, "  %s\n", repo)
				printList(w, verb+" sources", rep.AddedRegistries)
				printList(w, verb+" skills", rep.AddedSkills)
				printList(w, verb+" mirrors", rep.AddedMirrors)
				if r.Synced != nil {
					printList(w, "installed", r.Synced.Installed)
				}
				if len(rep.AddedRegistries)+len(rep.AddedSkills)+len(rep.AddedMirrors) == 0 {
					_, _ = fmt.Fprintln(w, "    ✓  already matches the baseline")
				}
			}
			if f.jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}
			return nil
		},
	}
	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&noSync, "no-sync", false, "Only update skell.toml; don't install")
	return cmd
}

func printList(w io.Writer, label string, items []string) {
	if len(items) > 0 {
		_, _ = fmt.Fprintf(w, "    %-18s %s\n", label+":", strings.Join(items, ", "))
	}
}

// loadBaseline reads a baseline manifest from a local path or an https URL.
func loadBaseline(ctx context.Context, src string) (*manifest.Manifest, error) {
	if strings.Contains(src, "://") {
		u, err := url.Parse(src)
		if err != nil || u.Scheme != "https" {
			return nil, fmt.Errorf("baseline URL must use https: %s", src)
		}
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "skell-cli")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("download baseline: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("download baseline: HTTP %d", resp.StatusCode)
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return nil, err
		}
		m, err := manifest.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("baseline is not a valid skell.toml: %w", err)
		}
		return m, nil
	}
	data, err := os.ReadFile(src) //nolint:gosec
	if err != nil {
		return nil, err
	}
	m, err := manifest.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("baseline is not a valid skell.toml: %w", err)
	}
	return m, nil
}
