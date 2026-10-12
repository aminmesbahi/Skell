package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncCheck_FlagsSkillsEditedSinceInstall(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))

	_, err := f.eng.Sync(f.repo, true, false, false)
	require.NoError(t, err, "a pristine install is in sync")

	skill := filepath.Join(f.repo, ".claude", "skills", "demo", "SKILL.md")
	orig, err := os.ReadFile(skill)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(skill, append(orig, []byte("tampered\n")...), 0o644))

	_, err = f.eng.Sync(f.repo, true, false, false)
	var diff *SyncDiffError
	require.ErrorAs(t, err, &diff)
	assert.Equal(t, []string{"demo"}, diff.Modified)
	assert.Contains(t, err.Error(), "modified since install: demo")

	// Deliberate customisation can be allowed.
	f.eng.SetAllowModified(true)
	_, err = f.eng.Sync(f.repo, true, false, false)
	require.NoError(t, err)
	f.eng.SetAllowModified(false)

	// A normal sync reports the edit but never overwrites it.
	rep, err := f.eng.Sync(f.repo, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"demo"}, rep.Modified)
	assert.Contains(t, f.installedBody(t), "tampered")
}
