// Package guiinstall downloads the Skell desktop app from a GitHub release,
// checks it against the SHA-256 published with it, and unpacks it where
// 'skell gui' will find it.
//
// What is downloaded depends on the system:
//
//	Windows   skell_bundle_<ver>_windows_<arch>.zip   skell.exe + skell-gui.exe
//	macOS     skell_bundle_<ver>_darwin_arm64.zip     skell + Skell.app
//	Linux     skell_gui_<ver>_linux_amd64.tar.gz      Skell
//
// Releases up to v0.15 used older file names; those are found too.
package guiinstall

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/selfupdate"
)

const (
	// maxDownload caps the archive size.
	maxDownload = 400 << 20
	// maxFile caps one unpacked file.
	maxFile = 300 << 20
)

// Choice is the release file to download and the file holding its checksum.
type Choice struct {
	Tag      string
	Asset    selfupdate.Asset
	Checksum selfupdate.Asset
}

// Choose picks the desktop app files for a system from a release. Current
// names are preferred over older ones.
func Choose(rel *selfupdate.Release, goos, goarch string) (*Choice, error) {
	v := strings.TrimPrefix(rel.TagName, "v")
	var names []string
	switch {
	case goos == "windows" && (goarch == "amd64" || goarch == "arm64"):
		names = []string{
			fmt.Sprintf("skell_bundle_%s_windows_%s.zip", v, goarch),
			fmt.Sprintf("skell_%s_windows_%s_bundle.zip", v, goarch),
		}
	case goos == "darwin" && goarch == "arm64":
		names = []string{
			fmt.Sprintf("skell_bundle_%s_darwin_arm64.zip", v),
			fmt.Sprintf("skell_%s_darwin_arm64_bundle.zip", v),
		}
	case goos == "linux" && goarch == "amd64":
		names = []string{
			fmt.Sprintf("skell_gui_%s_linux_amd64.tar.gz", v),
			fmt.Sprintf("Skell_%s_linux_amd64.tar.gz", v),
		}
	default:
		return nil, fmt.Errorf("no desktop app is published for %s/%s", goos, goarch)
	}
	find := func(name string) *selfupdate.Asset {
		for i := range rel.Assets {
			if rel.Assets[i].Name == name {
				return &rel.Assets[i]
			}
		}
		return nil
	}
	for _, name := range names {
		asset := find(name)
		if asset == nil {
			continue
		}
		sum := find(name + ".sha256")
		if sum == nil {
			return nil, fmt.Errorf("release %s has %s but no checksum for it — refusing to install an unverified download", rel.TagName, name)
		}
		return &Choice{Tag: rel.TagName, Asset: *asset, Checksum: *sum}, nil
	}
	return nil, fmt.Errorf("release %s has no desktop app for %s/%s yet", rel.TagName, goos, goarch)
}

// DefaultDir is where the app is installed for the current user, without
// needing administrator rights.
func DefaultDir(goos string) (string, error) {
	switch goos {
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			return "", errors.New("LOCALAPPDATA is not set")
		}
		return filepath.Join(local, "Programs", "skell"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Applications", "Skell"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "bin"), nil
	}
}

// Installer downloads and unpacks the app.
type Installer struct {
	HTTP *http.Client
	// Progress, when set, is told how many bytes of how many have arrived
	// (total is 0 when unknown).
	Progress func(done, total int64)
}

// NewInstaller returns an Installer whose connections time out when idle but
// whose downloads may take as long as the network needs.
func NewInstaller() *Installer {
	return &Installer{HTTP: &http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 15 * time.Second}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}}}
}

// Install downloads the app for goos/goarch from rel, verifies it, unpacks it
// into dir and returns the path of the app to launch.
func (in *Installer) Install(ctx context.Context, rel *selfupdate.Release, goos, goarch, dir string) (string, error) {
	choice, err := Choose(rel, goos, goarch)
	if err != nil {
		return "", err
	}
	expected, err := in.expectedHash(ctx, choice)
	if err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp("", "skell-gui-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	archive := filepath.Join(tmp, choice.Asset.Name)
	if err := in.download(ctx, choice.Asset, archive, expected); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create %s: %w", dir, err)
	}
	return Extract(goos, archive, dir)
}

