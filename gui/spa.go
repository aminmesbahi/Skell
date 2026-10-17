package main

import (
	"net/http"
	"path"
	"strings"
)

// spaFallback serves the app shell for client-side routes. The frontend uses
// browser-history routing (/projects, /catalog, …), so reloading the window on
// one of those routes asks the asset server for a path that isn't a file; without
// this the webview would show an empty 404 page instead of the app.
func spaFallback(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p != "/" && path.Ext(p) == "" && !strings.HasPrefix(p, "/wails/") {
			r2 := r.Clone(r.Context())
			u := *r.URL
			u.Path = "/"
			r2.URL = &u
			r = r2
		}
		next.ServeHTTP(w, r)
	})
}
