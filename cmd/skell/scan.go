package skell

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aminmesbahi/skell/internal/skillscan"
	"github.com/spf13/cobra"
)

// scanEntry is one skill's scan in the CLI's JSON output.
type scanEntry struct {
	Repo string `json:"repo,omitempty"`
	*skillscan.Report
}

func newScanCmd() *cobra.Command {
	var f repoFlags
	var path, targetID, failOn string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "scan [skill-name...]",
		Short: "Scan skills for risky content (prompt injection, hidden text, pipe-to-shell, secrets…)",
		Long: `Statically inspects skills for the patterns that make an agent skill dangerous:
instructions that try to override the agent or hide things from you, commands
that pipe a download into a shell, credential access and exfiltration,
obfuscated code, persistence, destructive commands, binaries and broad tool
permissions. Each skill gets a risk level (clean, low, medium, high, critical)
and a score from 0 to 100.

The scan is heuristic: it points at what deserves a human look, it does not
prove a skill safe or unsafe. Exits non-zero when a skill reaches --fail-on.

Skell also runs this scan during 'skell review' and 'skell install', and an
organisation can refuse risky skills with 'max-risk' in ~/.skell/config.toml.`,
		Example: `  # Every installed skill in the current project
  skell scan

  # One skill, with every finding listed
  skell scan pdf --verbose

  # A skill you are writing, or a whole folder of skills
  skell scan --path ./my-skill
  skell scan --path ~/src/team-skills --fail-on medium

  # Machine-readable
  skell scan --json`,
		ValidArgsFunction: completeInstalledSkills,
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, ok := skillscan.ParseLevel(failOn)
			if !ok {
				return fmt.Errorf("--fail-on must be one of low, medium, high, critical (or empty to never fail)")
			}
			refs, err := collectSkillRefs(cmd, f, path, targetID, args)
			if err != nil {
				return err
			}
			entries := []scanEntry{}
			for _, r := range refs {
				rep, err := skillscan.ScanDir(r.Dir)
				if err != nil {
					return err
				}
				rep.Skill = r.Name
				entries = append(entries, scanEntry{Repo: r.Repo, Report: rep})
			}
			w := cmd.OutOrStdout()
			if f.jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				if err := enc.Encode(entries); err != nil {
					return err
				}
			} else {
				printScans(w, entries, verbose)
			}
			if limit != "" {
				var bad []string
				for _, e := range entries {
					if e.Level.AtLeast(limit) {
						bad = append(bad, e.Skill)
					}
				}
				if len(bad) > 0 {
					return fmt.Errorf("%s: risk is %s or higher", strings.Join(bad, ", "), limit)
				}
			}
			return nil
		},
	}
	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&path, "path", "", "Scan a skill folder (or a folder of skills) instead of installed skills")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform to scan")
	cmd.Flags().StringVar(&failOn, "fail-on", "high", "Exit non-zero when a skill reaches this level: low | medium | high | critical (empty: never)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "List every finding, including notes")
	return cmd
}

func levelMark(l skillscan.Level) string {
	switch {
	case l.AtLeast(skillscan.LevelHigh):
		return markBad
	case l.AtLeast(skillscan.LevelMedium):
		return markWarn
	}
	return markOK
}

func printScans(w io.Writer, entries []scanEntry, verbose bool) {
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(w, "  no skills to scan")
		return
	}
	for _, e := range entries {
		_, _ = fmt.Fprintf(w, "\n  %s  %s  %s", levelMark(e.Level), e.Skill, e.Level)
		if e.Score > 0 {
			_, _ = fmt.Fprintf(w, " (score %d)", e.Score)
		}
		_, _ = fmt.Fprintf(w, "  · %d files\n", e.Files)
		shown, notable := 0, 0
		for _, fd := range e.Findings {
			if fd.Severity != skillscan.SeverityInfo {
				notable++
			}
		}
		for _, fd := range e.Findings {
			if fd.Severity == skillscan.SeverityInfo && !verbose {
				continue
			}
			if !verbose && shown >= 6 {
				break
			}
			shown++
			loc := fd.File
			if fd.Line > 0 {
				loc = fmt.Sprintf("%s:%d", fd.File, fd.Line)
			}
			_, _ = fmt.Fprintf(w, "       %-8s %s — %s\n", fd.Severity, loc, fd.Message)
			if verbose && fd.Evidence != "" {
				_, _ = fmt.Fprintf(w, "                %s\n", truncate(fd.Evidence, 110))
			}
		}
		if !verbose && notable > shown {
			_, _ = fmt.Fprintf(w, "       … %d more (use --verbose)\n", notable-shown)
		}
	}
	_, _ = fmt.Fprintln(w)
}
