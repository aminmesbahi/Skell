package audit_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog_WritesJSONLEntry(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger := audit.NewLogger(logPath)
	require.NoError(t, logger.Log(audit.ActionInstall, "pdf-processing", "1.2.0", "default", "my-project"))

	f, err := os.Open(logPath)
	require.NoError(t, err)
	defer f.Close()

	var entry audit.Entry
	require.NoError(t, json.NewDecoder(bufio.NewReader(f)).Decode(&entry))

	assert.Equal(t, audit.ActionInstall, entry.Action)
	assert.Equal(t, "pdf-processing", entry.Skill)
	assert.Equal(t, "1.2.0", entry.Version)
	assert.NotEmpty(t, entry.Timestamp)
}

func TestLog_AppendsMultipleEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	logger := audit.NewLogger(logPath)

	require.NoError(t, logger.Log(audit.ActionInstall, "skill-a", "1.0.0", "default", "repo"))
	require.NoError(t, logger.Log(audit.ActionRemove, "skill-b", "", "default", "repo"))

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)

	lines := 0
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		if scanner.Text() != "" {
			lines++
		}
	}
	assert.Equal(t, 2, lines)
}

func TestLog_CreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "subdir", "audit.log")

	logger := audit.NewLogger(logPath)
	require.NoError(t, logger.Log(audit.ActionPin, "skill-a", "1.0.0", "default", "repo"))

	assert.FileExists(t, logPath)
}

func TestLog_ErrorOnUnwritablePath(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; skipping unwritable-path test")
	}
	// Use a path nested under a file (not a directory) to force MkdirAll to fail.
	dir := t.TempDir()
	blockFile := filepath.Join(dir, "block")
	require.NoError(t, os.WriteFile(blockFile, []byte("x"), 0600))
	logPath := filepath.Join(blockFile, "subdir", "audit.log")

	logger := audit.NewLogger(logPath)
	err := logger.Log(audit.ActionInstall, "skill", "1.0.0", "reg", "repo")
	assert.Error(t, err)
}
