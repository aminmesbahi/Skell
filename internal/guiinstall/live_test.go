package guiinstall

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aminmesbahi/skell/internal/selfupdate"
	"github.com/stretchr/testify/require"
)

// TestLiveInstall downloads the real latest release. It is skipped unless
// SKELL_LIVE_DOWNLOAD=1, so normal test runs never touch the network.
func TestLiveInstall(t *testing.T) {
	if os.Getenv("SKELL_LIVE_DOWNLOAD") != "1" {
		t.Skip("set SKELL_LIVE_DOWNLOAD=1 to download the latest release")
	}
	rel, err := selfupdate.New("aminmesbahi", "skell").LatestRelease()
	require.NoError(t, err)
	t.Logf("latest release: %s", rel.TagName)

	dir := t.TempDir()
	path, err := NewInstaller().Install(context.Background(), rel, runtime.GOOS, runtime.GOARCH, dir)
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	t.Logf("installed %s (%d bytes) in %s", filepath.Base(path), info.Size(), dir)
	require.Greater(t, info.Size(), int64(1<<20))
}
