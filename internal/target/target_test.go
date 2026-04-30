package target

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupKnownTargets(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"claude", "claude"},
		{"Claude-Code", "claude"},
		{"anthropic", "claude"},
		{"codex", "codex"},
		{"openai", "codex"},
		{"copilot", "copilot"},
		{"github-copilot", "copilot"},
		{"vscode", "copilot"},
		{"github", "copilot"},
		{"cursor", "cursor"},
	}
	for _, tc := range cases {
		got, err := Lookup(tc.input)
		require.NoError(t, err, tc.input)
		assert.Equal(t, tc.want, got.ID, tc.input)
	}
}

func TestLookupUnknown(t *testing.T) {
	_, err := Lookup("does-not-exist")
	assert.Error(t, err)
	_, err = Lookup("")
	assert.Error(t, err)
}

func TestPathHelpers(t *testing.T) {
	tg := MustLookup("codex")
	repo := "/tmp/r"
	assert.Equal(t, filepath.Join(repo, ".codex", "skills"), tg.SkillsDir(repo))
	assert.Equal(t, filepath.Join(repo, ".codex", "skell.toml"), tg.ManifestPath(repo))
	assert.Equal(t, filepath.Join(repo, ".codex", "skell.lock"), tg.LockPath(repo))
	assert.Equal(t, filepath.Join(".codex", "skills", "foo"), tg.InstalledRelPath("foo"))
}

func TestDetectAndDetectPrimary(t *testing.T) {
	repo := t.TempDir()
	// Empty repo: no detection.
	assert.Empty(t, Detect(repo))
	_, ok := DetectPrimary(repo)
	assert.False(t, ok)

	// Cursor skills folder only.
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".cursor", "skills"), 0o755))
	got := Detect(repo)
	require.Len(t, got, 1)
	assert.Equal(t, "cursor", got[0].ID)

	// Claude manifest takes priority over cursor folder.
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".claude"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".claude", "skell.toml"), []byte(""), 0o600))
	prim, ok := DetectPrimary(repo)
	require.True(t, ok)
	assert.Equal(t, "claude", prim.ID)
}

func TestAllAndIDs(t *testing.T) {
	assert.ElementsMatch(t, []string{"claude", "codex", "copilot", "cursor"}, IDs())
	assert.Len(t, All(), 4)
}
