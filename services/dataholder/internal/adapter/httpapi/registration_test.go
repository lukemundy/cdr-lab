package httpapi_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const registrationPath = "/register"

type registrar struct {
	registration domain.Registration
	err          error
	// requestJWT is the JWT of the last request.
	requestJWT string
	// updated and deleted are the client IDs of the last update and the last deletion.
	updated, deleted string
}

func (r *registrar) Register(_ context.Context, requestJWT string) (domain.Registration, error) {
	r.requestJWT = requestJWT

	return r.registration, r.err
}

func (r *registrar) Update(_ context.Context, clientID, requestJWT string) (domain.Registration, error) {
	r.requestJWT, r.updated = requestJWT, clientID

	return r.registration, r.err
}

func (r *registrar) Delete(_ context.Context, clientID string) error {
	r.deleted = clientID

	return r.err
}

func postRegistration(t *testing.T, r *registrar, contentType, body string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	httpapi.RegisterSecure(mux, secure(r, tokens{}, registrations{}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, registrationPath, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func TestARegistrationGives201WithTheRegistration(t *testing.T) {
	r := &registrar{registration: domain.Registration{
		ClientID:          "client-1",
		ClientIDIssuedAt:  1791547200,
		ClientName:        "Recipient Web",
		RedirectURIs:      []string{"https://recipient.example/cb"},
		GrantTypes:        []domain.GrantType{domain.GrantClientCredentials},
		ResponseTypes:     []string{"code"},
		SoftwareStatement: "header.ssa.signature",
		SoftwareID:        "software-1",
		Scope:             "openid cdr:registration",
	}}

	rec := postRegistration(t, r, "application/jwt", " header.request.signature\n")

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "header.request.signature", r.requestJWT)
	assert.JSONEq(t, `{
		"client_id": "client-1",
		"client_id_issued_at": 1791547200,
		"client_name": "Recipient Web",
		"client_description": "",
		"client_uri": "",
		"org_id": "",
		"org_name": "",
		"redirect_uris": ["https://recipient.example/cb"],
		"logo_uri": "",
		"jwks_uri": "",
		"revocation_uri": "",
		"recipient_base_uri": "",
		"token_endpoint_auth_method": "",
		"token_endpoint_auth_signing_alg": "",
		"grant_types": ["client_credentials"],
		"response_types": ["code"],
		"application_type": "",
		"id_token_signed_response_alg": "",
		"authorization_signed_response_alg": "",
		"request_object_signing_alg": "",
		"software_statement": "header.ssa.signature",
		"software_id": "software-1",
		"software_roles": "",
		"scope": "openid cdr:registration"
	}`, rec.Body.String(), "the response has no optional value that is absent, and no registration access token")
}

func TestAContentTypeWithAParameterIsAccepted(t *testing.T) {
	rec := postRegistration(t, &registrar{}, "application/jwt; charset=utf-8", "header.request.signature")

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestARegistrationWithADifferentContentTypeGives415WithNoBody(t *testing.T) {
	for _, contentType := range []string{"", "application/json", "text/plain"} {
		t.Run(contentType, func(t *testing.T) {
			r := &registrar{}

			rec := postRegistration(t, r, contentType, "header.request.signature")

			assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
			assert.Empty(t, rec.Body.String())
			assert.Empty(t, r.requestJWT, "the request is not read")
		})
	}
}

func TestARegistrationThatIsRefusedGives400WithTheErrorCode(t *testing.T) {
	r := &registrar{err: domain.RegistrationError{
		Code:        domain.InvalidRedirectURI,
		Description: "a value of redirect_uris is not in the redirect_uris of the SSA",
	}}

	rec := postRegistration(t, r, "application/jwt", "header.request.signature")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		"error": "invalid_redirect_uri",
		"error_description": "a value of redirect_uris is not in the redirect_uris of the SSA"
	}`, rec.Body.String())
}

func TestARegistrationThatFailsGives500WithNoCause(t *testing.T) {
	r := &registrar{err: errors.New("get the Register JWKS: connection refused")}

	rec := postRegistration(t, r, "application/jwt", "header.request.signature")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error": "server_error"}`, rec.Body.String())
}

func TestARegistrationRequestThatIsTooLargeIsRefused(t *testing.T) {
	r := &registrar{}

	rec := postRegistration(t, r, "application/jwt", strings.Repeat("a", 64<<10+1))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid_client_metadata")
	assert.Empty(t, r.requestJWT)
}

func TestOtherMethodsAndPathsOfTheRegistrationEndpointAreNotFound(t *testing.T) {
	mux := http.NewServeMux()
	httpapi.RegisterSecure(mux, secure(&registrar{}, tokens{}, registrations{}))

	for _, target := range []struct{ method, path string }{
		{http.MethodGet, registrationPath},
		{http.MethodPost, registrationPath + "/client-1"},
		{http.MethodGet, registrationPath + "/client-1/other"},
		{http.MethodPost, "/token"},
	} {
		t.Run(target.method+" "+target.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), target.method, target.path, nil))

			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Contains(t, rec.Body.String(), "urn:au-cds:error:cds-all:Resource/NotFound")
		})
	}
}

