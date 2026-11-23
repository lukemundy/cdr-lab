package operatorapi_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const token = "test-operator-token"

// httptest gives example.com as the host of a request.
var access = operatorhttp.Config{Token: token, AllowedHosts: []string{"example.com"}}

type fakeOperator struct {
	legalEntityID     string
	command           domain.DataRecipientCommand
	dataHolderCommand domain.DataHolderCommand
	participants      domain.Participants
	load              domain.ParticipantLoad
	err               error
}

func (f *fakeOperator) LoadParticipants(_ context.Context, p domain.Participants) (domain.ParticipantLoad, error) {
	f.participants = p
	return f.load, f.err
}

func (f *fakeOperator) Execute(_ context.Context, legalEntityID string, command domain.DataRecipientCommand) error {
	f.legalEntityID, f.command = legalEntityID, command
	return f.err
}

func (f *fakeOperator) ExecuteDataHolder(
	_ context.Context,
	legalEntityID string,
	command domain.DataHolderCommand,
) error {
	f.legalEntityID, f.dataHolderCommand = legalEntityID, command
	return f.err
}

func serve(t *testing.T, operator *fakeOperator, method, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, operator, &fakeIssuer{}, &fakeRevoker{}, nil, &fakeKeys{}, access)

	req := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func authorised() http.Header {
	return http.Header{"Authorization": {"Bearer " + token}}
}

type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func requireProblem(t *testing.T, rec *httptest.ResponseRecorder, status int) problem {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var p problem
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p))
	assert.Equal(t, "about:blank", p.Type)
	assert.Equal(t, http.StatusText(status), p.Title)
	assert.Equal(t, status, p.Status)
	assert.NotEmpty(t, p.Detail)

	return p
}

// Verifies: REG-0801, REG-0802, REG-0863, REG-0864
func TestEachActionExecutesItsCommand(t *testing.T) {
	const (
		recipient = "/v1/data-recipients/dr-1"
		product   = recipient + "/software-products/sp-1"
	)

	tests := map[string]domain.DataRecipientCommand{
		recipient + "/suspend":   domain.SuspendAccreditation{},
		recipient + "/revoke":    domain.RevokeAccreditation{},
		recipient + "/surrender": domain.SurrenderAccreditation{},
		recipient + "/reinstate": domain.ReinstateAccreditation{},
		product + "/activate":    domain.ActivateSoftwareProduct{SoftwareProductID: "sp-1"},
		product + "/deactivate":  domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-1"},
		product + "/remove":      domain.RemoveSoftwareProduct{SoftwareProductID: "sp-1"},
	}
	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			operator := &fakeOperator{}

			rec := serve(t, operator, http.MethodPost, path, authorised())

			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			assert.Empty(t, rec.Body.String())
			assert.Equal(t, "dr-1", operator.legalEntityID)
			assert.Equal(t, want, operator.command)
		})
	}
}

func TestARequestWithoutTheOperatorTokenIsUnauthorised(t *testing.T) {
	const noErrorChallenge, invalidChallenge = `Bearer`, `Bearer error="invalid_token"`

	tests := map[string]struct {
		header    http.Header
		challenge string
	}{
		"no authorization":   {http.Header{}, noErrorChallenge},
		"not a bearer token": {http.Header{"Authorization": {"Basic " + token}}, noErrorChallenge},
		"no scheme":          {http.Header{"Authorization": {token}}, noErrorChallenge},
		"wrong token":        {http.Header{"Authorization": {"Bearer not-the-token"}}, invalidChallenge},
		"token as a prefix":  {http.Header{"Authorization": {"Bearer " + token + "x"}}, invalidChallenge},
		"empty token":        {http.Header{"Authorization": {"Bearer "}}, invalidChallenge},
	}
	for name, tt := range tests {
		for _, path := range []string{"/v1/data-recipients/dr-1/suspend", "/v1/unknown"} {
			t.Run(name+" "+path, func(t *testing.T) {
				operator := &fakeOperator{}

				rec := serve(t, operator, http.MethodPost, path, tt.header)

				requireProblem(t, rec, http.StatusUnauthorized)
				assert.Equal(t, tt.challenge, rec.Header().Get("WWW-Authenticate"))
				assert.Nil(t, operator.command)
			})
		}
	}
}

