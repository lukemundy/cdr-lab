// Package httpserver runs the HTTP listeners of a service and serves its health probes.
package httpserver

import (
	"context"
	"io"
	"net/http"
)

// RegisterProbes adds the probes to mux; readiness fails when ready returns an error, which the response does not show.
func RegisterProbes(mux *http.ServeMux, ready func(context.Context) error) {
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, "ok")
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready(r.Context()) != nil {
			write(w, http.StatusServiceUnavailable, "not ready")
			return
		}

		write(w, http.StatusOK, "ok")
	})
}

func write(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}
