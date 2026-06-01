package selfupdate_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aminmesbahi/skell/internal/selfupdate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveReleaseWithChecksums starts a test server exposing an asset archive and a
// checksums.txt, returning a Release wired to those URLs.
func serveReleaseWithChecksums(t *testing.T, assetName string, archive []byte, checksumBody string) (*selfupdate.Release, func()) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/asset", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/checksums", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(checksumBody)) })
	srv := httptest.NewServer(mux)

	rel := &selfupdate.Release{
		TagName: "v1.2.3",
		Assets: []selfupdate.Asset{
			{Name: assetName, BrowserDownloadURL: srv.URL + "/asset"},
			{Name: selfupdate.ChecksumsAssetName, BrowserDownloadURL: srv.URL + "/checksums"},
		},
	}
	return rel, srv.Close
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestDownloadVerified_Succeeds_WhenChecksumMatches(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tarball assets are produced for non-windows platforms")
	}
	archive := buildTarGz(t, "skell", []byte("real binary"))
	assetName := "skell_1.2.3_linux_amd64.tar.gz"
	body := fmt.Sprintf("%s  %s\n", sha256Hex(archive), assetName)
	rel, closeFn := serveReleaseWithChecksums(t, assetName, archive, body)
	defer closeFn()

	dest := filepath.Join(t.TempDir(), "skell")
	u := selfupdate.New("owner", "repo")
	require.NoError(t, u.DownloadVerified(rel, &rel.Assets[0], dest))

	got, err := readFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "real binary", string(got))
}

func TestDownloadVerified_Fails_WhenChecksumMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	archive := buildTarGz(t, "skell", []byte("real binary"))
	assetName := "skell_1.2.3_linux_amd64.tar.gz"
	body := fmt.Sprintf("%s  %s\n", sha256Hex([]byte("different content")), assetName)
	rel, closeFn := serveReleaseWithChecksums(t, assetName, archive, body)
	defer closeFn()

	dest := filepath.Join(t.TempDir(), "skell")
	u := selfupdate.New("owner", "repo")
	err := u.DownloadVerified(rel, &rel.Assets[0], dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestDownloadVerified_Fails_WhenNoChecksumsAsset(t *testing.T) {
	rel := &selfupdate.Release{
		TagName: "v1.2.3",
		Assets:  []selfupdate.Asset{{Name: "skell_1.2.3_linux_amd64.tar.gz", BrowserDownloadURL: "http://example.com/a"}},
	}
	u := selfupdate.New("owner", "repo")
	err := u.DownloadVerified(rel, &rel.Assets[0], filepath.Join(t.TempDir(), "skell"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no checksum published")
}

func TestChecksumFor_ReturnsNotFound_ForUnlistedAsset(t *testing.T) {
	rel, closeFn := serveReleaseWithChecksums(t, "skell_1.2.3_linux_amd64.tar.gz", []byte("x"), "abc123  some-other-file\n")
	defer closeFn()

	u := selfupdate.New("owner", "repo")
	_, found, err := u.ChecksumFor(rel, "skell_1.2.3_linux_amd64.tar.gz")
	require.NoError(t, err)
	assert.False(t, found)
}
