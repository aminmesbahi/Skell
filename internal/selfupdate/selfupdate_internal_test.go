package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFromTarGz_Success(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, buildTarGzArchive(t, "wanted-bin", []byte("tar content")), 0600))

	destPath := filepath.Join(t.TempDir(), "skell")
	require.NoError(t, extractFromTarGz(archivePath, "wanted-bin", destPath))

	data, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "tar content", string(data))
}

func TestExtractFromTarGz_BinaryMissing(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, buildTarGzArchive(t, "other", []byte("tar content")), 0600))

	err := extractFromTarGz(archivePath, "wanted-bin", filepath.Join(t.TempDir(), "skell"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in archive")
}

func TestExtractFromTarGz_InvalidArchive(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, []byte("not gzip"), 0600))

	err := extractFromTarGz(archivePath, "wanted-bin", filepath.Join(t.TempDir(), "skell"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid gzip stream")
}

func TestExtractFromZip_Success(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.zip")
	require.NoError(t, os.WriteFile(archivePath, buildZipArchive(t, "wanted-bin", []byte("zip content")), 0600))

	destPath := filepath.Join(t.TempDir(), "skell.exe")
	require.NoError(t, extractFromZip(archivePath, "wanted-bin", destPath))

	data, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "zip content", string(data))
}

func TestExtractFromZip_BinaryMissing(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.zip")
	require.NoError(t, os.WriteFile(archivePath, buildZipArchive(t, "other", []byte("zip content")), 0600))

	err := extractFromZip(archivePath, "wanted-bin", filepath.Join(t.TempDir(), "skell.exe"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in archive")
}

func TestExtractFromZip_InvalidArchive(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.zip")
	require.NoError(t, os.WriteFile(archivePath, []byte("not zip"), 0600))

	err := extractFromZip(archivePath, "wanted-bin", filepath.Join(t.TempDir(), "skell.exe"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot open zip")
}

func TestExtractSkellBinary_BareBinaryFallsBackToMove(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.bin")
	require.NoError(t, os.WriteFile(archivePath, []byte("plain binary"), 0600))

	destPath := filepath.Join(t.TempDir(), "skell")
	require.NoError(t, extractSkellBinary(archivePath, "skell_linux_amd64", destPath))

	data, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "plain binary", string(data))
}

func TestMoveFile_Success(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.bin")
	dst := filepath.Join(t.TempDir(), "dst.bin")
	require.NoError(t, os.WriteFile(src, []byte("moved"), 0600))

	require.NoError(t, moveFile(src, dst))
	_, err := os.Stat(src)
	assert.True(t, os.IsNotExist(err))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "moved", string(data))
}

func TestMoveFile_FallsBackWhenDestinationExists(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.bin")
	dst := filepath.Join(t.TempDir(), "dst.bin")
	require.NoError(t, os.WriteFile(src, []byte("moved"), 0600))
	require.NoError(t, os.WriteFile(dst, []byte("old"), 0600))

	require.NoError(t, moveFile(src, dst))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "moved", string(data))
}

func TestMoveFile_SourceMissing(t *testing.T) {
	err := moveFile(filepath.Join(t.TempDir(), "missing.bin"), filepath.Join(t.TempDir(), "dst.bin"))
	assert.Error(t, err)
}

func TestCopyFileContents_Success(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.bin")
	dst := filepath.Join(t.TempDir(), "dst.bin")
	require.NoError(t, os.WriteFile(src, []byte("copied"), 0600))

	require.NoError(t, copyFileContents(src, dst))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "copied", string(data))
}

func TestCopyFileContents_SourceMissing(t *testing.T) {
	err := copyFileContents(filepath.Join(t.TempDir(), "missing.bin"), filepath.Join(t.TempDir(), "dst.bin"))
	assert.Error(t, err)
}

func TestPlaceBinary_Success(t *testing.T) {
	src := filepath.Join(t.TempDir(), "new.bin")
	dst := filepath.Join(t.TempDir(), "current.bin")
	require.NoError(t, os.WriteFile(src, []byte("replacement"), 0600))

	require.NoError(t, placeBinary(src, dst))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "replacement", string(data))
}

func TestApplyToPath_BackupRenameFailure(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "skell.exe")
	newBin := filepath.Join(dir, "skell-new.exe")
	backupPath := currentExe + ".old"
	require.NoError(t, os.WriteFile(currentExe, []byte("old"), 0755))
	require.NoError(t, os.WriteFile(newBin, []byte("new"), 0755))
	require.NoError(t, os.MkdirAll(backupPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(backupPath, "blocker"), []byte("x"), 0600))

	err := ApplyToPath(currentExe, newBin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot back up current executable")

	data, readErr := os.ReadFile(currentExe)
	require.NoError(t, readErr)
	assert.Equal(t, "old", string(data))
}

func TestPlaceBinary_FallsBackWhenDestinationExists(t *testing.T) {
	src := filepath.Join(t.TempDir(), "new.bin")
	dst := filepath.Join(t.TempDir(), "current.bin")
	require.NoError(t, os.WriteFile(src, []byte("replacement"), 0600))
	require.NoError(t, os.WriteFile(dst, []byte("old"), 0600))

	require.NoError(t, placeBinary(src, dst))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "replacement", string(data))
}

func TestExtractSkellBinary_DispatchesZip(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.zip")
	require.NoError(t, os.WriteFile(archivePath, buildZipArchive(t, binaryNameInArchive(), []byte("zip dispatch")), 0600))
	destPath := filepath.Join(t.TempDir(), "skell.exe")

	require.NoError(t, extractSkellBinary(archivePath, "asset.zip", destPath))
	data, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "zip dispatch", string(data))
}

func TestExtractSkellBinary_DispatchesTarGz(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "asset.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, buildTarGzArchive(t, binaryNameInArchive(), []byte("tar dispatch")), 0600))
	destPath := filepath.Join(t.TempDir(), "skell")

	require.NoError(t, extractSkellBinary(archivePath, "asset.tar.gz", destPath))
	data, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "tar dispatch", string(data))
}

func buildTarGzArchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(content))}))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func buildZipArchive(t *testing.T, name string, content []byte) []byte {
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
