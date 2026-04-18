package selfupdate_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
	name := selfupdate.ExpectedAssetName()
	assert.Contains(t, name, "skell_")
	// Must start with "skell_" and contain the GOOS/GOARCH somewhere.
	assert.Regexp(t, `^skell_[a-z]+_[a-z0-9]+`, name)
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
	assetName := selfupdate.ExpectedAssetName()
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

// helpers

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func removeFile(path string) error {
	return os.Remove(path)
}
