// Package audit writes append-only JSONL audit log entries to ~/.skell/audit.log.
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

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
	return &Logger{logPath: logPath}
}

// Log appends an entry to the audit log.
func (l *Logger) Log(action, skill, version, registry, repo string) (err error) {
	if err = os.MkdirAll(filepath.Dir(l.logPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Action:    action,
		Skill:     skill,
		Version:   version,
		Registry:  registry,
		Repo:      repo,
	}
	err = json.NewEncoder(f).Encode(entry)
	return err
}
