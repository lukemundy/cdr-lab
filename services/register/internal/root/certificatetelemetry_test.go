package root

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/castore"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/idp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

// secretPEM gives a certificate and the texts that must not be in telemetry: each line of the PEM and the thumbprint.
func secretPEM(t *testing.T) (pemText, thumbprint string, secrets []string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Zebracorn Web"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	sum := sha256.Sum256(der)
	thumbprint = base64.RawURLEncoding.EncodeToString(sum[:])
	pemText = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))

	return pemText, thumbprint, append(pemLines(pemText), thumbprint)
}

func pemLines(text string) []string {
	var lines []string

	for line := range strings.Lines(text) {
		if !strings.HasPrefix(line, "-----") {
			lines = append(lines, strings.TrimSpace(line))
		}
	}

	return lines
}

// issuedSecrets gives each line of the PEM chain, and the thumbprint of the issued certificate.
func issuedSecrets(t *testing.T, chain []byte) []string {
	t.Helper()

	block, _ := pem.Decode(chain)
	require.NotNil(t, block)

	sum := sha256.Sum256(block.Bytes)

	return append(pemLines(string(chain)), base64.RawURLEncoding.EncodeToString(sum[:]))
}

const (
	revokePath             = "/v1/cdr-certificates/revoke"
	consoleCABundle        = "/api/cdr-ca-bundle"
	consoleCheckRevocation = "/api/cdr-certificates/check-revocation"
	consoleRevoke          = "/api/cdr-certificates/revoke"
	consoleRevocations     = "/api/cdr-certificates/revocations"
)

