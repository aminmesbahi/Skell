package hasher_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashDir_Deterministic(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0600))

	h1, err := hasher.HashDir(dir)
	require.NoError(t, err)

	h2, err := hasher.HashDir(dir)
	require.NoError(t, err)

	assert.Equal(t, h1, h2)
}

func TestHashDir_ChangesOnFileChange(t *testing.T) {
	dir := t.TempDir()
	skillFile := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte("# original"), 0600))

	h1, err := hasher.HashDir(dir)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(skillFile, []byte("# modified"), 0600))

	h2, err := hasher.HashDir(dir)
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2)
}

func TestHashDir_HasSha256Prefix(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0600))

	h, err := hasher.HashDir(dir)
	require.NoError(t, err)
	assert.Contains(t, h, "sha256:")
}

func TestHashDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	h, err := hasher.HashDir(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, h)
}

func TestHashDir_NonExistentDir(t *testing.T) {
	_, err := hasher.HashDir("/nonexistent/path")
	assert.Error(t, err)
}

func TestVerify_Match(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0600))

	hash, err := hasher.HashDir(dir)
	require.NoError(t, err)

	ok, err := hasher.Verify(dir, hash)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerify_Mismatch(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0600))

	ok, err := hasher.Verify(dir, "sha256:wronghash")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestVerify_NonExistentDir_ReturnsError(t *testing.T) {
	ok, err := hasher.Verify("/nonexistent/path/to/skill", "sha256:abc")
	assert.Error(t, err)
	assert.False(t, ok)
}
