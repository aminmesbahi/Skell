// Package config handles global Skell configuration (~/.skell/config.toml).
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// SourcesFrom returns the map of alias → URL/path from the [sources] section of
// the config.toml at the given Skell home root (e.g. ~/.skell). An empty root
// disables global sources entirely — this is how callers (notably tests with an
// isolated Engine) opt out without relying on filesystem heuristics. A missing
// file or absent [sources] section yields an empty map, not an error.
func SourcesFrom(root string) (map[string]string, error) {
	if root == "" {
		return map[string]string{}, nil
	}
	data, err := os.ReadFile(filepath.Join(root, "config.toml"))
	if err != nil {
		return map[string]string{}, nil // no config file yet
	}

	type file struct {
		Sources map[string]string `toml:"sources"`
	}
	var f file
	if _, err := toml.Decode(string(data), &f); err != nil {
		return nil, err
	}
	if f.Sources == nil {
		return map[string]string{}, nil
	}
	return f.Sources, nil
}

// DefaultRoot returns the default Skell home directory (~/.skell), honouring the
// SKELL_HOME override when set.
func DefaultRoot() (string, error) {
	if h := os.Getenv("SKELL_HOME"); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell"), nil
}

// GlobalSources returns the global sources from the default Skell home root.
func GlobalSources() (map[string]string, error) {
	root, err := DefaultRoot()
	if err != nil {
		return nil, err
	}
	return SourcesFrom(root)
}

// Path returns the path to the config.toml under the default Skell home root.
func Path() (string, error) {
	root, err := DefaultRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config.toml"), nil
}
