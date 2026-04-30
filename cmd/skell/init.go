package skell

import (
	"bufio"
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
	var nonInteractive bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create skell.toml from currently installed skills",
		Long: `Scans the repository for installed skills and generates a skell.toml manifest.

Skell supports multiple AI-agent layouts:

  claude    .claude/skills/   (Anthropic Claude Code, agentskills.io)
  codex     .codex/skills/    (OpenAI Codex CLI)
  copilot   .github/skills/   (VS Code & GitHub Copilot)
  cursor    .cursor/skills/   (Cursor)

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

			manifestPath := filepath.Join(targetRepo, t.Dir, "skell.toml")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  done  skell.toml created at %s (target: %s)\n", manifestPath, t.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (defaults to current directory)")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform layout: claude | codex | copilot | cursor")
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

	reader := bufio.NewReader(os.Stdin)
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
