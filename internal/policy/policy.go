// Package policy enforces allowed-registries and other enterprise controls.
package policy

import (
	"errors"
	"os"
	"slices"

	"github.com/BurntSushi/toml"
)

// Config mirrors the [policy] block in ~/.skell/config.toml.
type Config struct {
	AllowedRegistries []string `toml:"allowed-registries"`
	BlockUnlisted     bool     `toml:"block-unlisted"`
	// RequireValidation, when true, makes install/upgrade validate a skill
	// against the Agent Skills spec before writing it and refuse on errors.
	RequireValidation bool `toml:"require-validation"`
}

type configFile struct {
	Policy Config `toml:"policy"`
}

// Read parses the policy config from the given file path.
func Read(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cf configFile
	if _, err := toml.Decode(string(data), &cf); err != nil {
		return nil, err
	}
	return &cf.Policy, nil
}

// CheckRegistry returns an error if the given registry URL is blocked by policy.
func (c *Config) CheckRegistry(url string) error {
	if !c.BlockUnlisted {
		return nil
	}
	if slices.Contains(c.AllowedRegistries, url) {
		return nil
	}
	return errors.New("policy: registry not in allowed list: " + url)
}
