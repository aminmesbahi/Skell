package engine

// DiagnosticSeverity indicates how serious a diagnostic finding is.
type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "error"
	SeverityWarning DiagnosticSeverity = "warning"
	SeverityInfo    DiagnosticSeverity = "info"
)

// DiagnosticIssue is a single finding produced by skell doctor.
type DiagnosticIssue struct {
	Severity DiagnosticSeverity
	Code     string // stable identifier, e.g. "malformed-frontmatter"
	Message  string
	Hint     string
}
