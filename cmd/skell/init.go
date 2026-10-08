package skell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var repo string
	var targetID string
	var nonInteractive, user bool
	var mirrors []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create skell.toml from currently installed skills",
		Long: `Scans the repository for installed skills and generates a skell.toml manifest.

Skell supports multiple AI-agent layouts:

  claude    .claude/skills/   (Anthropic Claude Code, agentskills.io)
  codex     .codex/skills/    (OpenAI Codex CLI)
  copilot   .github/skills/   (VS Code & GitHub Copilot)
  cursor    .cursor/skills/   (Cursor)
  windsurf  .windsurf/skills/ (Windsurf / Cascade)
  opencode  .opencode/skills/ (OpenCode)
  cline     .cline/skills/    (Cline)
  grok      .grok/skills/     (xAI Grok)

If a known agent folder already exists in the repo it is used automatically.
Otherwise pass --target to choose, or run interactively to be prompted.`,
		Example: `  # Initialise the current directory (auto-detect or prompt)
  skell init

  # Initialise for a specific platform
  skell init --target copilot

  # Initialise a specific repository path
  skell init --repo /path/to/repo --target cursor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetRepo := repo
			if user {
				if repo != "" {
					return fmt.Errorf("--user cannot be combined with --repo")
				}
				root, err := userRoot()
				if err != nil {
					return err
				}
				targetRepo = root
			}
			if targetRepo == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				targetRepo = cwd
			}

			t, err := chooseInitTarget(cmd, targetRepo, targetID, nonInteractive)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			if err := eng.InitFor(targetRepo, t); err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			manifestPath := filepath.Join(targetRepo, t.Dir, "skell.toml")
			_, _ = fmt.Fprintf(w, "  done  skell.toml created at %s (target: %s)\n", manifestPath, t.ID)

			interactive := !nonInteractive && isInteractive()
			if len(mirrors) == 0 && interactive {
				mirrors = promptForMirrors(cmd, t.ID)
			}
			if len(mirrors) > 0 {
				rep, err := eng.SetMirrors(targetRepo, mirrors, nil)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(w, "  done  skills will also be provided to: %s\n", strings.Join(rep.Targets, ", "))
			}
			if interactive {
				promptForStarterSource(cmd, eng, targetRepo)
			}
			_, _ = fmt.Fprintln(w, "\n  next steps:")
			_, _ = fmt.Fprintln(w, "    skell catalog              browse well-known skill sources")
			_, _ = fmt.Fprintln(w, "    skell add <source>         add one (catalog id, owner/repo or URL)")
			_, _ = fmt.Fprintln(w, "    skell search               see the skills it offers")
			_, _ = fmt.Fprintln(w, "    skell install <skill>      install one")
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&mirrors, "mirror", nil, "Also provide skills to these agents, e.g. --mirror copilot,cursor (see 'skell mirror')")
	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (defaults to current directory)")
	cmd.Flags().BoolVar(&user, "user", false, "Set up your personal, user-level skills (e.g. ~/.claude/skills) instead of a project")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform layout: claude | codex | copilot | cursor | windsurf | opencode | cline | grok")
	cmd.Flags().BoolVar(&nonInteractive, "yes", false, "Do not prompt; use --target or the default (claude)")
	return cmd
}

// chooseInitTarget resolves which target to initialise for. Order:
//  1. explicit --target flag
//  2. existing layout already on disk (auto-detected)
//  3. interactive prompt on a TTY
//  4. fallback to the default target
func chooseInitTarget(cmd *cobra.Command, repoRoot, flag string, nonInteractive bool) (target.Target, error) {
	if flag != "" {
		return target.Lookup(flag)
	}
	if t, ok := target.DetectPrimary(repoRoot); ok {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  detected layout: %s (%s/)\n", t.ID, t.Dir)
		return t, nil
	}
	if nonInteractive || !isInteractive() {
		return target.MustLookup(target.Default), nil
	}
	return promptForTarget(cmd)
}

func isInteractive() bool {
	in, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	out, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (in.Mode()&os.ModeCharDevice) != 0 && (out.Mode()&os.ModeCharDevice) != 0
}

func promptForTarget(cmd *cobra.Command) (target.Target, error) {
	all := target.All()
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "Select an agent platform layout:")
	for i, t := range all {
		_, _ = fmt.Fprintf(out, "  %d) %-8s  %s   (%s/skills/)\n", i+1, t.ID, t.DisplayName, t.Dir)
	}
	_, _ = fmt.Fprintf(out, "Choice [1-%d, default 1]: ", len(all))

	reader := stdinReader
	line, err := reader.ReadString('\n')
	if err != nil {
		// Stdin closed/redirected with no data: take the default.
		return all[0], nil
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return all[0], nil
	}
	for i, t := range all {
		if line == fmt.Sprintf("%d", i+1) {
			return t, nil
		}
	}
	return target.Lookup(line)
}

// promptForMirrors asks which other agents should receive the skills.
func promptForMirrors(cmd *cobra.Command, primary string) []string {
	out := cmd.OutOrStdout()
	var others []string
	for _, t := range target.All() {
		if t.ID != primary {
			others = append(others, t.ID)
		}
	}
	_, _ = fmt.Fprintf(out, "\nDo you also use other agents in this repo? Skell can keep a copy of every\nskill in their folders too (%s).\n", strings.Join(others, ", "))
	_, _ = fmt.Fprint(out, "Agents to mirror to [comma-separated, Enter for none]: ")
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return nil
	}
	var ids []string
	for _, part := range strings.Split(line, ",") {
		if id := strings.TrimSpace(part); id != "" {
			if t, err := target.Lookup(id); err == nil && t.ID != primary {
				ids = append(ids, t.ID)
			} else {
				_, _ = fmt.Fprintf(out, "  skipping unknown or primary agent %q\n", id)
			}
		}
	}
	return ids
}

// promptForStarterSource offers to add a catalog source right away, so a new
// project has skills to browse immediately.
func promptForStarterSource(cmd *cobra.Command, eng *engine.Engine, repo string) {
	out := cmd.OutOrStdout()
	sources := loadCatalog(false).Search("")
	if len(sources) == 0 {
		return
	}
	_, _ = fmt.Fprintln(out, "\nAdd a skill source to start with?")
	for i, s := range sources {
		_, _ = fmt.Fprintf(out, "  %d) %-16s %s\n", i+1, s.ID, s.Name)
	}
	_, _ = fmt.Fprintf(out, "Choice [1-%d, Enter to skip]: ", len(sources))
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	var pick *catalogSource
	for i, s := range sources {
		if line == fmt.Sprintf("%d", i+1) || strings.EqualFold(line, s.ID) {
			pick = &catalogSource{ID: s.ID, URL: s.URL}
		}
	}
	if pick == nil {
		_, _ = fmt.Fprintf(out, "  skipping: %q is not a listed choice\n", line)
		return
	}
	res, err := eng.AddFromURLWith(repo, pick.URL, engine.AddOptions{Alias: pick.ID})
	if err != nil {
		_, _ = fmt.Fprintf(out, "  could not add %s: %v\n", pick.ID, err)
		return
	}
	_, _ = fmt.Fprintf(out, "  done  added source %q — try 'skell search'\n", res.Alias)
}

type catalogSource struct{ ID, URL string }