func TestTheProblemDetailSaysWhetherTheTokenIsMissingOrNotValid(t *testing.T) {
	missing := requireProblem(
		t,
		serve(t, &fakeOperator{}, http.MethodPost, "/v1/unknown", http.Header{}),
		http.StatusUnauthorized,
	)
	invalid := requireProblem(t, serve(t, &fakeOperator{}, http.MethodPost, "/v1/unknown",
		http.Header{"Authorization": {"Bearer not-the-token"}}), http.StatusUnauthorized)

	assert.Equal(t, "The request must have the operator token as a bearer token.", missing.Detail)
	assert.Equal(t, "The bearer token is not the operator token.", invalid.Detail)
}

func TestTheBearerSchemeIsNotCaseSensitive(t *testing.T) {
	rec := serve(
		t,
		&fakeOperator{},
		http.MethodPost,
		"/v1/data-recipients/dr-1/suspend",
		http.Header{"Authorization": {"bearer " + token}},
	)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// Verifies: REG-0860, REG-0861, REG-0862
func TestErrors(t *testing.T) {
	ruleErr := statusRuleError(t)

	tests := map[string]struct {
		err        error
		status     int
		wantDetail string
	}{
		"status rule": {ruleErr, http.StatusConflict, ruleErr.Error()},
		"not found": {
			fmt.Errorf("data recipient %w", domain.ErrNotFound),
			http.StatusNotFound,
			"data recipient does not exist",
		},
		"version conflict": {
			eventsourcing.ErrVersionConflict,
			http.StatusConflict,
			"The data recipient changed during the request. Send the request again.",
		},
		"other": {
			errors.New("connection to 10.0.0.1 failed"),
			http.StatusInternalServerError,
			"The Register could not complete the request.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(
				t,
				&fakeOperator{err: tt.err},
				http.MethodPost,
				"/v1/data-recipients/dr-1/suspend",
				authorised(),
			)

			p := requireProblem(t, rec, tt.status)
			assert.Equal(t, tt.wantDetail, p.Detail)
		})
	}
}

// Verifies: REG-1508
func TestTheOperatorInterfaceServesOnlyThePathsUnderV1(t *testing.T) {
	requireProblem(t, serve(t, &fakeOperator{}, http.MethodGet, "/v1/unknown", http.Header{}), http.StatusUnauthorized)

	for _, path := range []string{"/", "/unknown", "/v2/unknown"} {
		rec := serve(t, &fakeOperator{}, http.MethodGet, path, http.Header{})

		assert.Equal(t, http.StatusNotFound, rec.Code, path)
		assert.Empty(t, rec.Header().Get("WWW-Authenticate"), path)
	}
}

func TestAnUnknownPathOrMethodIsNotFound(t *testing.T) {
	for _, tt := range []struct{ method, path string }{
		{http.MethodPost, "/v1/unknown"},
		{http.MethodPost, "/v1/data-recipients/dr-1/pause"},
		{http.MethodGet, "/v1/data-recipients/dr-1/suspend"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := serve(t, &fakeOperator{}, tt.method, tt.path, authorised())

			requireProblem(t, rec, http.StatusNotFound)
		})
	}
}

func statusRuleError(t *testing.T) error {
	t.Helper()

	dr := domain.DataRecipient{
		LegalEntityID:       "dr-1",
		LegalEntityName:     "Recipient",
		AccreditationNumber: "ADR1",
		AccreditationLevel:  domain.Sponsored,
		LogoURI:             "https://recipient.example/logo.png",
		Status:              domain.DataRecipientSurrendered,
	}
	state := domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{DataRecipient: dr})
	_, err := domain.HandleDataRecipient(state, domain.ReinstateAccreditation{})
	require.Error(t, err)

	return err
}

