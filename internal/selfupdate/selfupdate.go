// Package selfupdate checks for a newer skell release on GitHub and replaces
// the running executable in-place.
package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const defaultAPIBase = "https://api.github.com"

// Release contains the fields we need from a GitHub release response.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is a single downloadable file attached to a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Updater checks for new versions and downloads the appropriate binary.
type Updater struct {
	Owner      string
	Repo       string
	APIBaseURL string
	HTTPClient *http.Client
}

// New returns an Updater targeting the given GitHub owner/repo.
func New(owner, repo string) *Updater {
	return &Updater{
		Owner:      owner,
		Repo:       repo,
		APIBaseURL: defaultAPIBase,
		HTTPClient: &http.Client{},
	}
}

// LatestRelease queries the GitHub Releases API and returns the latest release.
func (u *Updater) LatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", u.APIBaseURL, u.Owner, u.Repo)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("selfupdate: failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "skell-cli")

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("selfupdate: GitHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("selfupdate: GitHub API returned HTTP %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("selfupdate: failed to parse release JSON: %w", err)
	}
	return &rel, nil
}

// ExpectedAssetName returns the release asset filename for the running OS/arch.
//
//	Windows amd64 → skell_windows_amd64.exe
//	macOS   arm64 → skell_darwin_arm64
//	Linux   amd64 → skell_linux_amd64
func ExpectedAssetName() string {
	name := fmt.Sprintf("skell_%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// FindAsset returns the asset matching the current platform, or nil if none found.
func FindAsset(rel *Release) *Asset {
	want := ExpectedAssetName()
	for i := range rel.Assets {
		if rel.Assets[i].Name == want {
			return &rel.Assets[i]
		}
	}
	return nil
}

// IsNewer returns true when latest represents a higher semantic version than
// current. Both strings may carry an optional leading "v". Parse errors are
// treated as "not newer".
func IsNewer(current, latest string) bool {
	cur := parseSemver(current)
	lat := parseSemver(latest)
	for i := range cur {
		if lat[i] > cur[i] {
			return true
		}
		if lat[i] < cur[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	var out [3]int
	for i, p := range parts {
		if i >= 3 {
			break
		}
		// Strip any pre-release suffix (e.g. "1-beta" → 1).
		p = strings.SplitN(p, "-", 2)[0]
		out[i], _ = strconv.Atoi(p)
	}
	return out
}

// Download fetches the binary at the given URL and writes it to destPath.
func (u *Updater) Download(asset *Asset, destPath string) error {
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, asset.BrowserDownloadURL, nil,
	)
	if err != nil {
		return fmt.Errorf("selfupdate: failed to build download request: %w", err)
	}

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("selfupdate: download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("selfupdate: download returned HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("selfupdate: cannot create temp file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("selfupdate: write failed: %w", err)
	}
	return nil
}

// Apply replaces the currently running executable with the file at newBinaryPath.
//
// On Windows the current exe is renamed to <exe>.old (cannot remove a running
// binary) and the new binary takes its place. On other platforms the old file
// is removed before the rename.
func Apply(newBinaryPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("selfupdate: cannot locate current executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("selfupdate: cannot resolve symlink: %w", err)
	}
	return ApplyToPath(exe, newBinaryPath)
}

// ApplyToPath replaces currentExe with newBinaryPath using the same in-place
// swap logic as Apply, but accepts an explicit current-exe path.
// This function is exported to allow unit tests to exercise the swap logic
// without requiring the tests to replace the running test binary.
func ApplyToPath(currentExe, newBinaryPath string) error {
	if runtime.GOOS == "windows" {
		oldPath := currentExe + ".old"
		_ = os.Remove(oldPath) // remove stale backup from a previous update
		if err := os.Rename(currentExe, oldPath); err != nil {
			return fmt.Errorf("selfupdate: cannot rename current executable: %w", err)
		}
	} else {
		if err := os.Remove(currentExe); err != nil {
			return fmt.Errorf("selfupdate: cannot remove current executable: %w", err)
		}
	}

	if err := os.Rename(newBinaryPath, currentExe); err != nil {
		return fmt.Errorf("selfupdate: cannot place new binary at %s: %w", currentExe, err)
	}

	// Make executable (no-op on Windows but harmless).
	if err := os.Chmod(currentExe, 0755); err != nil { //nolint:gosec
		return fmt.Errorf("selfupdate: cannot set executable permissions: %w", err)
	}
	return nil
}

// TempPath returns a suitable temporary file path for the downloaded binary.
func TempPath(assetName string) string {
	return filepath.Join(os.TempDir(), "skell_update_"+assetName)
}
