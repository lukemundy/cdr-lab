package authserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

var ErrInvalidArrangement = errors.New("the arrangement does not exist, is of a different client or ended")

type ArrangementRevoker interface {
	// RevokeArrangement ends the arrangement and revokes its tokens before it returns, or gives ErrInvalidArrangement.
	RevokeArrangement(ctx context.Context, clientID, arrangementID string) error
	// RefreshTokenRevoked tells the service that the client revoked the active refresh token of the arrangement at the
	// token revocation endpoint. The service decides if the arrangement ends.
	RefreshTokenRevoked(ctx context.Context, clientID, arrangementID string) error
}

const (
	arrangementIDParameter = "cdr_arrangement_id"
	// placeholderToken matches no token, so the library revokes nothing when it authenticates the client.
	placeholderToken = "cdr-arrangement-revocation"
)

// arrangementRevocation ignores cdr_arrangement_jwt, because a data holder supports only the form parameter.
func (d *DataHolder) arrangementRevocation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "the body is not a form")
		return
	}

	clientID, refusal := d.authenticate(withClientCert(r))
	if clientID == "" {
		writeTokenResponse(w, refusal)
		return
	}

	arrangementID := r.PostForm.Get(arrangementIDParameter)
	if arrangementID == "" {
		cds.WriteError(w, http.StatusBadRequest, cds.FieldMissing, arrangementIDParameter)
		return
	}

	err := d.cfg.Arrangements.RevokeArrangement(r.Context(), clientID, arrangementID)
	switch {
	case errors.Is(err, ErrInvalidArrangement):
		cds.WriteError(w, http.StatusUnprocessableEntity, cds.InvalidArrangement, arrangementID)
	case err != nil:
		cds.WriteUnexpected(w, r, err, "the arrangement could not be revoked")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// authenticate gives the ID of the client, or the refusal of the library. The library exports no client
// authentication, so the assertion goes to its revocation endpoint, with the request URI of this endpoint as a valid
// audience.
func (d *DataHolder) authenticate(r *http.Request) (string, *recorder) {
	revoking := &revokingClient{}

	form := url.Values{"token": {placeholderToken}}

	for _, name := range []string{"client_id", "client_assertion_type", "client_assertion"} {
		if value, ok := r.PostForm[name]; ok {
			form[name] = value
		}
	}

	auth := r.Clone(context.WithValue(r.Context(), revokingClientKey{}, revoking))
	auth.URL.Path = dataHolderRevocationPath
	auth.Form, auth.PostForm = form, form

	rec := &recorder{header: http.Header{}, status: http.StatusOK}
	d.provider.ServeHTTP(rec, auth)

	return revoking.id, rec
}
