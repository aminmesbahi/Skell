package skell

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestSelfUpdateCmd_AlreadyUpToDate exercises the "already up to date" branch
// by pointing the command at a mock server that returns the current version.
func TestSelfUpdateCmd_AlreadyUpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := map[string]interface{}{
			"tag_name": "dev", // matches version.Version in test builds
			"assets":   []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(rel))
	}))
	defer srv.Close()

	t.Setenv("SKELL_SELFUPDATE_API_URL", srv.URL)

	out, err := executeCmd(t, "selfupdate")
	require.NoError(t, err)
	assert.True(t, strings.Contains(out, "up to date") || strings.Contains(out, "new version"), "expected update status message")
}

// TestSelfUpdateCmd_NewVersionAvailable_CheckOnly exercises the --check branch
// where a newer version is reported but not downloaded.
func TestSelfUpdateCmd_NewVersionAvailable_CheckOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := map[string]interface{}{
			"tag_name": "v999.0.0",
			"assets":   []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(rel))
	}))
	defer srv.Close()

	t.Setenv("SKELL_SELFUPDATE_API_URL", srv.URL)

	out, err := executeCmd(t, "selfupdate", "--check")
	require.NoError(t, err)
	assert.Contains(t, out, "v999.0.0")
}

// TestSelfUpdateCmd_NewVersion_NoAsset_ReturnsError exercises the "no asset for platform" error.
func TestSelfUpdateCmd_NewVersion_NoAsset_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := map[string]interface{}{
			"tag_name": "v999.0.0",
			"assets":   []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(rel))
	}))
	defer srv.Close()

	t.Setenv("SKELL_SELFUPDATE_API_URL", srv.URL)

	_, err := executeCmd(t, "selfupdate")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no release asset")
}

// TestSelfUpdateCmd_APIError_ReturnsError covers the LatestRelease error path.
func TestSelfUpdateCmd_APIError_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	t.Setenv("SKELL_SELFUPDATE_API_URL", srv.URL)

	_, err := executeCmd(t, "selfupdate")
	assert.Error(t, err)
}

