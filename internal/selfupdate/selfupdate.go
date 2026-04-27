// Package selfupdate checks for a newer skell release on GitHub and replaces
// the running executable in-place.
package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	"time"
)

const defaultAPIBase = "https://api.github.com"

// httpTimeout caps a single HTTP request to the GitHub API or asset download.
const httpTimeout = 30 * time.Second

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
		HTTPClient: &http.Client{Timeout: httpTimeout},
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

// ExpectedAssetName returns the release asset filename for the given release
// tag on the running OS/arch. The leading "v" of version is stripped to match
// goreleaser output (skell_<ver>_<os>_<arch>.tar.gz, .zip on Windows).
func ExpectedAssetName(version string) string {
	v := strings.TrimPrefix(version, "v")
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("skell_%s_windows_%s.zip", v, runtime.GOARCH)
	}
	return fmt.Sprintf("skell_%s_%s_%s.tar.gz", v, runtime.GOOS, runtime.GOARCH)
}

// binaryNameInArchive returns the file name of the skell binary inside a
// release archive.
func binaryNameInArchive() string {
	if runtime.GOOS == "windows" {
		return "skell.exe"
	}
	return "skell"
}

// FindAsset returns the release asset matching the current platform, or nil
// if none is found.
func FindAsset(rel *Release) *Asset {
	want := ExpectedAssetName(rel.TagName)
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

// Download fetches the release archive at the given URL, extracts the skell
// binary from it, and writes the binary to destPath. The archive is removed
// after extraction.
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

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("selfupdate: cannot prepare destination dir: %w", err)
	}

	archivePath := destPath + ".archive"
	archive, err := os.Create(archivePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("selfupdate: cannot create temp file: %w", err)
	}
	if _, err := io.Copy(archive, resp.Body); err != nil {
		_ = archive.Close()
		_ = os.Remove(archivePath)
		return fmt.Errorf("selfupdate: write failed: %w", err)
	}
	if err := archive.Close(); err != nil {
		_ = os.Remove(archivePath)
		return fmt.Errorf("selfupdate: close failed: %w", err)
	}
	defer func() { _ = os.Remove(archivePath) }()

	if err := extractSkellBinary(archivePath, asset.Name, destPath); err != nil {
		_ = os.Remove(destPath)
		return err
	}
	return nil
}

// extractSkellBinary unpacks the skell binary out of a release archive.
// Supports .tar.gz and .zip; bare-binary fallback is used for any other name.
func extractSkellBinary(archivePath, assetName, destPath string) error {
	binName := binaryNameInArchive()
	switch {
	case strings.HasSuffix(assetName, ".tar.gz") || strings.HasSuffix(assetName, ".tgz"):
		return extractFromTarGz(archivePath, binName, destPath)
	case strings.HasSuffix(assetName, ".zip"):
		return extractFromZip(archivePath, binName, destPath)
	default:
		return moveFile(archivePath, destPath)
	}
}

func extractFromTarGz(archivePath, binName, destPath string) error {
	f, err := os.Open(archivePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("selfupdate: cannot open archive: %w", err)
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("selfupdate: invalid gzip stream: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("selfupdate: tar read failed: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != binName {
			continue
		}
		out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755) //nolint:gosec
		if err != nil {
			return fmt.Errorf("selfupdate: cannot create binary file: %w", err)
		}
		if _, err := io.Copy(out, io.LimitReader(tr, 200<<20)); err != nil { //nolint:gosec
			_ = out.Close()
			return fmt.Errorf("selfupdate: tar extract failed: %w", err)
		}
		return out.Close()
	}
	return fmt.Errorf("selfupdate: %q not found in archive", binName)
}

func extractFromZip(archivePath, binName, destPath string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("selfupdate: cannot open zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	for _, zf := range zr.File {
		if filepath.Base(zf.Name) != binName {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return fmt.Errorf("selfupdate: cannot open zip entry: %w", err)
		}
		out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755) //nolint:gosec
		if err != nil {
			_ = rc.Close()
			return fmt.Errorf("selfupdate: cannot create binary file: %w", err)
		}
		if _, err := io.Copy(out, io.LimitReader(rc, 200<<20)); err != nil { //nolint:gosec
			_ = rc.Close()
			_ = out.Close()
			return fmt.Errorf("selfupdate: zip extract failed: %w", err)
		}
		_ = rc.Close()
		return out.Close()
	}
	return fmt.Errorf("selfupdate: %q not found in archive", binName)
}

// moveFile relocates src to dst, falling back to a copy across filesystems.
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755) //nolint:gosec
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	_ = os.Remove(src)
	return nil
}

// Apply replaces the currently running executable with the file at newBinaryPath.
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

// ApplyToPath replaces currentExe with newBinaryPath atomically: backup the
// current exe, place the new one (falling back to copy across filesystems),
// and restore the backup on any failure. Exported for tests.
func ApplyToPath(currentExe, newBinaryPath string) error {
	if _, err := os.Stat(newBinaryPath); err != nil {
		return fmt.Errorf("selfupdate: new binary not found: %w", err)
	}
	if _, err := os.Stat(currentExe); err != nil {
		return fmt.Errorf("selfupdate: current executable not found: %w", err)
	}

	backupPath := currentExe + ".old"
	_ = os.Remove(backupPath)

	if err := os.Rename(currentExe, backupPath); err != nil {
		return fmt.Errorf("selfupdate: cannot back up current executable: %w", err)
	}

	if err := placeBinary(newBinaryPath, currentExe); err != nil {
		_ = os.Rename(backupPath, currentExe)
		return fmt.Errorf("selfupdate: cannot place new binary at %s: %w", currentExe, err)
	}

	if err := os.Chmod(currentExe, 0755); err != nil { //nolint:gosec
		return fmt.Errorf("selfupdate: cannot set executable permissions: %w", err)
	}

	// Windows can't unlink a running binary, so keep the .old backup there.
	if runtime.GOOS != "windows" {
		_ = os.Remove(backupPath)
	}
	return nil
}

// placeBinary moves src to dst, falling back to copy+remove when os.Rename
// fails with EXDEV (cross-filesystem rename, e.g. /tmp → /usr/local/bin).
func placeBinary(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyFileContents(src, dst); err != nil {
		return err
	}
	_ = os.Remove(src)
	return nil
}

// copyFileContents copies src to dst with an fsync before close.
func copyFileContents(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755) //nolint:gosec
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// TempPath returns a suitable temporary file path for the downloaded binary.
func TempPath(assetName string) string {
	return filepath.Join(os.TempDir(), "skell_update_"+assetName)
}
