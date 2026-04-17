package skell

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executeCmd runs the given cobra sub-command with args and captures stdout.
func executeCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestInitCmd_CreatesManifest(t *testing.T) {
	repo := t.TempDir()
	// Create a skill so the manifest is non-trivial
	skillDir := filepath.Join(repo, ".claude", "skills", "pdf-processing")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(skillDir, "SKILL.md"),
		[]byte("---\nname: pdf-processing\nmetadata:\n  version: \"1.0.0\"\n---\n"),
		0600))

	out, err := executeCmd(t, "init", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "skell.toml created")

	m, readErr := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, readErr)
	assert.Contains(t, m.Skills, "pdf-processing")
}

func TestInitCmd_FailsIfManifestExists(t *testing.T) {
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "skell.toml"), []byte("[skills]\n"), 0600))

	_, err := executeCmd(t, "init", "--repo", repo)
	assert.Error(t, err)
}

func TestInitCmd_EmptyRepo_NoError(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "init", "--repo", repo)
	require.NoError(t, err)

	_, statErr := os.Stat(manifest.LocalPath(repo))
	assert.NoError(t, statErr, "skell.toml should exist")
}
