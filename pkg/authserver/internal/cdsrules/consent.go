package cdsrules

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

const (
	authTimeKey      = "auth_time"
	sharingEndsAtKey = "sharing_ends_at"
)

type Consent struct {
	ArrangementID string
	// Subject is the final sub, because the function of go-oidc for a pairwise sub cannot fail.
	Subject         string
	AuthenticatedAt time.Time
	SharingEndsAt   time.Time
}

// Approve records the consent in the session, whose store go-oidc copies to the grant.
func Approve(as *goidc.AuthnSession, consent Consent) {
	if as.Store == nil {
		as.Store = map[string]any{}
	}

	as.Subject = consent.Subject
	as.GrantedScopes = as.Scopes
	maps.Copy(as.Store, StoreOf(consent, SharingDuration(as.Store)))
}

func StoreOf(consent Consent, sharingDuration time.Duration) map[string]any {
	return map[string]any{
		SharingDurationClaim: int(sharingDuration / time.Second),
		arrangementIDClaim:   consent.ArrangementID,
		authTimeKey:          int(consent.AuthenticatedAt.Unix()),
		sharingEndsAtKey:     int(consent.SharingEndsAt.Unix()),
	}
}

// ConsentOf also gives the sharing duration of the grant.
func ConsentOf(grant *goidc.Grant) (Consent, time.Duration) {
	return Consent{
		ArrangementID:   ArrangementID(grant),
		Subject:         grant.Subject,
		AuthenticatedAt: time.Unix(int64(storedSeconds(grant.Store[authTimeKey])), 0).UTC(),
		SharingEndsAt:   time.Unix(int64(storedSeconds(grant.Store[sharingEndsAtKey])), 0).UTC(),
	}, SharingDuration(grant.Store)
}

// ArrangementID gives the cdr_arrangement_id of a grant, and nothing for a grant with no consent.
func ArrangementID(grant *goidc.Grant) string {
	id, _ := grant.Store[arrangementIDClaim].(string)

	return id
}

// RefreshTokenForSharingDuration sets the refresh token expiry for each grant, because go-oidc has one lifetime for
// all grants; do not also use provider.WithRefreshTokenLifetime or provider.WithRefreshTokenRotation, because they
// replace it.
func RefreshTokenForSharingDuration(_ context.Context, _ *goidc.Client, grant *goidc.Grant) bool {
	if storedSeconds(grant.Store[SharingDurationClaim]) <= 0 {
		return false
	}

	grant.RefreshTokenExpiresAt = storedSeconds(grant.Store[sharingEndsAtKey])

	return true
}

// AccessTokenLifetime gives no seconds after the end of the sharing of a consent (CDR Rules, rule 4.26(1)(g)).
func AccessTokenLifetime(grant *goidc.Grant, lifetimeSeconds int, now time.Time) int {
	if storedSeconds(grant.Store[SharingDurationClaim]) <= 0 {
		return lifetimeSeconds
	}

	remaining := int(time.Unix(int64(storedSeconds(grant.Store[sharingEndsAtKey])), 0).Sub(now) / time.Second)

	return max(0, min(lifetimeSeconds, remaining))
}

// IDTokenClaims gives the claims of the authentication, which go-oidc does not keep.
func IDTokenClaims(_ context.Context, grant *goidc.Grant) map[string]any {
	return map[string]any{
		goidc.ClaimACR:      SingleFactorACR,
		goidc.ClaimAuthTime: storedSeconds(grant.Store[authTimeKey]),
	}
}

type refreshGrants struct {
	goidc.RefreshTokenManager

	clientCert goidc.ClientCertFunc
}

// RefreshWithAnyCertificate gives each grant the certificate of the request, because go-oidc refuses a refresh from a
// certificate that is not that of the last token request, which ends a consent when a client replaces its certificate.
func RefreshWithAnyCertificate(
	grants goidc.RefreshTokenManager,
	clientCert goidc.ClientCertFunc,
) goidc.RefreshTokenManager {
	return refreshGrants{RefreshTokenManager: grants, clientCert: clientCert}
}

func (r refreshGrants) GrantByRefreshToken(ctx context.Context, token string) (*goidc.Grant, error) {
	grant, err := r.RefreshTokenManager.GrantByRefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	cert, err := r.clientCert(ctx)
	if err != nil || grant.CertThumbprint == "" {
		return grant, nil //nolint:nilerr // go-oidc refuses a request with no certificate.
	}

	bound := *grant
	bound.CertThumbprint = thumbprint(cert)

	return &bound, nil
}

// thumbprint is the x5t#S256 value of RFC 8705, section 3.1.
func thumbprint(cert *x509.Certificate) string {
	sum := sha256.Sum256(cert.Raw)

	return base64.RawURLEncoding.EncodeToString(sum[:])
}

type issuedKey struct{}

// issued carries the arrangement of a token from RecordArrangement to TokenResponse.
type issued struct {
	arrangementID string
}

