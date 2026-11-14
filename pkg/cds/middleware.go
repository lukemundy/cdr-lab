// Package cds has the HTTP rules of the Consumer Data Standards that all participants share.
package cds

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// AllowAnyOrigin is for unauthenticated APIs only.
func AllowAnyOrigin(next http.Handler, exposedHeaders ...string) http.Handler {
	exposed := strings.Join(exposedHeaders, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			h.Set("Access-Control-Allow-Methods", http.MethodGet)

			if requested := r.Header.Values("Access-Control-Request-Headers"); len(requested) > 0 {
				h.Set("Access-Control-Allow-Headers", strings.Join(requested, ","))
			}

			w.WriteHeader(http.StatusNoContent)

			return
		}

		h.Set("Access-Control-Expose-Headers", exposed)
		next.ServeHTTP(w, r)
	})
}

func InteractionID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("x-fapi-interaction-id")
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set("x-fapi-interaction-id", id)
		next.ServeHTTP(w, r)
	})
}