// expectedHash reads the hex SHA-256 from a "<hash>  <name>" checksum file.
func (in *Installer) expectedHash(ctx context.Context, c *Choice) (string, error) {
	body, err := in.get(ctx, c.Checksum.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("could not fetch the checksum: %w", err)
	}
	defer func() { _ = body.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(body.Body, 4096))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 1 && len(fields[0]) == 64 {
			if len(fields) == 1 || strings.TrimPrefix(fields[1], "*") == c.Asset.Name {
				return strings.ToLower(fields[0]), nil
			}
		}
	}
	return "", fmt.Errorf("the checksum file has no entry for %s", c.Asset.Name)
}

func (in *Installer) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "skell-cli")
	resp, err := in.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp, nil
}

// download streams asset to path and fails unless its SHA-256 is expected.
func (in *Installer) download(ctx context.Context, asset selfupdate.Asset, path, expected string) error {
	resp, err := in.get(ctx, asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600) //nolint:gosec
	if err != nil {
		return err
	}
	total := resp.ContentLength
	if total <= 0 {
		total = asset.Size
	}
	h := sha256.New()
	var done int64
	buf := make([]byte, 64<<10)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			done += int64(n)
			if done > maxDownload {
				_ = out.Close()
				return fmt.Errorf("download is larger than %d MB", maxDownload>>20)
			}
			h.Write(buf[:n])
			if _, werr := out.Write(buf[:n]); werr != nil {
				_ = out.Close()
				return werr
			}
			if in.Progress != nil {
				in.Progress(done, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			_ = out.Close()
			return fmt.Errorf("download interrupted: %w", rerr)
		}
	}
	if err := out.Close(); err != nil {
		return err
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != expected {
		return fmt.Errorf("the download does not match its published checksum (expected %s, got %s) — nothing was installed", expected, got)
	}
	return nil
}

// Extract unpacks a downloaded archive into dir and returns the app to launch.
func Extract(goos, archivePath, dir string) (string, error) {
	switch goos {
	case "windows":
		return extractWindows(archivePath, dir)
	case "darwin":
		return extractDarwin(archivePath, dir)
	default:
		return extractLinux(archivePath, dir)
	}
}

// within returns the path of name inside dir, or an error when name would
// land anywhere else (absolute paths, "..", drive letters).
func within(dir, name string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." || filepath.VolumeName(clean) != "" {
		return "", fmt.Errorf("archive entry %q is outside the install folder", name)
	}
	return filepath.Join(dir, clean), nil
}

// replaceFile writes r to path. An existing file (even a program that is
// running, which Windows will not let anyone overwrite) is moved aside first.
func replaceFile(path string, r io.Reader, mode os.FileMode) error {
	old := ""
	if _, err := os.Lstat(path); err == nil {
		old = path + ".old"
		_ = os.Remove(old)
		if err := os.Rename(path, old); err != nil {
			return fmt.Errorf("cannot replace %s — is it still running? Close it and try again: %w", filepath.Base(path), err)
		}
	}
	err := writeNew(path, r, mode)
	switch {
	case err != nil && old != "":
		_ = os.Rename(old, path) // keep what was there before
	case err == nil && old != "":
		_ = os.Remove(old) // best effort: a running program can keep its old file
	}
	return err
}

func writeNew(path string, r io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode) //nolint:gosec
	if err != nil {
		return err
	}
	n, err := io.Copy(out, io.LimitReader(r, maxFile+1))
	if err == nil && n > maxFile {
		err = errors.New("a file in the archive is too large")
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(path)
	}
	return err
}

