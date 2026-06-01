package skell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheStatusCmd_Succeeds(t *testing.T) {
	// CacheStatus now works; with an empty cache it returns a friendly message.
	out, err := executeCmd(t, "cache", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "cache")
}

func TestCacheClearCmd_Succeeds(t *testing.T) {
	// CacheClear removes the cache dir; succeeds even if it doesn't exist.
	_, err := executeCmd(t, "cache", "clear")
	assert.NoError(t, err)
}

func TestCacheRefreshCmd_NoManifest_Succeeds(t *testing.T) {
	// Refresh is a global operation: it no longer requires a repo manifest.
	// With no manifest, no global sources, and an empty cache, it is a no-op
	// that still reports success.
	repo := t.TempDir()
	out, err := executeCmd(t, "cache", "refresh", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "done")
}

func TestCacheRefreshCmd_EmptyRegistries_Succeeds(t *testing.T) {
	repo := makeRepoWithManifestCmd(t)
	out, err := executeCmd(t, "cache", "refresh", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "done")
}
