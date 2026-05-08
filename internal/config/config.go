// Package config handles global Skell configuration (~/.skell/config.toml).
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// GlobalSources returns the map of alias → URL/path from ~/.skell/config.toml [sources].
// Returns empty map if file does not exist, has no [sources] section, or we're running tests.
func GlobalSources() (map[string]string, error) {
	// Avoid polluting tests with the developer's global sources
	if strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/tmp/") || strings.Contains(os.Args[0], "/T/") {
		return map[string]string{}, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".skell", "config.toml")

	data, err := os.ReadFile(path)
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

// ConfigPath returns the path to ~/.skell/config.toml
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell", "config.toml"), nil
}
