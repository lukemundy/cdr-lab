package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"slices"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
)

type DisclosureRecorder interface {
	RecordDisclosure(ctx context.Context, arrangementID, scope string) error
}

// userInfoScope gives no scope for a token with no profile, because the response then discloses only the sub, which
// is in no data cluster.
func userInfoScope(token authserver.ConsentToken) string {
	if slices.Contains(token.Scopes, scopeProfile) {
		return scopeProfile
	}

	return ""
}

// recordDisclosure keeps the response until the record of a 2xx response is written, so that no data leaves with no
// record; scopeOf gives "" for a response that discloses no data cluster, and a HEAD response has no body to disclose.
func recordDisclosure(
	recorder DisclosureRecorder,
	scopeOf func(authserver.ConsentToken) string,
) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := consentTokenOf(r)

			scope := scopeOf(token)
			if scope == "" || r.Method == http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// The held response writes to the header of w, which goes to the client only at the WriteHeader of w.
			response := &heldResponse{header: w.Header()}
			next.ServeHTTP(response, r)

			if !response.successful() {
				response.sendTo(w)
				return
			}

			if err := recorder.RecordDisclosure(r.Context(), token.ArrangementID, scope); err != nil {
				cds.WriteUnexpected(w, r, err, "The data holder cannot record the disclosure.")
				return
			}

			response.sendTo(w)
		})
	}
}

// heldResponse keeps a response in memory, which is small because a response of a consent is at most one CDS page.
type heldResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (h *heldResponse) Header() http.Header { return h.header }

func (h *heldResponse) WriteHeader(status int) {
	if h.status == 0 {
		h.status = status
	}
}

func (h *heldResponse) Write(b []byte) (int, error) {
	h.WriteHeader(http.StatusOK)

	return h.body.Write(b)
}

func (h *heldResponse) sendTo(w http.ResponseWriter) {
	w.WriteHeader(h.statusCode())
	_, _ = w.Write(h.body.Bytes())
}

func (h *heldResponse) statusCode() int {
	return max(h.status, http.StatusOK)
}

func (h *heldResponse) successful() bool {
	return h.statusCode() >= http.StatusOK && h.statusCode() < http.StatusMultipleChoices
}
