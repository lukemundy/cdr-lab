package cds

import (
	"net/http"
	"strings"
)

const fallbackPattern = "/"

// The probe has no OPTIONS, CONNECT or TRACE, because no CDS API supports them.
var probedMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

// RegisterFallback serves each request that no other pattern of mux matches. A pattern with no method matches all
// methods, so the 405 of http.ServeMux does not occur when mux also has the pattern "/".
func RegisterFallback(mux *http.ServeMux, notFoundDetail string) {
	mux.HandleFunc(fallbackPattern, func(w http.ResponseWriter, r *http.Request) {
		if allowed := allowedMethods(mux, r); len(allowed) > 0 {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			WriteError(
				w,
				http.StatusMethodNotAllowed,
				Expected,
				"The endpoint does not support the method "+r.Method+".",
			)

			return
		}

		WriteError(w, http.StatusNotFound, ResourceNotFound, notFoundDetail)
	})
}

func allowedMethods(mux *http.ServeMux, r *http.Request) []string {
	var allowed []string

	for _, method := range probedMethods {
		probe := r.Clone(r.Context())
		probe.Method = method

		if _, pattern := mux.Handler(probe); pattern != fallbackPattern {
			allowed = append(allowed, method)
		}
	}

	return allowed
}
