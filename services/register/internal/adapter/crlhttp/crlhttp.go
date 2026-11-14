// Package crlhttp serves the CRL of the mock CDR intermediate to each relying party, with no authentication.
package crlhttp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

type Source interface {
	CurrentCRL(ctx context.Context) (app.CRL, error)
}

// Register uses a GET pattern, which also matches HEAD, so that the mux gives 405 with an Allow header for another
// method.
func Register(mux *http.ServeMux, crls Source) {
	mux.Handle("GET /crl/cdr-intermediate.crl", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		crl, err := crls.CurrentCRL(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "the CRL is not available", "error.type", fmt.Sprintf("%T", err))
			http.Error(w, "The CRL is not available.", http.StatusServiceUnavailable)

			return
		}

		w.Header().Set("Content-Type", "application/pkix-crl")
		w.Header().Set("Content-Length", strconv.Itoa(len(crl.DER)))

		if r.Method == http.MethodHead {
			return
		}

		_, _ = w.Write(crl.DER)
	}))
}