func extractWindows(archivePath, dir string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("cannot open the download: %w", err)
	}
	defer func() { _ = zr.Close() }()

	want := map[string]bool{"skell.exe": true, "skell-gui.exe": true}
	found := map[string]bool{}
	for _, zf := range zr.File {
		base := strings.ToLower(filepath.Base(filepath.FromSlash(zf.Name)))
		if zf.FileInfo().IsDir() || !want[base] || found[base] {
			continue
		}
		dest, err := within(dir, base)
		if err != nil {
			return "", err
		}
		rc, err := zf.Open()
		if err != nil {
			return "", err
		}
		err = replaceFile(dest, rc, 0o755)
		_ = rc.Close()
		if err != nil {
			return "", err
		}
		found[base] = true
	}
	if !found["skell-gui.exe"] {
		return "", errors.New("the download does not contain the desktop app (skell-gui.exe)")
	}
	return filepath.Join(dir, "skell-gui.exe"), nil
}

func extractDarwin(archivePath, dir string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("cannot open the download: %w", err)
	}
	defer func() { _ = zr.Close() }()

	app := filepath.Join(dir, "Skell.app")
	hasApp := false
	for _, zf := range zr.File {
		if strings.HasPrefix(zf.Name, "Skell.app/") {
			hasApp = true
		}
	}
	if !hasApp {
		return "", errors.New("the download does not contain Skell.app")
	}
	// Replace the whole bundle so no file of an older version is left behind.
	if err := os.RemoveAll(app); err != nil {
		return "", fmt.Errorf("cannot replace the existing Skell.app: %w", err)
	}
	for _, zf := range zr.File {
		name := zf.Name
		if name != "skell" && !strings.HasPrefix(name, "Skell.app/") {
			continue
		}
		dest, err := within(dir, name)
		if err != nil {
			return "", err
		}
		mode := zf.Mode()
		switch {
		case zf.FileInfo().IsDir():
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return "", err
			}
		case mode&os.ModeSymlink != 0:
			rc, err := zf.Open()
			if err != nil {
				return "", err
			}
			target, rerr := io.ReadAll(io.LimitReader(rc, 4096))
			_ = rc.Close()
			if rerr != nil {
				return "", rerr
			}
			if err := checkLink(dir, dest, string(target)); err != nil {
				return "", err
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return "", err
			}
			if err := os.Symlink(string(target), dest); err != nil {
				return "", err
			}
		default:
			rc, err := zf.Open()
			if err != nil {
				return "", err
			}
			perm := mode.Perm() | 0o600
			err = replaceFile(dest, rc, perm)
			_ = rc.Close()
			if err != nil {
				return "", err
			}
		}
	}
	return app, nil
}

// checkLink accepts a symbolic link only when it stays inside dir.
func checkLink(dir, link, target string) error {
	if filepath.IsAbs(target) || strings.HasPrefix(filepath.ToSlash(target), "/") {
		return fmt.Errorf("archive link %q points outside the install folder", target)
	}
	resolved := filepath.Join(filepath.Dir(link), filepath.FromSlash(target))
	rel, err := filepath.Rel(dir, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("archive link %q points outside the install folder", target)
	}
	return nil
}

func extractLinux(archivePath, dir string) (string, error) {
	f, err := os.Open(archivePath) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("cannot open the download: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != "Skell" {
			continue
		}
		dest := filepath.Join(dir, "skell-gui")
		if err := replaceFile(dest, tr, 0o755); err != nil {
			return "", err
		}
		return dest, nil
	}
	return "", errors.New("the download does not contain the desktop app (Skell)")
}

// Supported reports whether a desktop app is published for this system.
func Supported(goos, goarch string) bool {
	_, err := Choose(&selfupdate.Release{TagName: "v0.0.0"}, goos, goarch)
	return err == nil || !strings.HasPrefix(err.Error(), "no desktop app is published")
}

// Here is Supported for the running system.
func Here() bool { return Supported(runtime.GOOS, runtime.GOARCH) }
