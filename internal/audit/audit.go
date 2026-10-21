// Package audit writes append-only JSONL audit log entries to ~/.skell/audit.log.
package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"os/user"
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

// Log appends an entry to the audit log. A Logger with an empty path is a no-op,
// which keeps tests from writing to the developer's real ~/.skell/audit.log.
func (l *Logger) Log(action, skill, version, registry, repo string) (err error) {
	if l.logPath == "" {
		return nil
	}
	if err = os.MkdirAll(filepath.Dir(l.logPath), 0700); err != nil {
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
		User:      currentUser(),
	}
	err = json.NewEncoder(f).Encode(entry)
	return err
}

// currentUser returns the best-effort current username for audit attribution
// (design §16.3). Falls back through os/user and common environment variables.
func currentUser() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	for _, env := range []string{"USER", "USERNAME", "LOGNAME"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return ""
}

// ReadRecent returns up to limit of the newest entries in the audit log at
// path, newest first. A missing log is an empty history. Lines that are not
// valid entries are skipped.
func ReadRecent(path string, limit int) ([]Entry, error) {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var all []Entry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		var e Entry
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.Action != "" {
			all = append(all, e)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
