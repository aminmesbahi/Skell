package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncCheck_AllowModifiedToleratesDeliberateEdits(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))

	skill := filepath.Join(f.repo, ".claude", "skills", "demo", "SKILL.md")
	orig, err := os.ReadFile(skill)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(skill, append(orig, []byte("customised\n")...), 0o644))

	_, err = f.eng.SyncWith(f.repo, SyncOptions{CheckOnly: true})
	var diff *SyncDiffError
	require.ErrorAs(t, err, &diff)
	assert.Equal(t, []string{"demo"}, diff.Modified)

	_, err = f.eng.SyncWith(f.repo, SyncOptions{CheckOnly: true, AllowModified: true})
	require.NoError(t, err)
}
