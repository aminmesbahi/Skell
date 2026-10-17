package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// One unreadable source must not hide the skills of the others.
func TestSearchMerged_SkipsUnreadableSourceAndReportsIt(t *testing.T) {
	good := t.TempDir()
	skillDir := filepath.Join(good, "skills", "demo")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
		[]byte("---\nname: demo\ndescription: demo skill. Use when testing.\n---\nv1\n"), 0o644))
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	eng := New(filepath.Join(t.TempDir(), "cache"))
	repo := t.TempDir()
	require.NoError(t, eng.InitFor(repo, target.MustLookup("claude")))
	mp := manifest.LocalPathFor(repo, target.MustLookup("claude"))
	m, err := manifest.Read(mp)
	require.NoError(t, err)
	m.Registries["good"] = good
	m.Registries["broken"] = missing
	require.NoError(t, manifest.Write(mp, m))

	skills, warns, err := eng.SearchMergedWithWarnings(repo, "", "", "", "")
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "demo", skills[0].Name)
	require.Len(t, warns, 1)
	assert.Equal(t, "broken", warns[0].Alias)

	// The plain variant keeps working too.
	skills, err = eng.SearchMerged(repo, "demo", "", "", "")
	require.NoError(t, err)
	assert.Len(t, skills, 1)
}