// RecordArrangement is a token handler; it needs TokenResponse.
func RecordArrangement(ctx context.Context, _ *goidc.Token, grant *goidc.Grant) error {
	if i, ok := ctx.Value(issuedKey{}).(*issued); ok {
		i.arrangementID = ArrangementID(grant)
	}

	return nil
}

type oauthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// TokenResponse adds cdr_arrangement_id, for which go-oidc has no hook, and replaces the expired_token that go-oidc
// gives for the code of a revoked grant with invalid_grant; it needs RecordArrangement.
func TokenResponse(tokenPath string) goidc.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != tokenPath {
				next.ServeHTTP(w, r)
				return
			}

			record := &issued{}
			rec := &recorder{header: http.Header{}, status: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), issuedKey{}, record)))

			body := rec.body.Bytes()

			switch {
			case rec.status == http.StatusOK && record.arrangementID != "":
				body = withMember(body, arrangementIDClaim, record.arrangementID)
			case rec.status != http.StatusOK:
				var e oauthError
				if json.Unmarshal(body, &e) == nil && e.Code == string(goidc.ErrorCodeExpiredToken) {
					e.Code = string(goidc.ErrorCodeInvalidGrant)
					body = mustMarshal(e)
				}
			}

			rec.writeTo(w, body)
		})
	}
}

// withMember gives body with no change when body is not a JSON object.
func withMember(body []byte, name, value string) []byte {
	var object map[string]json.RawMessage
	if json.Unmarshal(body, &object) != nil {
		return body
	}

	object[name] = mustMarshal(value)

	return mustMarshal(object)
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}

type introspectionKey struct{}

// introspection carries the authenticated client from IntrospectingClient to IntrospectionResponse.
type introspection struct {
	clientID string
}

// IntrospectingClient refuses no client, because go-oidc answers a refusal with an error where RFC 7662 requires a
// response that is not active; go-oidc calls it only for an active token, and IntrospectionResponse decides.
func IntrospectingClient(ctx context.Context, client *goidc.Client, _ goidc.TokenInfo) bool {
	if i, ok := ctx.Value(introspectionKey{}).(*introspection); ok {
		i.clientID = client.ID
	}

	return true
}

type refreshTokenInfo struct {
	Active        bool   `json:"active"`
	ExpiresAt     int    `json:"exp,omitempty"`
	Scopes        string `json:"scope,omitempty"`
	ArrangementID string `json:"cdr_arrangement_id,omitempty"`
}

// IntrospectionResponse replaces the response of go-oidc, which is also active for access tokens and for the tokens
// of other clients, and which has no hook for cdr_arrangement_id; it needs IntrospectingClient.
func IntrospectionResponse(introspectionPath string, grants goidc.RefreshTokenManager) goidc.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != introspectionPath || r.ParseForm() != nil {
				next.ServeHTTP(w, r)
				return
			}

			caller := &introspection{}
			rec := &recorder{header: http.Header{}, status: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), introspectionKey{}, caller)))

			if rec.status != http.StatusOK {
				rec.writeTo(w, rec.body.Bytes())
				return
			}

			info, err := refreshTokenOf(r.Context(), grants, caller.clientID, r.PostForm.Get("token"))
			if err != nil {
				rec.status = http.StatusInternalServerError
				rec.writeTo(w, mustMarshal(oauthError{Code: string(goidc.ErrorCodeInternalError)}))

				return
			}

			rec.writeTo(w, mustMarshal(info))
		})
	}
}

// refreshTokenOf gives a response that is not active when go-oidc found no active token (clientID is then empty), and
// when the token is not a refresh token of the client.
func refreshTokenOf(
	ctx context.Context,
	grants goidc.RefreshTokenManager,
	clientID, token string,
) (refreshTokenInfo, error) {
	if clientID == "" {
		return refreshTokenInfo{}, nil
	}

	grant, err := grants.GrantByRefreshToken(ctx, token)
	if errors.Is(err, goidc.ErrNotFound) {
		return refreshTokenInfo{}, nil
	}

	if err != nil {
		return refreshTokenInfo{}, err
	}

	if grant.ClientID != clientID {
		return refreshTokenInfo{}, nil
	}

	return refreshTokenInfo{
		Active:        true,
		ExpiresAt:     grant.RefreshTokenExpiresAt,
		Scopes:        grant.Scopes,
		ArrangementID: ArrangementID(grant),
	}, nil
}

// recorder keeps the response of go-oidc, so that a middleware can change it.
type recorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *recorder) Header() http.Header         { return r.header }
func (r *recorder) WriteHeader(status int)      { r.status = status }
func (r *recorder) Write(b []byte) (int, error) { return r.body.Write(b) }

// writeTo sets the headers of RFC 6749, section 5.1.
func (r *recorder) writeTo(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(r.status)
	_, _ = w.Write(body) //nolint:gosec // G705: the body is JSON with a JSON content type, so no browser renders it.
}
