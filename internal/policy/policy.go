// Package policy enforces allowed-registries and other enterprise controls.
package policy

// Config mirrors the [policy] block in ~/.skell/config.toml.
type Config struct {
	AllowedRegistries []string `toml:"allowed-registries"`
	BlockUnlisted     bool     `toml:"block-unlisted"`
}

// Read parses the policy config from the given file path.
func Read(path string) (*Config, error) {
	// TODO: implement
	panic("not implemented")
}

// CheckRegistry returns an error if the given registry URL is blocked by policy.
func (c *Config) CheckRegistry(url string) error {
	// TODO: implement
	panic("not implemented")
}
