package signing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skillDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: x\ndescription: x\n---\nbody\n"), 0o644))
	return dir
}

func newStore(t *testing.T) *TrustStore {
	t.Helper()
	ts, err := LoadStore(StorePath(t.TempDir()))
	require.NoError(t, err)
	return ts
}

func TestSignAndVerifyTrusted(t *testing.T) {
	kp, err := GenerateKey("acme")
	require.NoError(t, err)
	dir := skillDir(t)
	_, err = Sign(dir, "acme", kp.Private)
	require.NoError(t, err)

	ts := newStore(t)
	_, err = ts.Add("Acme Inc", kp.Public)
	require.NoError(t, err)

	res := Verify(dir, ts)
	assert.Equal(t, StatusTrusted, res.Status)
	assert.True(t, res.OK())
	assert.Equal(t, "Acme Inc", res.Publisher)
}

func TestVerifyUntrustedKey(t *testing.T) {
	kp, _ := GenerateKey("acme")
	dir := skillDir(t)
	_, err := Sign(dir, "acme", kp.Private)
	require.NoError(t, err)

	res := Verify(dir, newStore(t))
	assert.Equal(t, StatusUntrusted, res.Status)
	assert.False(t, res.OK())
	assert.Equal(t, StatusUntrusted, Verify(dir, nil).Status)
}

func TestVerifyDetectsTamperingAndForgery(t *testing.T) {
	kp, _ := GenerateKey("acme")
	dir := skillDir(t)
	_, err := Sign(dir, "acme", kp.Private)
	require.NoError(t, err)
	ts := newStore(t)
	_, _ = ts.Add("acme", kp.Public)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "extra.md"), []byte("sneaky"), 0o644))
	assert.Equal(t, StatusTampered, Verify(dir, ts).Status)

	// A signature file edited by hand no longer verifies.
	dir2 := skillDir(t)
	sig, err := Sign(dir2, "acme", kp.Private)
	require.NoError(t, err)
	data, _ := os.ReadFile(filepath.Join(dir2, FileName))
	forged := []byte(string(data[:len(data)-1]))
	forged = []byte(replaceOnce(string(forged), sig.ContentHash, "sha256:deadbeef"))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, FileName), forged, 0o644))
	assert.Equal(t, StatusInvalid, Verify(dir2, ts).Status)
}

func replaceOnce(s, old, repl string) string {
	for i := 0; i+len(old) <= len(s); i++ {
		if s[i:i+len(old)] == old {
			return s[:i] + repl + s[i+len(old):]
		}
	}
	return s
}

func TestVerifyUnsignedAndDamaged(t *testing.T) {
	dir := skillDir(t)
	assert.Equal(t, StatusUnsigned, Verify(dir, nil).Status)
	require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), []byte("not json"), 0o644))
	assert.Equal(t, StatusInvalid, Verify(dir, nil).Status)
}

func TestSignRequiresSkillMDAndPublisher(t *testing.T) {
	kp, _ := GenerateKey("a")
	_, err := Sign(t.TempDir(), "a", kp.Private)
	assert.Error(t, err)
	_, err = Sign(skillDir(t), "", kp.Private)
	assert.Error(t, err)
}

func TestKeyFilesRoundTrip(t *testing.T) {
	kp, _ := GenerateKey("acme")
	dir := t.TempDir()
	keyPath, pubPath, err := WriteKeyFiles(dir, kp)
	require.NoError(t, err)

	priv, err := ReadPrivateKey(keyPath)
	require.NoError(t, err)
	assert.Equal(t, kp.Private, priv)

	text, _ := os.ReadFile(pubPath)
	pub, err := DecodePublic(string(text))
	require.NoError(t, err)
	assert.Equal(t, kp.Public, pub)
	pub2, err := DecodePublic(EncodePublic(kp.Public))
	require.NoError(t, err)
	assert.Equal(t, kp.Public, pub2)

	_, _, err = WriteKeyFiles(dir, kp)
	assert.Error(t, err, "never overwrite an existing key")
	_, err = ReadPrivateKey(pubPath)
	assert.Error(t, err)
	_, err = DecodePublic("nonsense")
	assert.Error(t, err)
}

func TestTrustStorePersistsAndRemoves(t *testing.T) {
	path := StorePath(t.TempDir())
	ts, err := LoadStore(path)
	require.NoError(t, err)
	kp, _ := GenerateKey("a")
	_, err = ts.Add("acme", kp.Public)
	require.NoError(t, err)
	_, err = ts.Add("acme-renamed", kp.Public)
	require.NoError(t, err)

	again, err := LoadStore(path)
	require.NoError(t, err)
	require.Len(t, again.Keys, 1)
	assert.Equal(t, "acme-renamed", again.Keys[0].Publisher)
	assert.True(t, again.Trusts(KeyID(kp.Public)))

	_, err = again.Remove("nobody")
	assert.Error(t, err)
	n, err := again.Remove("ACME-renamed")
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.False(t, again.Trusts(KeyID(kp.Public)))
	_, err = again.Add(" ", kp.Public)
	assert.Error(t, err)
}
