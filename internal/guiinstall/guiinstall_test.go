package guiinstall

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func makeZip(t *testing.T, files map[string]string, links map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		hdr := &zip.FileHeader{Name: name, Method: zip.Deflate}
		hdr.SetMode(0o755)
		w, err := zw.CreateHeader(hdr)
		require.NoError(t, err)
		_, err = w.Write([]byte(body))
		require.NoError(t, err)
	}
	for name, target := range links {
		hdr := &zip.FileHeader{Name: name}
		hdr.SetMode(os.ModeSymlink | 0o755)
		w, err := zw.CreateHeader(hdr)
		require.NoError(t, err)
		_, err = w.Write([]byte(target))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func makeTarGz(t *testing.T, name, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}))
	_, err := tw.Write([]byte(body))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// releaseServer serves one archive and its checksum file and returns a
// release that points at them.
func releaseServer(t *testing.T, tag, assetName string, archive []byte, checksum string) *selfupdate.Release {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/"+assetName, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/"+assetName+".sha256", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(checksum + "\n")) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &selfupdate.Release{TagName: tag, Assets: []selfupdate.Asset{
		{Name: assetName, BrowserDownloadURL: srv.URL + "/" + assetName, Size: int64(len(archive))},
		{Name: assetName + ".sha256", BrowserDownloadURL: srv.URL + "/" + assetName + ".sha256"},
	}}
}

func TestChooseUsesCurrentNamesThenOlderOnes(t *testing.T) {
	rel := &selfupdate.Release{TagName: "v1.2.3", Assets: []selfupdate.Asset{
		{Name: "skell_1.2.3_windows_amd64_bundle.zip"}, {Name: "skell_1.2.3_windows_amd64_bundle.zip.sha256"},
	}}
	c, err := Choose(rel, "windows", "amd64")
	require.NoError(t, err)
	assert.Equal(t, "skell_1.2.3_windows_amd64_bundle.zip", c.Asset.Name, "older release naming")

	rel.Assets = append(rel.Assets, selfupdate.Asset{Name: "skell_bundle_1.2.3_windows_amd64.zip"}, selfupdate.Asset{Name: "skell_bundle_1.2.3_windows_amd64.zip.sha256"})
	c, err = Choose(rel, "windows", "amd64")
	require.NoError(t, err)
	assert.Equal(t, "skell_bundle_1.2.3_windows_amd64.zip", c.Asset.Name, "current naming wins")

	for _, tc := range []struct{ goos, goarch, want string }{
		{"darwin", "arm64", "skell_bundle_1.2.3_darwin_arm64.zip"},
		{"linux", "amd64", "skell_gui_1.2.3_linux_amd64.tar.gz"},
		{"windows", "arm64", "skell_bundle_1.2.3_windows_arm64.zip"},
	} {
		rel := &selfupdate.Release{TagName: "v1.2.3", Assets: []selfupdate.Asset{{Name: tc.want}, {Name: tc.want + ".sha256"}}}
		got, err := Choose(rel, tc.goos, tc.goarch)
		require.NoError(t, err, tc.want)
		assert.Equal(t, tc.want, got.Asset.Name)
	}
}

func TestChooseRefusesWhatItCannotVerifyOrDoesNotExist(t *testing.T) {
	noSum := &selfupdate.Release{TagName: "v1.0.0", Assets: []selfupdate.Asset{{Name: "skell_bundle_1.0.0_windows_amd64.zip"}}}
	_, err := Choose(noSum, "windows", "amd64")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no checksum")

	_, err = Choose(&selfupdate.Release{TagName: "v1.0.0"}, "windows", "amd64")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no desktop app for windows/amd64 yet")

	_, err = Choose(&selfupdate.Release{TagName: "v1.0.0"}, "darwin", "amd64")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no desktop app is published for darwin/amd64")

	assert.True(t, Supported("windows", "amd64"))
	assert.False(t, Supported("freebsd", "amd64"))
	_ = Here()
}