func certificatePEM(t *testing.T) (pemData []byte, thumbprint string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Recipient App Web"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	sum := sha256.Sum256(der)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), base64.RawURLEncoding.EncodeToString(sum[:])
}

const certificatesPath = "/v1/data-recipients/dr-1/software-products/sp-1/certificates"

func serveBody(t *testing.T, operator *fakeOperator, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	return serveMethodBody(t, operator, http.MethodPost, path, body)
}

func serveMethodBody(
	t *testing.T,
	operator *fakeOperator,
	method, path string,
	body []byte,
) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, operator, &fakeIssuer{}, &fakeRevoker{}, nil, &fakeKeys{}, access)

	req := httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(body))
	req.Header = authorised()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// Verifies: REG-0836, REG-0865
func TestACertificateIsRegisteredWithItsSHA256Thumbprint(t *testing.T) {
	pemData, thumbprint := certificatePEM(t)
	operator := &fakeOperator{}

	rec := serveBody(t, operator, certificatesPath, pemData)

	assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	assert.Equal(t, "dr-1", operator.legalEntityID)
	assert.Equal(
		t,
		domain.RegisterClientCertificate{SoftwareProductID: "sp-1", Thumbprint: thumbprint},
		operator.command,
	)
}

// Verifies: REG-0835
func TestABodyThatIsNotOnePEMCertificateIsRefused(t *testing.T) {
	pemData, _ := certificatePEM(t)

	tests := map[string][]byte{
		"empty":             nil,
		"not PEM":           []byte("not a certificate"),
		"not a certificate": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")}),
		"a private key":     pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("key")}),
		"two certificates":  append(append([]byte{}, pemData...), pemData...),
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			operator := &fakeOperator{}

			rec := serveBody(t, operator, certificatesPath, body)

			p := requireProblem(t, rec, http.StatusBadRequest)
			assert.Equal(t, "The request body must be one PEM certificate.", p.Detail)
			assert.Nil(t, operator.command)
		})
	}
}

// Verifies: REG-0854
func TestACertificateBodyLargerThan16KiBGives413(t *testing.T) {
	pemData, _ := certificatePEM(t)
	operator := &fakeOperator{}

	rec := serveBody(
		t,
		operator,
		certificatesPath,
		append(append([]byte{}, pemData...), bytes.Repeat([]byte("\n"), 1<<16)...),
	)

	p := requireProblem(t, rec, http.StatusRequestEntityTooLarge)
	assert.Equal(t, "The request body is larger than 16 KiB.", p.Detail)
	assert.Nil(t, operator.command)
}

func TestACertificateRequestWithoutTheOperatorTokenIsUnauthorised(t *testing.T) {
	pemData, _ := certificatePEM(t)
	operator := &fakeOperator{}
	mux := http.NewServeMux()
	operatorapi.Register(mux, operator, &fakeIssuer{}, &fakeRevoker{}, nil, &fakeKeys{}, access)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, certificatesPath, bytes.NewReader(pemData))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	requireProblem(t, rec, http.StatusUnauthorized)
	assert.Nil(t, operator.command)
}

const removalPath = certificatesPath + "/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

// Verifies: REG-0840, REG-0866
func TestARemovalExecutesItsCommandWithTheThumbprintOfThePath(t *testing.T) {
	operator := &fakeOperator{}

	rec := serve(t, operator, http.MethodDelete, removalPath, authorised())

	assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	assert.Empty(t, rec.Body.String())
	assert.Equal(t, "dr-1", operator.legalEntityID)
	assert.Equal(t, domain.RemoveClientCertificate{
		SoftwareProductID: "sp-1", Thumbprint: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}, operator.command)
}

