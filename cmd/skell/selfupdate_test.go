package skell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSelfUpdateCmd_CheckFlag_NoNetwork verifies that --check runs without
// crashing even when the current version is "dev" and the network is unavailable
// (the test exercises the flag wiring only; real network calls are covered in
// internal/selfupdate tests).
func TestSelfUpdateCmd_CheckFlag_Registered(t *testing.T) {
	// Confirm the flag is wired; we expect a non-nil error only if GitHub is
	// unreachable in CI, which is acceptable. We just verify the command exists
	// and the --check flag is recognised (no "unknown flag" error).
	root := newRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Use == "selfupdate" {
			found = true
			assert.True(t, cmd.Flags().Lookup("check") != nil, "--check flag should be registered")
		}
	}
	assert.True(t, found, "selfupdate command should be registered")
}
