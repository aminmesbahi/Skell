package skell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/skilltest"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/spf13/cobra"
)

// agentCommand picks the command that answers prompts: the flag, then the
// SKELL_TEST_AGENT variable. There is deliberately no default: every case is a
// real agent call, so running them is always an explicit choice.
func agentCommand(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("SKELL_TEST_AGENT"); env != "" {
		return env
	}
	return ""
}

func skillsRelDir(targetID string) string {
	t, err := target.Lookup(targetID)
	if err != nil {
		t = target.MustLookup(target.Default)
	}
	return path.Join(t.Dir, target.SkillsSubdir)
}

func newTestCmd() *cobra.Command {
	var f repoFlags
	var skillPath, targetID, agent string
	var only []string
	var timeout int

	cmd := &cobra.Command{
		Use:   "test [skill-name...]",
		Short: "Run the eval cases a skill ships with (evals/evals.yaml) against your agent",
		Long: `Skills can ship test cases in evals/evals.yaml: a prompt and what a good answer
looks like. 'skell test' installs the skill into a scratch workspace, gives each
prompt to your agent, and checks the answer. It tells you whether a skill —
yours or one you are about to upgrade to — still does its job.

The agent is any command that reads a prompt on standard input (or takes it
where you write {prompt}) and prints its answer, e.g. --agent "claude -p".
It defaults to $SKELL_TEST_AGENT. Every case is a real agent call, so there is
no built-in default. Without an agent only the suite itself is checked and cases
are reported as skipped, so this still works in plain CI.

An eval file looks like:

  cases:
    - name: summarises a report
      prompt: Summarise report.pdf in three bullets.
      timeout: 180
      expect:
        contains: ["summary"]
        not_contains: ["I cannot"]
        matches: ["(?m)^- "]

'skell new --evals' writes a starter file. 'skell upgrade --verify' runs the
tests of the new version before it replaces the old one.`,
		Example: `  # A skill you are writing
  skell test --path ./release-notes --agent "claude -p"

  # Every installed skill that has evals
  skell test

  # One case, as JSON
  skell test pdf --case "extracts text" --json`,
		ValidArgsFunction: completeInstalledSkills,
		RunE: func(cmd *cobra.Command, args []string) error {
			refs, err := collectSkillRefs(cmd, f, skillPath, targetID, args)
			if err != nil {
				return err
			}
			tid := targetID
			if tid == "" {
				tid = target.Default
				if skillPath == "" {
					if repos, err := resolveRepos(f); err == nil && len(repos) > 0 {
						tid = engine.ResolveTarget(repos[0]).ID
					}
				}
			}
			cmdline := agentCommand(agent)

			results := []*skilltest.Result{}
			noSuite := []string{}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			for _, r := range refs {
				suite, err := skilltest.Load(r.Dir)
				if errors.Is(err, skilltest.ErrNoSuite) {
					noSuite = append(noSuite, r.Name)
					continue
				}
				if err != nil {
					return fmt.Errorf("%s: %w", r.Name, err)
				}
				res := skilltest.Run(ctx, r.Name, r.Dir, suite, skilltest.Options{
					Agent: cmdline, SkillsDir: skillsRelDir(tid), Only: only,
					Timeout: time.Duration(timeout) * time.Second,
				})
				results = append(results, res)
			}

			w := cmd.OutOrStdout()
			if f.jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				if err := enc.Encode(map[string]any{"results": results, "without_evals": noSuite}); err != nil {
					return err
				}
			} else {
				printTests(w, results, noSuite, cmdline)
			}
			bad := 0
			for _, r := range results {
				if !r.OK() {
					bad++
				}
			}
			if bad > 0 {
				return fmt.Errorf("%d skill(s) have failing evals", bad)
			}
			if len(refs) > 0 && len(results) == 0 && (len(args) > 0 || skillPath != "") {
				return fmt.Errorf("%s ships no evals — see 'skell test --help' for the format", strings.Join(noSuite, ", "))
			}
			return nil
		},
	}
	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&skillPath, "path", "", "Test a skill folder (or a folder of skills) instead of installed skills")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform the skill is installed for")
	cmd.Flags().StringVar(&agent, "agent", "", `Command that answers a prompt, e.g. "claude -p" (default: $SKELL_TEST_AGENT)`)
	cmd.Flags().StringArrayVar(&only, "case", nil, "Run only this case (repeatable)")
	cmd.Flags().IntVar(&timeout, "timeout", 120, "Seconds allowed per case that sets no timeout of its own")
	return cmd
}