func TestInstallWindowsBundleVerifiesAndUnpacksBoth(t *testing.T) {
	archive := makeZip(t, map[string]string{"skell.exe": "cli", "skell-gui.exe": "gui", "README.txt": "ignored"}, nil)
	rel := releaseServer(t, "v1.0.0", "skell_bundle_1.0.0_windows_amd64.zip", archive, sum(archive)+"  skell_bundle_1.0.0_windows_amd64.zip")
	dir := filepath.Join(t.TempDir(), "Programs", "skell")

	var last [2]int64
	in := NewInstaller()
	in.Progress = func(done, total int64) { last = [2]int64{done, total} }
	got, err := in.Install(context.Background(), rel, "windows", "amd64", dir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "skell-gui.exe"), got)
	assert.Equal(t, int64(len(archive)), last[0])
	assert.Equal(t, int64(len(archive)), last[1])

	b, _ := os.ReadFile(filepath.Join(dir, "skell.exe"))
	assert.Equal(t, "cli", string(b))
	b, _ = os.ReadFile(got)
	assert.Equal(t, "gui", string(b))
	assert.NoFileExists(t, filepath.Join(dir, "README.txt"))
}

func TestInstallRefusesATamperedDownload(t *testing.T) {
	archive := makeZip(t, map[string]string{"skell-gui.exe": "gui"}, nil)
	rel := releaseServer(t, "v1.0.0", "skell_bundle_1.0.0_windows_amd64.zip", archive, sum([]byte("something else"))+"  skell_bundle_1.0.0_windows_amd64.zip")
	dir := t.TempDir()

	_, err := NewInstaller().Install(context.Background(), rel, "windows", "amd64", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not match its published checksum")
	entries, _ := os.ReadDir(dir)
	assert.Empty(t, entries, "nothing may be installed")
}

func TestInstallReportsMissingChecksumEntryAndBadDownloads(t *testing.T) {
	archive := makeZip(t, map[string]string{"skell-gui.exe": "gui"}, nil)
	rel := releaseServer(t, "v1.0.0", "skell_bundle_1.0.0_windows_amd64.zip", archive, sum(archive)+"  some-other-file.zip")
	_, err := NewInstaller().Install(context.Background(), rel, "windows", "amd64", t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no entry for")

	good := releaseServer(t, "v1.0.0", "skell_bundle_1.0.0_windows_amd64.zip", archive, sum(archive))
	good.Assets[0].BrowserDownloadURL += "-missing"
	_, err = NewInstaller().Install(context.Background(), good, "windows", "amd64", t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestWindowsExtractReplacesAnExistingInstallAndKeepsItOnFailure(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skell-gui.exe"), []byte("old"), 0o755))
	zipPath := filepath.Join(t.TempDir(), "b.zip")
	require.NoError(t, os.WriteFile(zipPath, makeZip(t, map[string]string{"skell-gui.exe": "new"}, nil), 0o600))

	got, err := Extract("windows", zipPath, dir)
	require.NoError(t, err)
	b, _ := os.ReadFile(got)
	assert.Equal(t, "new", string(b))
	assert.NoFileExists(t, filepath.Join(dir, "skell-gui.exe.old"))

	noGUI := filepath.Join(t.TempDir(), "n.zip")
	require.NoError(t, os.WriteFile(noGUI, makeZip(t, map[string]string{"skell.exe": "cli"}, nil), 0o600))
	_, err = Extract("windows", noGUI, dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain the desktop app")
}

func TestReplaceFileRestoresTheOldFileWhenTheWriteFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0o755))
	err := replaceFile(path, failingReader{}, 0o755)
	require.Error(t, err)
	b, _ := os.ReadFile(path)
	assert.Equal(t, "old", string(b))
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, os.ErrDeadlineExceeded }

func TestDarwinExtractKeepsBundleLinksAndModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need extra rights on Windows")
	}
	zipPath := filepath.Join(t.TempDir(), "b.zip")
	require.NoError(t, os.WriteFile(zipPath, makeZip(t,
		map[string]string{"skell": "cli", "Skell.app/Contents/MacOS/Skell": "gui", "Skell.app/Contents/Info.plist": "plist"},
		map[string]string{"Skell.app/Contents/Current": "Info.plist"}), 0o600))
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "Skell.app", "Stale"), 0o755))

	app, err := Extract("darwin", zipPath, dir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "Skell.app"), app)
	assert.NoDirExists(t, filepath.Join(dir, "Skell.app", "Stale"), "the old bundle is replaced, not merged")
	info, err := os.Stat(filepath.Join(dir, "Skell.app", "Contents", "MacOS", "Skell"))
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0o100, "the app must stay executable")
	target, err := os.Readlink(filepath.Join(dir, "Skell.app", "Contents", "Current"))
	require.NoError(t, err)
	assert.Equal(t, "Info.plist", target)
	assert.FileExists(t, filepath.Join(dir, "skell"))
}

