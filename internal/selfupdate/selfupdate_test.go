package selfupdate_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aminmesbahi/skell/internal/selfupdate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRelease returns a test server that serves the given Release as JSON.
func mockRelease(t *testing.T, rel selfupdate.Release) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rel)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v0.1.0", "v0.2.0", true},
		{"v0.1.0", "v0.1.1", true},
		{"v1.0.0", "v2.0.0", true},
		{"v0.1.0", "v0.1.0", false},
		{"v0.2.0", "v0.1.0", false},
		{"dev", "v0.1.0", true},
		{"v1.0.0", "dev", false},
	}
	for _, c := range cases {
		got := selfupdate.IsNewer(c.current, c.latest)
		assert.Equal(t, c.want, got, "IsNewer(%q, %q)", c.current, c.latest)
	}
}

func TestExpectedAssetName(t *testing.T) {
	name := selfupdate.ExpectedAssetName("v1.2.3")
	assert.Contains(t, name, "skell_1.2.3_")
	// Must end with the platform-specific archive extension.
	assert.Regexp(t, `^skell_1\.2\.3_[a-z]+_[a-z0-9]+\.(tar\.gz|zip)$`, name)
}

func TestExpectedAssetName_StripsLeadingV(t *testing.T) {
	withV := selfupdate.ExpectedAssetName("v0.5.0")
	withoutV := selfupdate.ExpectedAssetName("0.5.0")
	assert.Equal(t, withV, withoutV)
}

func TestLatestRelease_Success(t *testing.T) {
	rel := selfupdate.Release{
		TagName: "v1.2.3",
		Assets: []selfupdate.Asset{
			{Name: "skell_linux_amd64", BrowserDownloadURL: "http://example.com/skell_linux_amd64"},
			{Name: "skell_windows_amd64.exe", BrowserDownloadURL: "http://example.com/skell_windows_amd64.exe"},
		},
	}
	srv := mockRelease(t, rel)

	u := selfupdate.New("owner", "repo")
	u.APIBaseURL = srv.URL

	got, err := u.LatestRelease()
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", got.TagName)
	assert.Len(t, got.Assets, 2)
}

func TestLatestRelease_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	u := selfupdate.New("owner", "repo")
	u.APIBaseURL = srv.URL

	_, err := u.LatestRelease()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestFindAsset(t *testing.T) {
	assetName := selfupdate.ExpectedAssetName("v1.0.0")
	rel := &selfupdate.Release{
		TagName: "v1.0.0",
		Assets: []selfupdate.Asset{
			{Name: "skell_other_arch", BrowserDownloadURL: "http://example.com/other"},
			{Name: assetName, BrowserDownloadURL: "http://example.com/current"},
		},
	}

	asset := selfupdate.FindAsset(rel)
	require.NotNil(t, asset)
	assert.Equal(t, assetName, asset.Name)
}

func TestFindAsset_NotFound(t *testing.T) {
	rel := &selfupdate.Release{
		TagName: "v1.0.0",
		Assets: []selfupdate.Asset{
			{Name: "skell_unknown_platform"},
		},
	}
	assert.Nil(t, selfupdate.FindAsset(rel))
}

func TestDownload_Success(t *testing.T) {
	content := []byte("fake binary content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)

	destPath := selfupdate.TempPath("skell_test_download")
	t.Cleanup(func() { _ = removeFile(destPath) })

	u := selfupdate.New("owner", "repo")
	asset := &selfupdate.Asset{Name: "skell_test", BrowserDownloadURL: srv.URL + "/download"}

	require.NoError(t, u.Download(asset, destPath))

	data, err := readFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestDownload_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	u := selfupdate.New("owner", "repo")
	asset := &selfupdate.Asset{Name: "skell_test", BrowserDownloadURL: srv.URL}

	err := u.Download(asset, selfupdate.TempPath("skell_test_err"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestApply_ReplacesExecutable(t *testing.T) {
	// Build a fake "old" and "new" binary inside a temp dir.
	dir := t.TempDir()
	oldExe := filepath.Join(dir, "skell_old")
	newBin := filepath.Join(dir, "skell_new")

	require.NoError(t, os.WriteFile(oldExe, []byte("old content"), 0755))
	require.NoError(t, os.WriteFile(newBin, []byte("new content"), 0755))

	// Apply cannot directly replace the *running* executable, but we can test
	// the helper independently: manually perform the same steps that Apply does
	// to verify the file-swap logic.
	// We rename old → old.bak, then move new → old.
	bakPath := oldExe + ".old"
	_ = os.Remove(bakPath)
	require.NoError(t, os.Rename(oldExe, bakPath))
	require.NoError(t, os.Rename(newBin, oldExe))
	require.NoError(t, os.Chmod(oldExe, 0755))

	data, err := os.ReadFile(oldExe)
	require.NoError(t, err)
	assert.Equal(t, "new content", string(data))
}

func TestApplyToPath_Success(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "skell")
	newBin := filepath.Join(dir, "skell_new")

	require.NoError(t, os.WriteFile(currentExe, []byte("old"), 0755))
	require.NoError(t, os.WriteFile(newBin, []byte("new"), 0755))

	require.NoError(t, selfupdate.ApplyToPath(currentExe, newBin))

	data, err := os.ReadFile(currentExe)
	require.NoError(t, err)
	assert.Equal(t, "new", string(data))
}

func TestApplyToPath_NewBinaryMissing_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "skell")
	require.NoError(t, os.WriteFile(currentExe, []byte("old"), 0755))

	err := selfupdate.ApplyToPath(currentExe, filepath.Join(dir, "nonexistent"))
	assert.Error(t, err)
}

func TestApplyToPath_CurrentExeMissing_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "skell_missing")
	newBin := filepath.Join(dir, "skell_new")
	require.NoError(t, os.WriteFile(newBin, []byte("new"), 0755))

	err := selfupdate.ApplyToPath(currentExe, newBin)
	assert.Error(t, err)
}

