package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbedded_IsValid(t *testing.T) {
	c := Embedded()
	require.NotEmpty(t, c.Sources)
	s, ok := c.Find("anthropic")
	require.True(t, ok)
	assert.Equal(t, "https://github.com/anthropics/skills", s.URL)
}

func TestParse_DropsInvalidEntries(t *testing.T) {
	c, err := Parse([]byte(`{"version":1,"sources":[
		{"id":"ok","url":"https://github.com/a/b"},
		{"id":"Bad ID","url":"https://github.com/a/b"},
		{"id":"http","url":"http://github.com/a/b"},
		{"id":"ok","url":"https://github.com/dup/dup"}]}`))
	require.NoError(t, err)
	require.Len(t, c.Sources, 1)
	assert.Equal(t, "ok", c.Sources[0].ID)

	_, err = Parse([]byte(`{"sources":[]}`))
	assert.Error(t, err)
}

func TestLoad_RemoteThenCacheThenEmbedded(t *testing.T) {
	body := `{"version":1,"sources":[{"id":"remote-one","url":"https://github.com/x/y"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	dir := t.TempDir()

	c := Load(context.Background(), Options{CacheDir: dir, URL: srv.URL})
	assert.Equal(t, "remote", c.Origin)
	_, ok := c.Find("remote-one")
	assert.True(t, ok)

	// Fresh cache is used without hitting the network.
	c = Load(context.Background(), Options{CacheDir: dir, URL: "http://127.0.0.1:1/unreachable"})
	assert.Equal(t, "cache", c.Origin)

	// Stale cache + failing network → stale cache.
	old := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(filepath.Join(dir, "catalog.json"), old, old))
	c = Load(context.Background(), Options{CacheDir: dir, URL: "http://127.0.0.1:1/unreachable"})
	assert.Equal(t, "cache", c.Origin)

	// No cache, offline → embedded.
	c = Load(context.Background(), Options{CacheDir: t.TempDir(), Offline: true})
	assert.Equal(t, "embedded", c.Origin)
}

func TestSearch(t *testing.T) {
	c := Embedded()
	assert.NotEmpty(t, c.Search("dotnet"))
	assert.Empty(t, c.Search("no-such-thing-xyz"))
	assert.Len(t, c.Search(""), len(c.Sources))
}
