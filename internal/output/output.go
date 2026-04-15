// Package output formats and prints CLI output using lipgloss.
package output

import "github.com/aminmesbahi/skell/internal/model"

// Printer renders CLI output. When jsonMode is true all methods emit JSON instead.
type Printer struct {
	jsonMode bool
}

// NewPrinter creates a Printer.
func NewPrinter(jsonMode bool) *Printer {
	return &Printer{jsonMode: jsonMode}
}

// PrintStatusTable renders the tabular output for `skell status`.
func (p *Printer) PrintStatusTable(entries []model.StatusEntry) {
	// TODO: implement using lipgloss table
	panic("not implemented")
}

// PrintSkillList renders the output for `skell list`.
func (p *Printer) PrintSkillList(skills []model.InstalledSkill) {
	// TODO: implement
	panic("not implemented")
}

// PrintDiagnostics renders the output for `skell doctor`.
func (p *Printer) PrintDiagnostics(issues []interface{}) {
	// TODO: implement
	panic("not implemented")
}

// Success prints a success line (e.g. "done  1 skill installed").
func (p *Printer) Success(msg string) {
	// TODO: implement
	panic("not implemented")
}

// Error prints an error line with an optional hint.
func (p *Printer) Error(msg, hint string) {
	// TODO: implement
	panic("not implemented")
}
