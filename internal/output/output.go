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

// PrintSkillList renders the output for `skell list`.
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

// DiagnosticEntry is the generic representation of a doctor issue for output purposes.
type DiagnosticEntry struct {
	Severity string `json:"severity"`
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
	_, _ = fmt.Fprintf(p.w, "  ✓  %s\n", msg)
}

// Error prints an error line with an optional hint.
func (p *Printer) Error(msg, hint string) {
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
