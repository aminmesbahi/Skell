package registry

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRegistryURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "empty", raw: "", wantErr: true},
		{name: "dash", raw: "-bad", wantErr: true},
		{name: "https", raw: "https://example.com/skills.git"},
		{name: "scp", raw: "git@github.com:owner/repo.git"},
		{name: "unsupported scheme", raw: "ftp://example.com/repo", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRegistryURL(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateRegistryURL_LocalPath(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "skills")
	require.NoError(t, os.MkdirAll(localPath, 0755))
	assert.NoError(t, validateRegistryURL(localPath))
}

func TestIsLocalRegistryURL(t *testing.T) {
	assert.True(t, IsLocalRegistryURL(filepath.Join(t.TempDir(), "skills")))
	assert.True(t, IsLocalRegistryURL("file:///tmp/skills"))
	assert.False(t, IsLocalRegistryURL("https://example.com/repo.git"))
}

func TestIsSCPStyleGitURL(t *testing.T) {
	assert.True(t, isSCPStyleGitURL("git@github.com:owner/repo.git"))
	assert.False(t, isSCPStyleGitURL("https://github.com/owner/repo.git"))
	assert.False(t, isSCPStyleGitURL("owner/repo"))
}

func TestAdapterCacheDirAndSourceRoot(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := NewAdapter(cacheRoot)
	assert.Equal(t, filepath.Join(cacheRoot, "default"), adapter.cacheDir("default"))

	local := filepath.Join(t.TempDir(), "skills")
	require.NoError(t, os.MkdirAll(local, 0755))
	assert.Equal(t, local, adapter.sourceRoot(Registry{Alias: "local", URL: local}))
	assert.Equal(t, local, adapter.sourceRoot(Registry{Alias: "local", URL: "file://" + local}))
	assert.Equal(t, filepath.Join(cacheRoot, "remote"), adapter.sourceRoot(Registry{Alias: "remote", URL: "https://example.com/repo.git"}))
}

func TestCopyFile_Success(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.txt")
	dst := filepath.Join(t.TempDir(), "nested", "dst.txt")
	require.NoError(t, os.WriteFile(src, []byte("copied"), 0600))

	require.NoError(t, copyFile(src, dst, 0644))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "copied", string(data))
}

func TestCopyDir_PreservesContents(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "copy")
	require.NoError(t, os.MkdirAll(filepath.Join(src, "nested"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "nested", "SKILL.md"), []byte("---\nname: test\n---\n"), 0600))

	require.NoError(t, copyDir(src, dst))
	data, err := os.ReadFile(filepath.Join(dst, "nested", "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: test")
}

func TestCopyDir_PreservesSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliable on windows without elevated privileges")
	}
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "copy")
	require.NoError(t, os.WriteFile(filepath.Join(src, "target.txt"), []byte("target"), 0600))
	require.NoError(t, os.Symlink("target.txt", filepath.Join(src, "link.txt")))

	require.NoError(t, copyDir(src, dst))
	link, err := os.Readlink(filepath.Join(dst, "link.txt"))
	require.NoError(t, err)
	assert.Equal(t, "target.txt", link)
}

func TestCacheStatus_WithRegistryDir(t *testing.T) {
	cacheRoot := t.TempDir()
	adapter := NewAdapter(cacheRoot)
	regDir := filepath.Join(cacheRoot, "default", "skill-a")
	require.NoError(t, os.MkdirAll(regDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(regDir, "SKILL.md"), []byte("---\nname: skill-a\n---\n"), 0600))

	status, err := adapter.CacheStatus()
	require.NoError(t, err)
	assert.Contains(t, status, "remote registry cache status")
	assert.Contains(t, status, "default")
}

func TestRunGit_Success(t *testing.T) {
	out, err := runGit("--version")
	require.NoError(t, err)
	assert.Contains(t, out, "git version")
}

func TestRunGit_Failure(t *testing.T) {
	_, err := runGit("definitely-not-a-real-git-subcommand")
	assert.Error(t, err)
}

func TestFetch_LocalFileURL_Succeeds(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file:// local source normalization is not portable on windows in the current implementation")
	}
	regDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(regDir, "skill-a"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(regDir, "skill-a", "SKILL.md"), []byte("---\nname: skill-a\n---\n"), 0600))

	adapter := NewAdapter(t.TempDir())
	err := adapter.Fetch(Registry{Alias: "local", URL: "file://" + filepath.ToSlash(regDir)})
	assert.NoError(t, err)
}

func TestFetch_LocalPathMissingReturnsError(t *testing.T) {
	adapter := NewAdapter(t.TempDir())
	err := adapter.Fetch(Registry{Alias: "local", URL: filepath.Join(t.TempDir(), "missing")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "local skill source not found")
}

func TestCacheRefresh_LocalMissingReturnsError(t *testing.T) {
	adapter := NewAdapter(t.TempDir())
	err := adapter.CacheRefresh(Registry{Alias: "local", URL: filepath.Join(t.TempDir(), "missing")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "local skill source not found")
}

func TestCopySkillTo_UnsafeDestinationReturnsError(t *testing.T) {
	regDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(regDir, "skill-a"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(regDir, "skill-a", "SKILL.md"), []byte("---\nname: skill-a\n---\n"), 0600))

	adapter := NewAdapter(t.TempDir())
	err := adapter.CopySkillTo(Registry{Alias: "local", URL: regDir}, "skill-a", "", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe destination")
}
