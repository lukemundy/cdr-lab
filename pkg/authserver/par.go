package authserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

// The codes of a request that the client can correct, for which the response gives the cause.
var correctableCodes = []goidc.ErrorCode{
	goidc.ErrorCodeInvalidRequest,
	goidc.ErrorCodeInvalidRequestObject,
	goidc.ErrorCodeInvalidScope,
	goidc.ErrorCodeInvalidGrant,
}

var errJTINotChecked = errors.New("the JWT ID could not be checked")

type refusalCauseKey struct{}

// refusalCause carries the cause of a refusal from the error handler of the library to the handler of the endpoint,
// because the library does not put the cause in its response.
type refusalCause struct {
	description string
}

// recordRefusal is the error handler of the library.
func recordRefusal(ctx context.Context, err error) {
	logRefusal(ctx, err)

	cause, ok := ctx.Value(refusalCauseKey{}).(*refusalCause)
	if !ok {
		return
	}

	refusal, ok := errors.AsType[goidc.Error](err)
	if !ok || !slices.Contains(correctableCodes, refusal.Code) {
		return
	}

	if wrapped := errors.Unwrap(refusal); wrapped != nil {
		cause.description = wrapped.Error()
	}
}

// consumeJTI hides the detail of a store failure, which the library can put in the cause of a refusal.
func consumeJTI(store JTIStore) goidc.ConsumeJTIFunc {
	return func(ctx context.Context, jti string) error {
		err := store.Consume(ctx, jti)
		if err != nil && !errors.Is(err, ErrJTIUsed) {
			return errJTINotChecked
		}

		return err
	}
}

type pushedRequestError struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// pushedRequestHandler needs a listener that verifies the client certificate.
func pushedRequestHandler(provider http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cause := &refusalCause{}
		ctx := context.WithValue(req.Context(), refusalCauseKey{}, cause)

		if req.TLS != nil && len(req.TLS.PeerCertificates) > 0 {
			ctx = context.WithValue(ctx, ctxKeyClientCert{}, req.TLS.PeerCertificates[0])
		}

		rec := &recorder{header: http.Header{}, status: http.StatusOK}
		provider.ServeHTTP(rec, req.WithContext(ctx))

		// RFC 9126, section 2.2 requires that the response is not cached.
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")

		if rec.status == http.StatusCreated {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(rec.body.Bytes())

			return
		}

		writePushedRequestError(w, rec, cause.description)
	}
}

func writePushedRequestError(w http.ResponseWriter, rec *recorder, cause string) {
	var e pushedRequestError

	_ = json.Unmarshal(rec.body.Bytes(), &e)

	status := rec.status

	switch {
	case status >= http.StatusInternalServerError || e.Code == "":
		status, e = http.StatusInternalServerError, pushedRequestError{Code: "server_error"}
	case cause != "":
		e.Description = cause
	}

	data, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(status)
	_, _ = w.Write(data)
}
