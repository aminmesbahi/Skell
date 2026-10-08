package hasher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerify_AcceptsLegacyNativeSeparatorHash(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "a.txt"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644))

	h, err := HashDir(dir)
	require.NoError(t, err)
	ok, err := Verify(dir, h)
	require.NoError(t, err)
	assert.True(t, ok)

	legacy, err := hashDir(dir, true)
	require.NoError(t, err)
	ok, err = Verify(dir, legacy)
	require.NoError(t, err)
	assert.True(t, ok, "hashes written with native separators must still verify")

	ok, err = Verify(dir, "sha256:deadbeef")
	require.NoError(t, err)
	assert.False(t, ok)
}
