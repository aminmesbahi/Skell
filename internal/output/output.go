// Package output formats and prints CLI output.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/aminmesbahi/skell/internal/model"
)

// Printer renders CLI output. When jsonMode is true all methods emit JSON instead.
type Printer struct {
	jsonMode bool
	w        io.Writer
}

// NewPrinter creates a Printer writing to stdout.
func NewPrinter(jsonMode bool) *Printer {
	return &Printer{jsonMode: jsonMode, w: os.Stdout}
}

// NewPrinterTo creates a Printer writing to the given writer (useful for tests).
func NewPrinterTo(w io.Writer, jsonMode bool) *Printer {
	return &Printer{jsonMode: jsonMode, w: w}
}

// PrintStatusTable renders the tabular output for `skell status`.
func (p *Printer) PrintStatusTable(entries []model.StatusEntry) {
	if p.jsonMode {
		p.printJSON(entries)
		return
	}
	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "NAME\tINSTALLED\tLATEST\tSTATUS")
	for _, e := range entries {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", e.Name, e.Installed, e.Latest, e.Status)
	}
	_ = tw.Flush()
}

// PrintSkillList renders the output for `skell list` (installed skills).
func (p *Printer) PrintSkillList(skills []model.InstalledSkill) {
	if p.jsonMode {
		p.printJSON(skills)
		return
	}
	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "NAME\tVERSION\tREGISTRY\tPINNED")
	for _, s := range skills {
		pinned := ""
		if s.Pinned {
			pinned = "yes"
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Name, s.Version, s.Registry, pinned)
	}
	_ = tw.Flush()
}

// PrintRegistrySkillList renders the output for `skell list --source registry` and `skell search`.
func (p *Printer) PrintRegistrySkillList(skills []model.RegistrySkill) {
	if p.jsonMode {
		p.printJSON(skills)
		return
	}
	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "NAME\tVERSION\tLIFECYCLE\tOWNER")
	for _, s := range skills {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			s.Name, s.Metadata.Version, s.Metadata.Lifecycle, s.Metadata.Owner)
	}
	_ = tw.Flush()
}

// PrintInfoResult renders the output for `skell info`.
func (p *Printer) PrintInfoResult(name string, r *model.InfoResult) {
	if p.jsonMode {
		type infoJSON struct {
			Name   string                `json:"name"`
			Skill  *model.RegistrySkill  `json:"skill,omitempty"`
			Lock   *model.InstalledSkill `json:"lock,omitempty"`
			Status model.SkillStatus     `json:"status,omitempty"`
		}
		p.printJSON(infoJSON{Name: name, Skill: r.Skill, Lock: r.Lock, Status: r.Status})
		return
	}
	fw := &fmtLineWriter{w: p.w}
	fw.printf("  name       %s\n", name)
	if r.Skill != nil {
		if r.Skill.Metadata.Version != "" {
			fw.printf("  version    %s\n", r.Skill.Metadata.Version)
		}
		if r.Skill.Description != "" {
			fw.printf("  desc       %s\n", r.Skill.Description)
		}
		if r.Skill.Metadata.Owner != "" {
			fw.printf("  owner      %s\n", r.Skill.Metadata.Owner)
		}
		if r.Skill.Metadata.Lifecycle != "" {
			fw.printf("  lifecycle  %s\n", r.Skill.Metadata.Lifecycle)
		}
		if r.Skill.Metadata.Scope != "" {
			fw.printf("  scope      %s\n", r.Skill.Metadata.Scope)
		}
		if r.Skill.Metadata.Tags != "" {
			fw.printf("  tags       %s\n", r.Skill.Metadata.Tags)
		}
		if r.Skill.Metadata.SourceRepo != "" {
			fw.printf("  source     %s\n", r.Skill.Metadata.SourceRepo)
		}
		if r.Skill.License != "" {
			fw.printf("  license    %s\n", r.Skill.License)
		}
	}
	if r.Lock != nil {
		fw.printf("  installed  %s\n", r.Lock.Version)
		fw.printf("  locked at  %s\n", r.Lock.InstalledAt)
		fw.printf("  hash       %s\n", r.Lock.ContentHash)
		if r.Lock.Pinned {
			fw.printf("  pinned     true\n")
		}
	}
	if r.Status != "" {
		fw.printf("  status     %s\n", r.Status)
	}
}

// ActionEvent is a single structured event emitted by write commands when --json is set.
type ActionEvent struct {
	Action  string `json:"action"`
	Skill   string `json:"skill,omitempty"`
	Repo    string `json:"repo,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	DryRun  bool   `json:"dry_run,omitempty"`
	Message string `json:"message,omitempty"`
}

// PrintAction emits a single action event (plain text or JSON depending on mode).
func (p *Printer) PrintAction(event ActionEvent) {
	if p.jsonMode {
		p.printJSON(event)
		return
	}
	label := event.Action
	detail := event.Skill
	if event.Message != "" {
		detail = event.Message
	}
	if event.DryRun {
		label = "dry-run"
		detail = fmt.Sprintf("would %s %s", event.Action, event.Skill)
	}
	_, _ = fmt.Fprintf(p.w, "  %-14s %s\n", label, detail)
}

// DiagnosticEntry is the generic representation of a doctor issue for output purposes.
type DiagnosticEntry struct {
	Severity string `json:"severity"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message"`
	Hint     string `json:"hint,omitempty"`
}

// PrintDiagnostics renders the output for `skell doctor`.
func (p *Printer) PrintDiagnostics(issues []DiagnosticEntry) {
	if p.jsonMode {
		p.printJSON(issues)
		return
	}
	if len(issues) == 0 {
		_, _ = fmt.Fprintln(p.w, "  ok  no issues found")
		return
	}
	for _, issue := range issues {
		_, _ = fmt.Fprintf(p.w, "  [%s]  %s\n", issue.Severity, issue.Message)
		if issue.Hint != "" {
			_, _ = fmt.Fprintf(p.w, "         hint: %s\n", issue.Hint)
		}
	}
}

// Success prints a success line.
func (p *Printer) Success(msg string) {
	if p.jsonMode {
		p.printJSON(ActionEvent{Action: "success", Message: msg})
		return
	}
	_, _ = fmt.Fprintf(p.w, "  ✓  %s\n", msg)
}

// Error prints an error line with an optional hint.
func (p *Printer) Error(msg, hint string) {
	if p.jsonMode {
		p.printJSON(ActionEvent{Action: "error", Message: msg})
		return
	}
	_, _ = fmt.Fprintf(p.w, "  ✗  %s\n", msg)
	if hint != "" {
		_, _ = fmt.Fprintf(p.w, "     hint: %s\n", hint)
	}
}

func (p *Printer) printJSON(v interface{}) {
	enc := json.NewEncoder(p.w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// fmtLineWriter records the first write error so callers can check it after a sequence of prints.
type fmtLineWriter struct {
	w   io.Writer
	err error
}

func (fw *fmtLineWriter) printf(format string, args ...any) {
	if fw.err != nil {
		return
	}
	_, fw.err = fmt.Fprintf(fw.w, format, args...)
}
