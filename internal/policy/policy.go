// Package policy enforces allowed-registries and other enterprise controls.
package policy

import (
	"errors"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config mirrors the [policy] block in ~/.skell/config.toml.
type Config struct {
	AllowedRegistries []string `toml:"allowed-registries"`
	BlockUnlisted     bool     `toml:"block-unlisted"`
	// RequireValidation, when true, makes install/upgrade validate a skill
	// against the Agent Skills spec before writing it and refuse on errors.
	RequireValidation bool `toml:"require-validation"`

	// invalid is set by Invalid when the policy file could not be parsed; every
	// registry is then refused so a typo can never silently disable the policy.
	invalid error
}

type configFile struct {
	Policy Config `toml:"policy"`
}

// Invalid returns a fail-closed policy that rejects every registry, carrying
// the reason the real policy could not be loaded.
func Invalid(err error) *Config {
	return &Config{BlockUnlisted: true, invalid: err}
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

// normalizeURL canonicalises a registry URL for comparison so that trivially
// different spellings (case, trailing slash, ".git" suffix) of the same
// repository match.
func normalizeURL(u string) string {
	u = strings.ToLower(strings.TrimSpace(u))
	u = strings.TrimRight(u, "/")
	u = strings.TrimSuffix(u, ".git")
	return strings.TrimRight(u, "/")
}

// CheckRegistry returns an error if the given registry URL is blocked by policy.
func (c *Config) CheckRegistry(url string) error {
	if c.invalid != nil {
		return errors.New("policy: config.toml is invalid, refusing all registries: " + c.invalid.Error())
	}
	if !c.BlockUnlisted {
		return nil
	}
	want := normalizeURL(url)
	for _, allowed := range c.AllowedRegistries {
		if normalizeURL(allowed) == want {
			return nil
		}
	}
	return errors.New("policy: registry not in allowed list: " + url)
}