// The test does not check url.path, because recordTelemetry has no URL redactor to set it to the route.
func TestACertificateRequestPutsNoPEMTextNoSubjectAndNoThumbprintInTheSpansOrLogs(t *testing.T) {
	const (
		recipientID, productID  = "zebracorn-recipient", "zebracorn-product"
		holderID, holderBrandID = "zebracorn-holder", "zebracorn-holder-brand"
	)

	spans, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{{
		LegalEntityID: holderID, LegalEntityName: "Zebracorn Holdings Ltd", LogoURI: "https://holder.example/logo.png",
		Status: domain.DataHolderActive,
		Brands: []domain.DataHolderBrand{
			{
				DataHolderBrandID: holderBrandID,
				BrandName:         "Zebracorn Bank",
				Industries:        []domain.Industry{domain.Banking},
				LogoURI:           "https://holder.example/brand.png",
				Status:            domain.Active,
				EndpointDetail: domain.EndpointDetail{
					Version:         "1",
					PublicBaseURI:   "https://api.holder.example",
					ResourceBaseURI: "https://mtls.holder.example",
					InfosecBaseURI:  "https://id.holder.example",
					WebsiteURI:      "https://holder.example",
				},
				JWKSEndpoint: "https://holder.example/jwks",
			},
		},
	}}, DataRecipients: []domain.DataRecipient{{
		LegalEntityID: recipientID, LegalEntityName: "Zebracorn Pty Ltd", AccreditationNumber: "ADR-1",
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
		Status: domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "zebracorn-brand", BrandName: "Zebracorn", LogoURI: "https://recipient.example/b.png",
			Status: domain.Active,
			SoftwareProducts: []domain.SoftwareProduct{{
				SoftwareProductID: productID, SoftwareProductName: "Zebracorn Web",
				SoftwareProductDescription: "A product", LogoURI: "https://recipient.example/p.png",
				Status: domain.Active, ClientURI: "https://recipient.example",
				RedirectURIs: []string{"https://recipient.example/callback"}, JWKSURI: "https://recipient.example/jwks",
				RevocationURI: "https://recipient.example/revoke", RecipientBaseURI: "https://cdr.recipient.example",
				Scopes: []string{"openid", "profile"},
			}},
		}},
	}}}))

	caCert, caKey := tlsFiles(t)
	authority, err := ca.LoadAuthority(caCert, caKey, testCRLURL)
	require.NoError(t, err)

	operator := app.NewOperator(store)
	listeners, err := consoleListeners(testConsole, ui, console.Ports{
		Readers: readmodel.NewQueries(pool), Operator: operator,
		CertificateAuthority: app.NewCertificateAuthority(castore.New(pool), authority, time.Now),
		CABundle:             []*x509.Certificate{authority.Certificate},
	})
	require.NoError(t, err)

	operatorMux := http.NewServeMux()
	operatorapi.Register(
		operatorMux,
		operator,
		app.NewCertificateIssuer(store, authority, time.Now),
		app.NewCertificateAuthority(castore.New(pool), authority, time.Now),
		[]*x509.Certificate{authority.Certificate},
		idp.NewStore(pool),
		operatorhttp.Config{Token: "operator-token", AllowedHosts: []string{"example.com"}},
	)
	operatorHandler := telemetry.NewHTTPHandler(operatorMux)

	pemText, thumbprint, secrets := secretPEM(t)
	registration, err := json.Marshal(map[string]any{"version": 1, "certificate": pemText})
	require.NoError(t, err)
	notPEM, err := json.Marshal(map[string]any{"version": 1, "certificate": "zebracorn " + pemText})
	require.NoError(t, err)

	issueCSR := string(clientCSRPEM(t, "Zebracorn Web", "Zebracorn"))
	otherNamesCSR := string(clientCSRPEM(t, "Zebracorn Mobile", "Zebracorn Pty Ltd"))

	_, serverCSR := serverCSRPEM(t, "Zebracorn Bank", "zebracorn.holder.example", "localhost")
	_, otherServerCSR := serverCSRPEM(t, "Zebracorn Group", "zebracorn-mtls.holder.example")

	secrets = append(secrets, pemLines(issueCSR)...)
	secrets = append(secrets, pemLines(string(serverCSR))...)
	secrets = append(secrets, "zebracorn.holder.example", "zebracorn-mtls.holder.example", "Zebracorn Group")

	seedSpans := len(spans.Ended())

	consolePath := "/api/data-recipients/" + recipientID + "/software-products/" + productID + "/certificates"
	operatorPath := "/v1/data-recipients/" + recipientID + "/software-products/" + productID + "/certificates"
	serverIssuePath := holderBrandIssue(holderID, holderBrandID)

	requests := []struct {
		handler              http.Handler
		method, target, body string
		status               int
	}{
		{listeners[0].Handler, http.MethodPost, consolePath, string(registration), http.StatusNoContent},
		{listeners[0].Handler, http.MethodPost, consolePath, string(notPEM), http.StatusBadRequest},
		{
			listeners[0].Handler, http.MethodPost, consolePath + "/" + thumbprint + "/remove", `{"version": 2}`,
			http.StatusNoContent,
		},
		{operatorHandler, http.MethodPost, operatorPath, pemText, http.StatusNoContent},
		{operatorHandler, http.MethodPost, operatorPath, "zebracorn " + pemText, http.StatusBadRequest},
		{operatorHandler, http.MethodDelete, operatorPath + "/" + thumbprint, "", http.StatusNoContent},
		{operatorHandler, http.MethodPost, operatorPath + "/issue", issueCSR, http.StatusOK},
		{operatorHandler, http.MethodPost, operatorPath + "/issue", otherNamesCSR, http.StatusBadRequest},
		{operatorHandler, http.MethodPost, operatorPath + "/issue", "zebracorn " + issueCSR, http.StatusBadRequest},
		{operatorHandler, http.MethodPost, serverIssuePath, string(serverCSR), http.StatusOK},
		{operatorHandler, http.MethodPost, serverIssuePath, string(otherServerCSR), http.StatusBadRequest},
	}

	var issuedChain []byte

	send := func(handler http.Handler, method, target, body string, status int) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), method, target, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer operator-token")
		handler.ServeHTTP(rec, req)
		require.Equal(t, status, rec.Code, "%s %s: %s", method, target, rec.Body.String())

		if rec.Header().Get("Content-Type") == "application/pem-certificate-chain" {
			secrets = append(secrets, issuedSecrets(t, rec.Body.Bytes())...)
			issuedChain = rec.Body.Bytes()
		}
	}

	for _, r := range requests {
		send(r.handler, r.method, r.target, r.body, r.status)
	}

	revocation := func(certificate, reason string) string {
		data, err := json.Marshal(map[string]string{"certificate": certificate, "reason": reason})
		require.NoError(t, err)

		return string(data)
	}
	consoleHandler := listeners[0].Handler
	send(consoleHandler, http.MethodGet, consoleCABundle, "", http.StatusOK)
	send(consoleHandler, http.MethodPost, consoleCheckRevocation, revocation(string(issuedChain), "keyCompromise"),
		http.StatusOK)
	send(consoleHandler, http.MethodPost, consoleCheckRevocation,
		revocation("zebracorn "+string(issuedChain), "zebracorn"), http.StatusBadRequest)
	send(consoleHandler, http.MethodPost, consoleRevoke, revocation(string(issuedChain), "keyCompromise"),
		http.StatusNoContent)
	send(consoleHandler, http.MethodPost, consoleRevoke, revocation(string(issuedChain), "superseded"),
		http.StatusConflict)
	send(consoleHandler, http.MethodGet, consoleRevocations, "", http.StatusOK)
	send(operatorHandler, http.MethodPost, revokePath, revocation(string(issuedChain), "superseded"),
		http.StatusConflict)
	send(operatorHandler, http.MethodPost, revokePath, revocation("zebracorn "+string(issuedChain), "zebracorn"),
		http.StatusBadRequest)

	routes := map[string]int{}

	for _, span := range spans.Ended()[seedSpans:] {
		parts := []string{span.Name(), span.Status().Description}

		for _, a := range span.Attributes() {
			switch a.Key {
			case semconv.URLPathKey:
			case semconv.HTTPRouteKey:
				routes[a.Value.AsString()]++
			default:
				parts = append(parts, a.Value.String())
			}
		}

		for _, e := range span.Events() {
			parts = append(parts, e.Name)
			for _, a := range e.Attributes {
				parts = append(parts, a.Value.String())
			}
		}

		text := strings.Join(parts, "\n")
		for _, secret := range secrets {
			assert.NotContains(t, text, secret, "span %q", span.Name())
		}

		assert.NotContains(t, strings.ToLower(text), "zebracorn", "span %q", span.Name())
	}

	consoleRoute := "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates"
	operatorRoute := "/v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates"
	assert.Equal(t, map[string]int{
		consoleRoute:                          2,
		consoleRoute + "/{thumbprint}/remove": 1,
		operatorRoute:                         2,
		operatorRoute + "/{thumbprint}":       1,
		operatorRoute + "/issue":              3,
		"/v1/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue": 2,
		revokePath:             2,
		consoleCABundle:        1,
		consoleCheckRevocation: 2,
		consoleRevoke:          2,
		consoleRevocations:     1,
	}, routes, "each server span has the route pattern")

	for _, secret := range append(secrets, "zebracorn") {
		assert.NotContains(t, strings.ToLower(logs.String()), strings.ToLower(secret))
	}
}