// Verifies: REG-0841
func TestARemovalOfAThumbprintThatIsNotValidIsABadRequest(t *testing.T) {
	_, notValid := domain.HandleDataRecipient(
		domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{
			DataRecipient: domain.DataRecipient{
				LegalEntityID: "dr-1",
				Brands: []domain.DataRecipientBrand{{SoftwareProducts: []domain.SoftwareProduct{
					{SoftwareProductID: "sp-1"},
				}}},
			},
		}),
		domain.RemoveClientCertificate{SoftwareProductID: "sp-1", Thumbprint: "abc"},
	)
	require.ErrorIs(t, notValid, domain.ErrNotValid)

	rec := serve(t, &fakeOperator{err: notValid}, http.MethodDelete, certificatesPath+"/abc", authorised())

	p := requireProblem(t, rec, http.StatusBadRequest)
	assert.Equal(t, notValid.Error(), p.Detail)
}

func TestARemovalWithoutTheOperatorTokenIsUnauthorised(t *testing.T) {
	operator := &fakeOperator{}

	rec := serve(t, operator, http.MethodDelete, removalPath, nil)

	requireProblem(t, rec, http.StatusUnauthorized)
	assert.Nil(t, operator.command)
}

type fakeKeys struct {
	published, activated []app.KeySet
	err                  error
}

func (f *fakeKeys) PublishNextKey(_ context.Context, set app.KeySet) error {
	f.published = append(f.published, set)
	return f.err
}

func (f *fakeKeys) ActivateNextKey(_ context.Context, set app.KeySet) error {
	f.activated = append(f.activated, set)
	return f.err
}

func serveKeys(t *testing.T, keys *fakeKeys, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, &fakeOperator{}, &fakeIssuer{}, &fakeRevoker{}, nil, keys, access)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// Verifies: REG-0846, REG-0896
func TestTheKeyChangeActionsCallTheStepOfTheKeySet(t *testing.T) {
	tests := map[string]struct {
		path                 string
		published, activated []app.KeySet
	}{
		"publish SSA":           {"/v1/key-sets/ssa/publish-next", []app.KeySet{app.SSAKeys}, nil},
		"activate SSA":          {"/v1/key-sets/ssa/activate-next", nil, []app.KeySet{app.SSAKeys}},
		"publish access token":  {"/v1/key-sets/access-token/publish-next", []app.KeySet{app.AccessTokenKeys}, nil},
		"activate access token": {"/v1/key-sets/access-token/activate-next", nil, []app.KeySet{app.AccessTokenKeys}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			keys := &fakeKeys{}

			rec := serveKeys(t, keys, tt.path, authorised())

			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			assert.Equal(t, tt.published, keys.published)
			assert.Equal(t, tt.activated, keys.activated)
		})
	}
}

// Verifies: REG-0847, REG-0848, REG-0849
func TestKeyChangeErrors(t *testing.T) {
	tests := map[string]struct {
		path       string
		err        error
		status     int
		wantDetail string
	}{
		"change in progress": {
			"/v1/key-sets/ssa/publish-next",
			app.ErrKeyChangeInProgress,
			http.StatusConflict,
			"has a next key that is not active",
		},
		"no next key": {
			"/v1/key-sets/ssa/activate-next",
			app.ErrNoNextKey,
			http.StatusConflict,
			"has no next key",
		},
		"unknown key set": {
			"/v1/key-sets/other/publish-next",
			nil,
			http.StatusNotFound,
			"The key set does not exist",
		},
		"failure": {
			"/v1/key-sets/ssa/publish-next",
			errors.New("connection to 10.0.0.1 failed"),
			http.StatusInternalServerError,
			"could not complete",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			keys := &fakeKeys{err: tt.err}

			rec := serveKeys(t, keys, tt.path, authorised())

			p := requireProblem(t, rec, tt.status)
			assert.Contains(t, p.Detail, tt.wantDetail)
			assert.NotContains(t, p.Detail, "10.0.0.1")
		})
	}
}

func TestAKeyChangeWithoutTheOperatorTokenIsUnauthorised(t *testing.T) {
	keys := &fakeKeys{}

	rec := serveKeys(t, keys, "/v1/key-sets/ssa/publish-next", http.Header{})

	requireProblem(t, rec, http.StatusUnauthorized)
	assert.Empty(t, keys.published)
}
