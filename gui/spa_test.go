package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSPAFallbackServesShellForClientRoutes(t *testing.T) {
	var seen string
	h := spaFallback(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r.URL.Path }))
	for path, want := range map[string]string{
		"/":                          "/",
		"/projects":                  "/",
		"/projects/project-1/skills": "/",
		"/skills/some-skill":         "/",
		"/assets/index-abc.js":       "/assets/index-abc.js",
		"/favicon.svg":               "/favicon.svg",
		"/wails/runtime":             "/wails/runtime",
	} {
		seen = ""
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
		if seen != want {
			t.Errorf("%s was served as %q, want %q", path, seen, want)
		}
	}
}
