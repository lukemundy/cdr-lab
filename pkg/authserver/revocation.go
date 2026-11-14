package authserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

// RevokedTokenStore keeps the ID of each revoked access token. It can forget an ID after expiresAt, when the token
// is no longer valid.
type RevokedTokenStore interface {
	RevokeToken(ctx context.Context, id string, expiresAt time.Time) error
	TokenRevoked(ctx context.Context, id string) (bool, error)
}

var unauthorizedClient = []byte(`{"error":"unauthorized_client"}`)

type revokingClientKey struct{}

// revokingClient carries the authenticated client from the library to revocation.
func revokingClient(ctx context.Context, client *goidc.Client) bool {
	if id, ok := ctx.Value(revokingClientKey{}).(*string); ok {
		*id = client.ID
	}

	return true
}

// revocation revokes one access token, because the library can revoke an access token only with its grant. It also
// gives unauthorized_client for a token of a different client, where the library gives access_denied.
func (d *DataHolder) revocation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The library reads the body of a copy of r, so r keeps no form unless it is parsed first.
		if err := r.ParseForm(); err != nil {
			next.ServeHTTP(w, r)
			return
		}

		var clientID string

		rec := &recorder{header: http.Header{}, status: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), revokingClientKey{}, &clientID)))

		var refusal tokenError
		if rec.status == http.StatusForbidden && json.Unmarshal(rec.body.Bytes(), &refusal) == nil &&
			refusal.Code == string(goidc.ErrorCodeAccessDenied) {
			respond(w, http.StatusBadRequest, unauthorizedClient)
			return
		}

		if rec.status != http.StatusOK {
			respond(w, rec.status, rec.body.Bytes())
			return
		}

		err := d.revokeAccessToken(r.Context(), r.PostForm.Get("token"), clientID)
		switch {
		case errors.Is(err, errOtherClient):
			respond(w, http.StatusBadRequest, unauthorizedClient)
		case err != nil:
			slog.ErrorContext(r.Context(), "the access token could not be revoked", slog.Any("error", err))
			respond(w, http.StatusInternalServerError, nil)
		default:
			respond(w, rec.status, rec.body.Bytes())
		}
	})
}

var errOtherClient = errors.New("the access token is of a different client")

// revokeAccessToken changes nothing for a value that is not an access token of the data holder.
func (d *DataHolder) revokeAccessToken(ctx context.Context, token, clientID string) error {
	claims, err := parseAccessToken(ctx, token, d.cfg.Issuer, d.cfg.SigningKeys, d.now())
	if errors.Is(err, ErrInvalidToken) {
		return nil
	}

	if err != nil {
		return err
	}

	if claims.ClientID != clientID {
		return errOtherClient
	}

	return d.cfg.RevokedTokens.RevokeToken(ctx, claims.ID, claims.ExpiresAt)
}

// respond writes to tokenHandler, which gives each response the headers and the error codes of RFC 6749.
func respond(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
