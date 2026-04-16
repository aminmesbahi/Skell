// Package audit writes append-only JSONL audit log entries to ~/.skell/audit.log.
package audit

import "time"

// Action names logged by the audit package.
const (
	ActionInstall = "install"
	ActionUpgrade = "upgrade"
	ActionRemove  = "remove"
	ActionPin     = "pin"
	ActionUnpin   = "unpin"
	ActionSync    = "sync"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Skill     string `json:"skill"`
	Version   string `json:"version,omitempty"`
	Registry  string `json:"registry,omitempty"`
	Repo      string `json:"repo"`
	User      string `json:"user"`
}

// Logger appends audit entries to a JSONL file.
type Logger struct {
	logPath string
}

// NewLogger creates a Logger that writes to the given file path.
func NewLogger(logPath string) *Logger {
	// TODO: implement
	panic("not implemented")
}

// Log appends an entry to the audit log.
func (l *Logger) Log(action, skill, version, registry, repo string) error {
	// TODO: implement using time.Now().UTC().Format(time.RFC3339)
	_ = time.RFC3339
	_ = l.logPath
	panic("not implemented")
}
