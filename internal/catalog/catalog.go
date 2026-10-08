// Package catalog provides a curated list of well-known skill sources so a
// fresh Skell install has something to browse ("skell catalog") and users can
// add a source by short id ("skell add anthropic").
//
// The list ships embedded in the binary and is refreshed from the Skell
// repository at most once a day; any network or parse failure falls back to
// the cached or embedded copy, so the catalog always works offline.
package catalog

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DefaultURL is where the maintained catalog lives. It is the same file that
// is embedded at build time.
const DefaultURL = "https://raw.githubusercontent.com/aminmesbahi/skell/main/internal/catalog/index.json"

const (
	maxBytes = 1 << 20
	maxAge   = 24 * time.Hour
	timeout  = 4 * time.Second
)

//go:embed index.json
var embedded []byte

// Source is one well-known skill source.
type Source struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
}

// Catalog is the parsed catalog file.
type Catalog struct {
	Version int      `json:"version"`
	Sources []Source `json:"sources"`
	// Origin says where this copy came from: "remote", "cache" or "embedded".
	Origin string `json:"-"`
}

// Options controls how Load finds the catalog.
type Options struct {
	// CacheDir is where the downloaded catalog is cached (e.g. ~/.skell).
	// Empty disables caching.
	CacheDir string
	// URL overrides DefaultURL (also settable via SKELL_CATALOG_URL).
	URL string
	// Offline skips the network entirely.
	Offline bool
	// Refresh ignores the cache age and tries the network first.
	Refresh bool
	// HTTPClient is used for the download (nil = default client).
	HTTPClient *http.Client
}

var idRx = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

// Parse decodes and validates catalog JSON. Invalid entries (bad id, non-https
// URL) are dropped rather than failing the whole catalog.
func Parse(data []byte) (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("catalog: %w", err)
	}
	valid := c.Sources[:0]
	seen := map[string]bool{}
	for _, s := range c.Sources {
		s.ID = strings.ToLower(strings.TrimSpace(s.ID))
		if !idRx.MatchString(s.ID) || seen[s.ID] {
			continue
		}
		u, err := url.Parse(s.URL)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			continue
		}
		seen[s.ID] = true
		valid = append(valid, s)
	}
	c.Sources = valid
	if len(c.Sources) == 0 {
		return nil, errors.New("catalog: no valid sources")
	}
	return &c, nil
}

// Embedded returns the catalog compiled into the binary.
func Embedded() *Catalog {
	c, err := Parse(embedded)
	if err != nil {
		panic(err) // the embedded file is validated by tests
	}
	c.Origin = "embedded"
	return c
}

// Load returns the freshest catalog available without ever failing: a fresh
// cache, then the network, then a stale cache, then the embedded copy.
func Load(ctx context.Context, opts Options) *Catalog {
	cachePath := ""
	if opts.CacheDir != "" {
		cachePath = filepath.Join(opts.CacheDir, "catalog.json")
	}

	if !opts.Refresh && cachePath != "" {
		if info, err := os.Stat(cachePath); err == nil && time.Since(info.ModTime()) < maxAge {
			if c := readCache(cachePath); c != nil {
				return c
			}
		}
	}

	if !opts.Offline {
		if data, err := download(ctx, opts); err == nil {
			if c, err := Parse(data); err == nil {
				c.Origin = "remote"
				if cachePath != "" {
					_ = os.MkdirAll(filepath.Dir(cachePath), 0o700)
					_ = os.WriteFile(cachePath, data, 0o600)
				}
				return c
			}
		}
	}

	if cachePath != "" {
		if c := readCache(cachePath); c != nil {
			return c
		}
	}
	return Embedded()
}

func readCache(path string) *Catalog {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil
	}
	c, err := Parse(data)
	if err != nil {
		return nil
	}
	c.Origin = "cache"
	return c
}

func download(ctx context.Context, opts Options) ([]byte, error) {
	u := opts.URL
	if u == "" {
		u = os.Getenv("SKELL_CATALOG_URL")
	}
	if u == "" {
		u = DefaultURL
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "skell-cli")
	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxBytes))
}

// Find returns the source with the given id (case-insensitive).
func (c *Catalog) Find(id string) (Source, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, s := range c.Sources {
		if s.ID == id {
			return s, true
		}
	}
	return Source{}, false
}

// Search returns sources whose id, name, description or tags contain query
// (case-insensitive), sorted by id. An empty query returns everything.
func (c *Catalog) Search(query string) []Source {
	q := strings.ToLower(strings.TrimSpace(query))
	var out []Source
	for _, s := range c.Sources {
		hay := strings.ToLower(s.ID + " " + s.Name + " " + s.Description + " " + strings.Join(s.Tags, " "))
		if q == "" || strings.Contains(hay, q) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
