package skell

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A project that was just set up (manifest, no skills, no lock file) must give
// machine-readable, error-free answers: the desktop app calls these right
// after "Add skill" sets a folder up.
func TestFreshProject_JSONCommandsReturnEmptyArrays(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "init", "--repo", repo, "--target", "claude", "--yes")
	require.NoError(t, err)

	for _, cmd := range []string{"list", "status", "doctor"} {
		out, err := executeCmd(t, cmd, "--repo", repo, "--json")
		require.NoError(t, err, cmd)
		assert.Equal(t, "[]", strings.TrimSpace(out), cmd)
	}

	out, err := executeCmd(t, "status", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "no skills to show")
}
