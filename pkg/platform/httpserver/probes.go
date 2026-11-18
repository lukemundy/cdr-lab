// Package httpserver runs the HTTP listeners of a service and serves its health probes.
package httpserver

import (
	"context"
	"io"
	"net/http"
)

// RegisterProbes adds the probes to mux; readiness fails when ready returns an error, which the response does not show.
func RegisterProbes(mux *http.ServeMux, ready func(context.Context) error) {
	mux.Handle("/livez", GetOnly(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, "ok")
	})))
	mux.Handle("/readyz", GetOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ready(r.Context()) != nil {
			write(w, http.StatusServiceUnavailable, "not ready")
			return
		}

		write(w, http.StatusOK, "ok")
	})))
}

// GetOnly is for a route that has no method in its pattern. With a method in the pattern, a request with a different
// method goes to a catch-all route of a different group on the same mux, if one exists, and not to a 405.
func GetOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func write(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}
