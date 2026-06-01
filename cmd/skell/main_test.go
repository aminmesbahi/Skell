package skell

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain isolates the whole CLI test suite from the developer's real ~/.skell
// by pointing SKELL_HOME at an empty temp directory. This keeps global sources
// and the registry cache hermetic regardless of the host machine's config.
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "skell-cli-test-home-")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("SKELL_HOME", filepath.Clean(tmp))
	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}