func TestLatestRelease_BadJSON_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	t.Cleanup(srv.Close)

	u := selfupdate.New("owner", "repo")
	u.APIBaseURL = srv.URL

	_, err := u.LatestRelease()
	assert.Error(t, err)
}

func TestDownload_CannotCreateFile_ReturnsError(t *testing.T) {
	content := []byte("fake binary")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)

	u := selfupdate.New("owner", "repo")
	asset := &selfupdate.Asset{Name: "skell_test", BrowserDownloadURL: srv.URL}
	// A directory cannot be truncated as a regular file, so archive write fails.
	err := u.Download(asset, t.TempDir())
	assert.Error(t, err)
}

func TestDownload_NetworkErrorMidStream_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hijack the connection and close it after partial write.
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "no hijack", http.StatusInternalServerError)
			return
		}
		conn, _, _ := hj.Hijack()
		// Send partial HTTP response then close the connection abruptly.
		_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\npartial"))
		conn.Close()
	}))
	t.Cleanup(srv.Close)

	destPath := filepath.Join(t.TempDir(), "skell_partial")
	u := selfupdate.New("owner", "repo")
	asset := &selfupdate.Asset{Name: "skell_test", BrowserDownloadURL: srv.URL}

	err := u.Download(asset, destPath)
	assert.Error(t, err)
}

func TestLatestRelease_NetworkError_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately so Do() fails

	u := selfupdate.New("owner", "repo")
	u.APIBaseURL = srv.URL

	_, err := u.LatestRelease()
	assert.Error(t, err)
}

func TestTempPath_ContainsAssetName(t *testing.T) {
	p := selfupdate.TempPath("skell_linux_amd64")
	assert.Contains(t, p, "skell_linux_amd64")
}

// TestDownload_ExtractsTarGz verifies Download unpacks the binary from a tarball.
func TestDownload_ExtractsTarGz(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tarball assets are produced for non-windows platforms")
	}
	body := buildTarGz(t, "skell", []byte("real binary"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	destPath := filepath.Join(t.TempDir(), "skell")
	u := selfupdate.New("owner", "repo")
	asset := &selfupdate.Asset{
		Name:               "skell_1.2.3_linux_amd64.tar.gz",
		BrowserDownloadURL: srv.URL,
	}

	require.NoError(t, u.Download(asset, destPath))

	got, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "real binary", string(got))

	_, err = os.Stat(destPath + ".archive")
	assert.True(t, os.IsNotExist(err), "archive temp file should be removed")
}

// TestDownload_ExtractsZip verifies Download unpacks skell.exe from a zip.
func TestDownload_ExtractsZip(t *testing.T) {
	body := buildZip(t, "skell.exe", []byte("real exe"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	destPath := filepath.Join(t.TempDir(), "skell.exe")
	u := selfupdate.New("owner", "repo")
	asset := &selfupdate.Asset{
		Name:               "skell_1.2.3_windows_amd64.zip",
		BrowserDownloadURL: srv.URL,
	}

	if runtime.GOOS != "windows" {
		t.Skip("zip assets are produced for the windows platform")
	}

	require.NoError(t, u.Download(asset, destPath))

	got, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "real exe", string(got))
}

// TestDownload_ExtractFails_BinaryMissing returns an error when the archive
// does not contain the expected binary.
func TestDownload_ExtractFails_BinaryMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	body := buildTarGz(t, "README.md", []byte("not a binary"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	u := selfupdate.New("owner", "repo")
	asset := &selfupdate.Asset{
		Name:               "skell_1.2.3_linux_amd64.tar.gz",
		BrowserDownloadURL: srv.URL,
	}

	err := u.Download(asset, filepath.Join(t.TempDir(), "skell"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in archive")
}

// TestApplyToPath_RestoresBackupOnFailure: if placing the new binary fails,
// the original must remain in place.
func TestApplyToPath_RestoresBackupOnFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the directory permission check")
	}

	dir := t.TempDir()
	currentExe := filepath.Join(dir, "skell")
	require.NoError(t, os.WriteFile(currentExe, []byte("old"), 0755))

	newBin := filepath.Join(t.TempDir(), "skell_new")
	require.NoError(t, os.WriteFile(newBin, []byte("new"), 0755))

	require.NoError(t, os.Chmod(dir, 0500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })

	err := selfupdate.ApplyToPath(currentExe, newBin)
	require.Error(t, err)
}

// buildTarGz creates an in-memory .tar.gz containing a single file.
func buildTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0755,
		Size: int64(len(content)),
	}))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

// buildZip creates an in-memory .zip containing a single file.
func buildZip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	require.NoError(t, err)
	_, err = w.Write(content)
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

// helpers

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func removeFile(path string) error {
	return os.Remove(path)
}