func TestExtractRejectsPathsAndLinksThatEscape(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(t.TempDir(), "evil.zip")

	require.NoError(t, os.WriteFile(zipPath, makeZip(t, map[string]string{"Skell.app/../../evil": "x", "Skell.app/ok": "y"}, nil), 0o600))
	_, err := Extract("darwin", zipPath, dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the install folder")

	require.NoError(t, os.WriteFile(zipPath, makeZip(t, map[string]string{"Skell.app/ok": "y"}, map[string]string{"Skell.app/link": "../../../etc"}), 0o600))
	_, err = Extract("darwin", zipPath, t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the install folder")

	require.NoError(t, os.WriteFile(zipPath, makeZip(t, map[string]string{"Skell.app/ok": "y"}, map[string]string{"Skell.app/abs": "/etc/passwd"}), 0o600))
	_, err = Extract("darwin", zipPath, t.TempDir())
	require.Error(t, err)

	assert.NoFileExists(t, filepath.Join(filepath.Dir(dir), "evil"))
}

func TestWithin(t *testing.T) {
	dir := t.TempDir()
	got, err := within(dir, "a/b.txt")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "a", "b.txt"), got)
	for _, bad := range []string{"../x", "a/../../x", "..", "."} {
		_, err := within(dir, bad)
		assert.Error(t, err, bad)
	}
	if runtime.GOOS != "windows" {
		_, err := within(dir, "/etc/passwd")
		assert.Error(t, err)
	}
}

func TestLinuxExtractInstallsTheAppAsSkellGui(t *testing.T) {
	tgz := filepath.Join(t.TempDir(), "gui.tar.gz")
	require.NoError(t, os.WriteFile(tgz, makeTarGz(t, "Skell", "elf"), 0o600))
	dir := t.TempDir()
	got, err := Extract("linux", tgz, dir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "skell-gui"), got)
	b, _ := os.ReadFile(got)
	assert.Equal(t, "elf", string(b))

	other := filepath.Join(t.TempDir(), "other.tar.gz")
	require.NoError(t, os.WriteFile(other, makeTarGz(t, "readme", "x"), 0o600))
	_, err = Extract("linux", other, dir)
	assert.Error(t, err)
}

func TestDefaultDir(t *testing.T) {
	t.Setenv("LOCALAPPDATA", `C:\Users\me\AppData\Local`)
	got, err := DefaultDir("windows")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(`C:\Users\me\AppData\Local`, "Programs", "skell"), got)

	t.Setenv("LOCALAPPDATA", "")
	_, err = DefaultDir("windows")
	assert.Error(t, err)

	mac, err := DefaultDir("darwin")
	require.NoError(t, err)
	assert.Contains(t, filepath.ToSlash(mac), "Applications/Skell")
	lin, err := DefaultDir("linux")
	require.NoError(t, err)
	assert.Contains(t, filepath.ToSlash(lin), ".local/bin")
}
