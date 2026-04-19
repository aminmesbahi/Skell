package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew verifies that New returns a non-nil engine wired with the real provider.
func TestNew_ReturnsNonNil(t *testing.T) {
	cacheRoot := t.TempDir()
	e := New(cacheRoot)
	assert.NotNil(t, e)
}

// TestSyncDiffError_Error validates the Error() string.
func TestSyncDiffError_Error_MissingAndExtra(t *testing.T) {
	err := &SyncDiffError{Missing: []string{"a", "b"}, Extra: []string{"c"}}
	msg := err.Error()
	assert.Contains(t, msg, "missing")
	assert.Contains(t, msg, "a")
	assert.Contains(t, msg, "extra")
	assert.Contains(t, msg, "c")
}

func TestSyncDiffError_Error_OnlyMissing(t *testing.T) {
	err := &SyncDiffError{Missing: []string{"foo"}}
	assert.Contains(t, err.Error(), "missing")
	assert.NotContains(t, err.Error(), "extra")
}

func TestSyncDiffError_Error_OnlyExtra(t *testing.T) {
	err := &SyncDiffError{Extra: []string{"bar"}}
	assert.Contains(t, err.Error(), "extra")
	assert.NotContains(t, err.Error(), "missing")
}

// TestCacheStatus_EmptyCache verifies CacheStatus works on an empty cache.
func TestCacheStatus_EmptyCache(t *testing.T) {
	e := New(t.TempDir())
	status, err := e.CacheStatus()
	require.NoError(t, err)
	assert.Contains(t, status, "empty")
}

// TestCacheStatus_WithCachedRegistry verifies that a populated cache is reported.
func TestCacheStatus_WithCachedRegistry(t *testing.T) {
	cacheRoot := t.TempDir()
	// Simulate a cached registry directory.
	require.NoError(t, os.MkdirAll(filepath.Join(cacheRoot, "my-registry"), 0755))

	e := New(cacheRoot)
	status, err := e.CacheStatus()
	require.NoError(t, err)
	assert.Contains(t, status, "cache status")
}

// TestCacheClear removes all cached data.
func TestCacheClear_RemovesCache(t *testing.T) {
	cacheRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(cacheRoot, "reg1"), 0755))

	e := New(cacheRoot)
	require.NoError(t, e.CacheClear())
	_, err := os.Stat(cacheRoot)
	assert.True(t, os.IsNotExist(err))
}

// TestCacheRefresh_EmptyManifest does not error when manifest has no registries.
func TestCacheRefresh_EmptyManifest_NoError(t *testing.T) {
	e := New(t.TempDir())
	m := &manifest.Manifest{Registries: map[string]string{}}
	require.NoError(t, e.CacheRefresh(m))
}