// secure has the registration dependencies of a test and none of the data of a consent.
func secure(r httpapi.Registrar, v tokens, regs registrations) httpapi.Secure {
	return httpapi.Secure{
		ResourceBaseURL:  resourceBaseURL,
		RegistrationPath: registrationPath,
		UserInfoPath:     userInfoPath,
		Registrar:        r,
		Tokens:           v,
		Registrations:    regs,
		ConsentTokens:    consentTokens{},
		Customers:        arrangementCustomers{},
	}
}

// tokens gives the access token of each bearer value, and an invalid token error for a value that it does not have.
type tokens map[string]authserver.AccessToken

func (v tokens) VerifyAccessToken(
	_ context.Context,
	token string,
	_ *x509.Certificate,
) (authserver.AccessToken, error) {
	if token == "failure" {
		return authserver.AccessToken{}, errors.New("signing keys not available")
	}

	verified, ok := v[token]
	if !ok {
		return authserver.AccessToken{}, authserver.ErrInvalidToken
	}

	return verified, nil
}

type registrations map[string]domain.Registration

func (r registrations) Registration(_ context.Context, clientID string) (domain.Registration, error) {
	if clientID == "failure" {
		return domain.Registration{}, errors.New("database failure")
	}

	registration, ok := r[clientID]
	if !ok {
		return domain.Registration{}, app.ErrRegistrationNotFound
	}

	return registration, nil
}

func getRegistration(t *testing.T, clientID, authorization string) *httptest.ResponseRecorder {
	t.Helper()

	return manageRegistration(t, &registrar{}, http.MethodGet, clientID, authorization, "", "")
}

// manageRegistration sends a registration management request with the tokens and the registrations of the tests.
func manageRegistration(
	t *testing.T,
	r *registrar,
	method, clientID, authorization, contentType, body string,
) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	httpapi.RegisterSecure(mux, secure(
		r,
		tokens{
			"token-1":  {ClientID: "client-1", Scopes: []string{"cdr:registration"}},
			"token-2":  {ClientID: "client-2", Scopes: []string{"cdr:registration"}},
			"no-scope": {ClientID: "client-1", Scopes: []string{"bank:accounts.basic:read"}},
			"token-f":  {ClientID: "failure", Scopes: []string{"cdr:registration"}},
		},
		registrations{
			"client-1": {ClientID: "client-1", SoftwareID: "software-1", Scope: "openid cdr:registration"},
			"client-2": {ClientID: "client-2", SoftwareID: "software-2"},
		},
	))

	req := httptest.NewRequestWithContext(t.Context(), method, registrationPath+"/"+clientID, strings.NewReader(body))
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func TestAClientReadsItsRegistrationWithItsToken(t *testing.T) {
	rec := getRegistration(t, "client-1", "Bearer token-1")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "client-1", body["client_id"])
	assert.Equal(t, "software-1", body["software_id"])
	assert.Equal(t, "openid cdr:registration", body["scope"])
}

func TestTheBearerSchemeIsNotCaseSensitive(t *testing.T) {
	assert.Equal(t, http.StatusOK, getRegistration(t, "client-1", "bearer token-1").Code)
}

func TestAReadOfARegistrationThatIsRefusedHasNoBody(t *testing.T) {
	tests := map[string]struct {
		clientID, authorization string
		status                  int
		challenge               string
	}{
		"no token":           {"client-1", "", http.StatusUnauthorized, "Bearer"},
		"a different scheme": {"client-1", "Basic dXNlcjpwYXNz", http.StatusUnauthorized, "Bearer"},
		"a token that is not valid": {
			"client-1",
			"Bearer unknown",
			http.StatusUnauthorized,
			`Bearer error="invalid_token"`,
		},
		"a client ID that does not exist": {
			"client-9",
			"Bearer token-1",
			http.StatusUnauthorized,
			`Bearer error="invalid_token"`,
		},
		"a token with no registration scope": {
			"client-1", "Bearer no-scope", http.StatusForbidden,
			`Bearer error="insufficient_scope", scope="cdr:registration"`,
		},
		"the client ID of a different registration": {"client-2", "Bearer token-1", http.StatusForbidden, ""},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := getRegistration(t, tc.clientID, tc.authorization)

			assert.Equal(t, tc.status, rec.Code)
			assert.Equal(t, tc.challenge, rec.Header().Get("WWW-Authenticate"))
			assert.Empty(t, rec.Body.String())
		})
	}
}