func printTests(w io.Writer, results []*skilltest.Result, noSuite []string, agent string) {
	if len(results) == 0 && len(noSuite) == 0 {
		_, _ = fmt.Fprintln(w, "  no skills to test")
		return
	}
	for _, r := range results {
		mark := markOK
		if !r.OK() {
			mark = markBad
		}
		_, _ = fmt.Fprintf(w, "\n  %s  %s", mark, r.Skill)
		switch {
		case r.Skipped == len(r.Cases) && len(r.Cases) > 0:
			_, _ = fmt.Fprintf(w, "  (%d cases, not run)\n", r.Skipped)
		default:
			_, _ = fmt.Fprintf(w, "  %d passed, %d failed, %d errors\n", r.Passed, r.Failed, r.Errored)
		}
		for _, p := range r.Static {
			_, _ = fmt.Fprintf(w, "       ! %s\n", p)
		}
		for _, c := range r.Cases {
			switch c.Status {
			case skilltest.StatusPass:
				_, _ = fmt.Fprintf(w, "       %s %s  %.1fs\n", markOK, c.Name, float64(c.DurationMS)/1000)
			case skilltest.StatusSkipped:
				_, _ = fmt.Fprintf(w, "       - %s\n", c.Name)
			default:
				_, _ = fmt.Fprintf(w, "       %s %s\n", markBad, c.Name)
				for _, fl := range c.Failures {
					_, _ = fmt.Fprintf(w, "           %s\n", truncate(fl, 140))
				}
			}
		}
		if r.Skipped > 0 && strings.TrimSpace(agent) == "" {
			_, _ = fmt.Fprintln(w, `       cases were not run: pass --agent "claude -p" (or set SKELL_TEST_AGENT)`)
		}
	}
	if len(noSuite) > 0 {
		_, _ = fmt.Fprintf(w, "\n  no evals: %s\n", strings.Join(limit(noSuite, 8), ", "))
	}
	_, _ = fmt.Fprintln(w)
}

// upgradeVerifier builds the check behind 'skell upgrade --verify': run the
// incoming version's evals (and the installed version's, as a baseline) and
// refuse the upgrade when it breaks cases that used to pass.
func upgradeVerifier(agentFlag, targetID string, repo string, timeout int) func(name, installedDir, incomingDir string) (string, bool) {
	return func(name, installedDir, incomingDir string) (string, bool) {
		suite, err := skilltest.Load(incomingDir)
		if errors.Is(err, skilltest.ErrNoSuite) {
			return "", true // nothing to verify against
		}
		if err != nil {
			return "its evals could not be read: " + err.Error(), false
		}
		tid := targetID
		if tid == "" {
			tid = engine.ResolveTarget(repo).ID
		}
		agent := agentCommand(agentFlag)
		if strings.TrimSpace(agent) == "" {
			return "--verify needs an agent: pass --agent or set SKELL_TEST_AGENT", false
		}
		opts := skilltest.Options{Agent: agent, SkillsDir: skillsRelDir(tid), Timeout: time.Duration(timeout) * time.Second}
		ctx := context.Background()

		after := skilltest.Run(ctx, name, incomingDir, suite, opts)
		if len(after.Static) > 0 {
			return "its evals have problems: " + after.Static[0], false
		}
		var before *skilltest.Result
		if old, err := skilltest.Load(installedDir); err == nil {
			before = skilltest.Run(ctx, name, installedDir, old, opts)
		}
		if reg := skilltest.Regressions(before, after); len(reg) > 0 {
			return "evals fail on the new version: " + strings.Join(limit(reg, 4), ", "), false
		}
		return "", true
	}
}
