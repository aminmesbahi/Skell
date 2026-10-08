package skell

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/spf13/cobra"
)

func newMirrorCmd() *cobra.Command {
	var repo string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "mirror",
		Short: "Share one manifest's skills with other agents in the same repo",
		Long: `Many repositories are used with several AI agents at once (for example Claude
Code, GitHub Copilot and Cursor). Mirrors let one skell.toml serve them all:
the manifest's own agent folder holds the real skills, and every mirror
target's skills/ folder receives an identical copy, kept up to date by
install, upgrade, remove and sync.

Mirror copies are derived files — 'skell sync' rebuilds them and
'skell doctor' reports when they drift.`,
		Example: `  # Also provide the skills to Copilot and Cursor
  skell mirror add copilot cursor

  # See what is mirrored
  skell mirror list

  # Stop mirroring to Cursor (removes the copies Skell placed there)
  skell mirror remove cursor`,
	}

	print := func(cmd *cobra.Command, rep *engine.MirrorReport) error {
		w := cmd.OutOrStdout()
		if jsonOut {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			return enc.Encode(rep)
		}
		if len(rep.Targets) == 0 {
			_, _ = fmt.Fprintln(w, "  no mirrors configured — add one with 'skell mirror add <target>'")
			return nil
		}
		_, _ = fmt.Fprintf(w, "  mirrors: %s\n", strings.Join(rep.Targets, ", "))
		if len(rep.Copied) > 0 {
			_, _ = fmt.Fprintf(w, "  ✓  %d skill(s) mirrored: %s\n", len(rep.Copied), strings.Join(rep.Copied, ", "))
		}
		for _, r := range rep.Removed {
			_, _ = fmt.Fprintf(w, "  removed  %s\n", r)
		}
		return nil
	}

	add := &cobra.Command{
		Use:               "add <target>...",
		Short:             "Mirror skills into more agent folders",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completeTargets,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			rep, err := engine.New(defaultCacheRoot()).SetMirrors(root, args, nil)
			if err != nil {
				return err
			}
			return print(cmd, rep)
		},
	}
	remove := &cobra.Command{
		Use:               "remove <target>...",
		Aliases:           []string{"rm"},
		Short:             "Stop mirroring into agent folders (removes mirrored copies)",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completeTargets,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			rep, err := engine.New(defaultCacheRoot()).SetMirrors(root, nil, args)
			if err != nil {
				return err
			}
			return print(cmd, rep)
		},
	}
	list := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show mirror targets",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			m, err := manifest.Resolve(root)
			if err != nil {
				return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", root, err)
			}
			return print(cmd, &engine.MirrorReport{Targets: m.Mirrors})
		},
	}
	refresh := &cobra.Command{
		Use:   "refresh",
		Short: "Rebuild every mirror from the primary install",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			rep, err := engine.New(defaultCacheRoot()).RefreshMirrors(root)
			if err != nil {
				return err
			}
			return print(cmd, rep)
		},
	}

	cmd.PersistentFlags().StringVar(&repo, "repo", "", "Target repository path (defaults to current directory)")
	cmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	cmd.AddCommand(add, remove, list, refresh)
	return cmd
}