func TestAReadOfARegistrationThatFailsGives500WithNoCause(t *testing.T) {
	tests := map[string]struct{ clientID, authorization string }{
		"the token verification fails": {"client-1", "Bearer failure"},
		"the read model fails":         {"failure", "Bearer token-f"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := getRegistration(t, tc.clientID, tc.authorization)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.JSONEq(t, `{"error": "server_error"}`, rec.Body.String())
		})
	}
}

func TestAClientUpdatesItsRegistrationWithItsToken(t *testing.T) {
	r := &registrar{registration: domain.Registration{ClientID: "client-1", SoftwareID: "software-1"}}

	rec := manageRegistration(
		t, r, http.MethodPut, "client-1", "Bearer token-1", "application/jwt", "header.request.signature",
	)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "client-1", r.updated)
	assert.Equal(t, "header.request.signature", r.requestJWT)
	assert.Contains(t, rec.Body.String(), `"client_id":"client-1"`)
}

func TestAnUpdateThatIsRefusedGivesTheResponseOfARegistrationThatIsRefused(t *testing.T) {
	t.Run("a request that is not valid", func(t *testing.T) {
		r := &registrar{err: domain.RegistrationError{
			Code: domain.InvalidSoftwareStatement, Description: "the SSA is expired",
		}}

		rec := manageRegistration(
			t, r, http.MethodPut, "client-1", "Bearer token-1", "application/jwt", "header.request.signature",
		)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(
			t, `{"error": "invalid_software_statement", "error_description": "the SSA is expired"}`, rec.Body.String(),
		)
	})

	t.Run("a different content type", func(t *testing.T) {
		r := &registrar{}

		rec := manageRegistration(t, r, http.MethodPut, "client-1", "Bearer token-1", "application/json", "{}")

		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
		assert.Empty(t, r.updated)
	})
}

func TestAClientDeletesItsRegistrationWithItsToken(t *testing.T) {
	r := &registrar{}

	rec := manageRegistration(t, r, http.MethodDelete, "client-1", "Bearer token-1", "", "")

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
	assert.Equal(t, "client-1", r.deleted)
}

func TestAnUpdateOrADeletionWithNoTokenOfTheClientChangesNothing(t *testing.T) {
	tests := map[string]struct {
		clientID, authorization string
		status                  int
	}{
		"no token":                                  {"client-1", "", http.StatusUnauthorized},
		"a token that is not valid":                 {"client-1", "Bearer unknown", http.StatusUnauthorized},
		"a client ID that does not exist":           {"client-9", "Bearer token-1", http.StatusUnauthorized},
		"the client ID of a different registration": {"client-2", "Bearer token-1", http.StatusForbidden},
	}

	for name, tc := range tests {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			t.Run(method+" with "+name, func(t *testing.T) {
				r := &registrar{}

				rec := manageRegistration(
					t, r, method, tc.clientID, tc.authorization, "application/jwt", "header.request.signature",
				)

				assert.Equal(t, tc.status, rec.Code)
				assert.Empty(t, rec.Body.String())
				assert.Empty(t, r.updated)
				assert.Empty(t, r.deleted)
			})
		}
	}
}

func TestAnUpdateOrADeletionOfARegistrationThatIsDeletedAfterTheTokenCheckGives401(t *testing.T) {
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			r := &registrar{err: app.ErrRegistrationNotFound}

			rec := manageRegistration(
				t, r, method, "client-1", "Bearer token-1", "application/jwt", "header.request.signature",
			)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, `Bearer error="invalid_token"`, rec.Header().Get("WWW-Authenticate"))
		})
	}
}

func TestADeletionThatFailsGives500WithNoCause(t *testing.T) {
	r := &registrar{err: errors.New("database failure")}

	rec := manageRegistration(t, r, http.MethodDelete, "client-1", "Bearer token-1", "", "")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error": "server_error"}`, rec.Body.String())
}
