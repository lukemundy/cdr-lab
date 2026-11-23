package root

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/seedfile"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

const (
	testCRLURL       = "http://ca.example/crl/intermediate.crl"
	labOperatorToken = "test-operator-token"
	operatorHost     = "register.test:8080"
)

// Participants of seed/example.json.
const (
	statusChangeRecipient = "9a99b5d7-fd53-49e7-8a1c-152fcd8f380b"
	statusChangeBrand     = "72afef7f-716a-463f-b007-10a4313d5e96" // ACTIVE, with statusChangeWeb
	statusChangeWeb       = "89fcb232-f453-4f9a-9ea9-5655d9203c3c"
	statusChangeMobile    = "b442771c-9c80-4b07-8249-f4aedfe8a2e0"
	statusChangeBeta      = "385e5475-0448-46b9-a740-cec9213b95d9"
	statusChangeLegacy    = "a0dfca7e-4688-473d-b2a9-a3c2657879f0" // under an INACTIVE brand
	surrenderRecipient    = "416448ea-d963-4391-89a7-6de950b6bc91"
	surrenderWeb          = "7bae25e7-ec64-445e-9e10-e9717bd2363d"
	mockBudgetRecipient   = "c7b88455-4a14-479e-b53e-fbc4000716b3"
	mockBudgetWeb         = "66b83c54-e4f4-434d-aa9b-789f64358022" // ACTIVE
	mockBudgetBeta        = "6d0b646f-475d-4c7b-af85-0eb558f50d47" // REMOVED
	unknownID             = "00000000-0000-0000-0000-000000000000"
)

func setLabSettings(t *testing.T) {
	t.Helper()

	certFile, keyFile := tlsFiles(t)
	for name, value := range map[string]string{
		"PUBLIC_BASE_URL":        "https://register.example",
		"SECURE_BASE_URL":        "https://secure.register.example",
		"OPERATOR_TOKEN":         labOperatorToken,
		"INTERNAL_ALLOWED_HOSTS": operatorHost,
		"PUBLIC_TLS_CERT_FILE":   certFile,
		"PUBLIC_TLS_KEY_FILE":    keyFile,
		"SECURE_TLS_DNS_NAMES":   "register,mtls.register.example",
		"CDR_CA_BUNDLE_FILE":     certFile,
		// The one certificate of the bundle is also the CA that issues.
		"CDR_INTERMEDIATE_CERT_FILE": certFile,
		"CDR_INTERMEDIATE_KEY_FILE":  keyFile,
		"CDR_INTERMEDIATE_CRL_URL":   testCRLURL,
	} {
		t.Setenv(name, value)
	}
}

// lab gives the handler of each listener of a Register that has the participants of seed/example.json.
func lab(t *testing.T) map[int]http.Handler {
	t.Helper()

	handlers, _ := labWithPool(t)

	return handlers
}

func labWithPool(t *testing.T) (map[int]http.Handler, *pgxpool.Pool) {
	t.Helper()

	setLabSettings(t)

	cfg, err := loadConfig(serveMode)
	require.NoError(t, err)

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	f, err := os.Open("../../seed/example.json")
	require.NoError(t, err)

	defer func() { _ = f.Close() }()

	participants, err := seedfile.Read(f)
	require.NoError(t, err)
	require.NoError(t, app.Seed(t.Context(), newEventStore(pool), participants))

	proc, err := newProcess(t.Context(), pool, cfg, serveMode)
	require.NoError(t, err)

	handlers := map[int]http.Handler{}
	for _, listener := range proc.listeners {
		handlers[listener.Port] = listener.Handler
	}

	return handlers, pool
}

// tlsFiles writes one self-signed certificate, which is also a CA bundle, and its key.
func tlsFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "register.example"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	dir := t.TempDir()
	certFile, keyFile = filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key")
	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0o600))

	return certFile, keyFile
}

type operatorProblem struct {
	Type   string `json:"type"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func serveOperator(
	t *testing.T,
	handlers map[int]http.Handler,
	path string,
	header http.Header,
	body []byte,
) *httptest.ResponseRecorder {
	t.Helper()

	return serveOperatorHost(t, handlers, operatorHost, http.MethodPost, path, header, body)
}

func serveOperatorHost(
	t *testing.T,
	handlers map[int]http.Handler,
	requestHost, method, path string,
	header http.Header,
	body []byte,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, "http://"+requestHost+path, bytes.NewReader(body))
	req.Header = header
	rec := httptest.NewRecorder()
	handlers[internalPort].ServeHTTP(rec, req)

	return rec
}

func authorisedOperator() http.Header {
	return http.Header{"Authorization": {"Bearer " + labOperatorToken}}
}

func operate(t *testing.T, handlers map[int]http.Handler, path string, body []byte) (int, operatorProblem) {
	t.Helper()

	rec := serveOperator(t, handlers, path, authorisedOperator(), body)

	return rec.Code, problemOf(t, rec)
}

func problemOf(t *testing.T, rec *httptest.ResponseRecorder) operatorProblem {
	t.Helper()

	var p operatorProblem

	if rec.Code != http.StatusNoContent {
		assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p), rec.Body.String())
		assert.Equal(t, "about:blank", p.Type)
		assert.Equal(t, rec.Code, p.Status)
	}

	return p
}

func recipientAction(id, action string) string {
	return "/v1/data-recipients/" + id + "/" + action
}

func productAction(recipient, product, action string) string {
	return "/v1/data-recipients/" + recipient + "/software-products/" + product + "/" + action
}

func refused(t *testing.T, handlers map[int]http.Handler, path string, wantStatus int, wantDetail string) {
	t.Helper()

	status, p := operate(t, handlers, path, nil)
	assert.Equal(t, wantStatus, status)
	assert.Contains(t, p.Detail, wantDetail)
}

const (
	recipientStatusList = "data recipient statuses"
	softwareProductList = "software product statuses"
	dataRecipientList   = "data recipients"
)

// publicView is the statuses on the public APIs, keyed by API and then by ID, and the ETag of each API.
type publicView struct {
	etags      map[string]string
	recipients map[string]map[string]string
	products   map[string]map[string]string
}

func getPublic(t *testing.T, handlers map[int]http.Handler, path, version string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	req.Header.Set("X-V", version)

	rec := httptest.NewRecorder()
	handlers[apiPort].ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	return rec
}

func statusesOf(t *testing.T, rec *httptest.ResponseRecorder, idField string) map[string]string {
	t.Helper()

	var list struct {
		Data []map[string]string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))

	statuses := map[string]string{}
	for _, item := range list.Data {
		statuses[item[idField]] = item["status"]
	}

	return statuses
}

func observe(t *testing.T, handlers map[int]http.Handler) publicView {
	t.Helper()

	v := publicView{
		etags:      map[string]string{},
		recipients: map[string]map[string]string{},
		products:   map[string]map[string]string{},
	}

	recipientStatuses := getPublic(t, handlers, "/cdr-register/v1/all/data-recipients/status", "3")
	v.etags[recipientStatusList] = recipientStatuses.Header().Get("ETag")
	v.recipients[recipientStatusList] = statusesOf(t, recipientStatuses, "legalEntityId")

	productStatuses := getPublic(
		t,
		handlers,
		"/cdr-register/v1/all/data-recipients/brands/software-products/status",
		"3",
	)
	v.etags[softwareProductList] = productStatuses.Header().Get("ETag")
	v.products[softwareProductList] = statusesOf(t, productStatuses, "softwareProductId")

	recipients := getPublic(t, handlers, "/cdr-register/v1/all/data-recipients", "4")
	v.etags[dataRecipientList] = recipients.Header().Get("ETag")

	var body struct {
		Data []struct {
			LegalEntityID string `json:"legalEntityId"`
			Status        string `json:"status"`
			Brands        []struct {
				SoftwareProducts []struct {
					SoftwareProductID string `json:"softwareProductId"`
					Status            string `json:"status"`
				} `json:"softwareProducts"`
			} `json:"dataRecipientBrands"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recipients.Body.Bytes(), &body))

	v.recipients[dataRecipientList], v.products[dataRecipientList] = map[string]string{}, map[string]string{}
	for _, dr := range body.Data {
		v.recipients[dataRecipientList][dr.LegalEntityID] = dr.Status
		for _, brand := range dr.Brands {
			for _, sp := range brand.SoftwareProducts {
				v.products[dataRecipientList][sp.SoftwareProductID] = sp.Status
			}
		}
	}

	return v
}

type statusChange struct {
	action string
	// recipients and products give the statuses that the public APIs must show after the action, by ID.
	recipients, products map[string]string
	changedAPIs          []string
}

// change checks that the public APIs show the status change at the end of the operator request, with a new ETag on
// each changed API.
func change(t *testing.T, handlers map[int]http.Handler, c statusChange) {
	t.Helper()

	before := observe(t, handlers)

	status, p := operate(t, handlers, c.action, nil)
	require.Equal(t, http.StatusNoContent, status, p.Detail)

	after := observe(t, handlers)

	for _, list := range []string{recipientStatusList, dataRecipientList} {
		for id, want := range c.recipients {
			assert.Equal(t, want, after.recipients[list][id], "%s: data recipient %s", list, id)
		}
	}

	for _, list := range []string{softwareProductList, dataRecipientList} {
		for id, want := range c.products {
			assert.Equal(t, want, after.products[list][id], "%s: software product %s", list, id)
		}
	}

	for _, api := range c.changedAPIs {
		assert.NotEqual(t, before.etags[api], after.etags[api], "the ETag of %s did not change", api)
	}
}

// Verifies: REG-0805, REG-0806
func TestTheOperatorChangesStatusesWithTheStatusRules(t *testing.T) {
	handlers := lab(t)

	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{name: "deactivate a software product", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeWeb, "deactivate"),
				products:    map[string]string{statusChangeWeb: "INACTIVE"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "activate a software product", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeWeb, "activate"),
				products:    map[string]string{statusChangeWeb: "ACTIVE"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "remove a software product", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeBeta, "remove"),
				products:    map[string]string{statusChangeBeta: "REMOVED"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "a removed software product cannot change", run: func(t *testing.T) {
			refused(t, handlers, productAction(statusChangeRecipient, statusChangeBeta, "activate"),
				http.StatusConflict, "REMOVED is final")
		}},
		{name: "a software product cannot be more active than its brand", run: func(t *testing.T) {
			refused(
				t,
				handlers,
				productAction(statusChangeRecipient, statusChangeLegacy, "activate"),
				http.StatusConflict,
				"a software product of an INACTIVE data recipient brand must be INACTIVE or REMOVED",
			)
		}},
		{name: "suspend makes active software products inactive", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:     recipientAction(statusChangeRecipient, "suspend"),
				recipients: map[string]string{statusChangeRecipient: "SUSPENDED"},
				products: map[string]string{
					statusChangeWeb: "INACTIVE", statusChangeMobile: "INACTIVE",
					statusChangeBeta: "REMOVED", statusChangeLegacy: "INACTIVE",
				},
				changedAPIs: []string{recipientStatusList, softwareProductList, dataRecipientList},
			})
		}},
		{name: "a software product of a suspended data recipient cannot be active", run: func(t *testing.T) {
			refused(t, handlers, productAction(statusChangeRecipient, statusChangeWeb, "activate"), http.StatusConflict,
				"a software product of a SUSPENDED data recipient must be INACTIVE or REMOVED")
		}},
		{name: "a suspended data recipient cannot be suspended", run: func(t *testing.T) {
			refused(t, handlers, recipientAction(statusChangeRecipient, "suspend"), http.StatusConflict,
				"from SUSPENDED, the allowed statuses are ACTIVE, REVOKED and SURRENDERED")
		}},
		{name: "reinstate does not change software products", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:      recipientAction(statusChangeRecipient, "reinstate"),
				recipients:  map[string]string{statusChangeRecipient: "ACTIVE"},
				products:    map[string]string{statusChangeWeb: "INACTIVE", statusChangeMobile: "INACTIVE"},
				changedAPIs: []string{recipientStatusList, dataRecipientList},
			})
		}},
		{name: "activate a software product after reinstatement", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeWeb, "activate"),
				products:    map[string]string{statusChangeWeb: "ACTIVE"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "revoke removes software products", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:     recipientAction(statusChangeRecipient, "revoke"),
				recipients: map[string]string{statusChangeRecipient: "REVOKED"},
				products: map[string]string{
					statusChangeWeb: "REMOVED", statusChangeMobile: "REMOVED", statusChangeLegacy: "REMOVED",
				},
				changedAPIs: []string{recipientStatusList, softwareProductList, dataRecipientList},
			})
		}},
		{name: "a revoked data recipient can be reinstated", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:      recipientAction(statusChangeRecipient, "reinstate"),
				recipients:  map[string]string{statusChangeRecipient: "ACTIVE"},
				products:    map[string]string{statusChangeWeb: "REMOVED"},
				changedAPIs: []string{recipientStatusList, dataRecipientList},
			})
		}},
		{name: "surrender removes software products", run: func(t *testing.T) {
			change(t, handlers, statusChange{
				action:      recipientAction(surrenderRecipient, "surrender"),
				recipients:  map[string]string{surrenderRecipient: "SURRENDERED"},
				products:    map[string]string{surrenderWeb: "REMOVED"},
				changedAPIs: []string{recipientStatusList, softwareProductList, dataRecipientList},
			})
		}},
		{name: "a surrendered data recipient cannot change", run: func(t *testing.T) {
			refused(t, handlers, recipientAction(surrenderRecipient, "reinstate"), http.StatusConflict,
				"SURRENDERED is final")
		}},
	}
	for _, step := range steps {
		if !t.Run(step.name, step.run) {
			t.Fatalf("stopped at %q: the next steps depend on it", step.name)
		}
	}
}

// Verifies: REG-0861
func TestTheOperatorCannotChangeARecordThatDoesNotExist(t *testing.T) {
	handlers := lab(t)

	refused(t, handlers, recipientAction(unknownID, "suspend"), http.StatusNotFound, "does not exist")
	refused(t, handlers, productAction(statusChangeRecipient, unknownID, "deactivate"), http.StatusNotFound,
		"does not exist")
}

func TestTheOperatorInterfaceRefusesARequestWithoutTheOperatorTokenAndChangesNothing(t *testing.T) {
	handlers := lab(t)

	for name, tt := range map[string]struct {
		header    http.Header
		challenge string
	}{
		"no token": {http.Header{}, `Bearer`},
		"wrong token": {
			http.Header{"Authorization": {"Bearer not-the-operator-token"}}, `Bearer error="invalid_token"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			before := observe(t, handlers)

			rec := serveOperator(t, handlers, recipientAction(statusChangeRecipient, "suspend"), tt.header, nil)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			problemOf(t, rec)
			assert.Equal(t, tt.challenge, rec.Header().Get("WWW-Authenticate"))
			assert.Equal(t, before, observe(t, handlers), "a refused request changed a status")
		})
	}
}

func TestTheOperatorInterfaceRefusesEachRequestForAHostThatIsNotAllowedAndChangesNothing(t *testing.T) {
	handlers := lab(t)
	certificates := productAction(mockBudgetRecipient, mockBudgetWeb, "certificates")
	holder := "/v1/data-holders/" + seedDataHolder
	holderBrand := holder + "/brands/" + seedHolderBrand
	recipient := "/v1/data-recipients/" + statusChangeRecipient
	recipientBrand := recipient + "/brands/" + statusChangeBrand
	product := recipient + "/software-products/" + statusChangeWeb

	requests := map[string]struct {
		method, path string
		body         []byte
	}{
		"publish the next key":  {method: http.MethodPost, path: "/v1/key-sets/ssa/publish-next"},
		"activate the next key": {method: http.MethodPost, path: "/v1/key-sets/ssa/activate-next"},
		"suspend a recipient":   {method: http.MethodPost, path: recipientAction(statusChangeRecipient, "suspend")},
		"deactivate a product": {
			method: http.MethodPost,
			path:   productAction(statusChangeRecipient, statusChangeWeb, "deactivate"),
		},
		"register a certificate": {method: http.MethodPost, path: certificates, body: clientCertificatePEM(t)},
		"remove a certificate":   {method: http.MethodDelete, path: certificates + "/" + strings.Repeat("a", 64)},
		"create a data holder": {
			method: http.MethodPost,
			path:   "/v1/data-holders",
			body:   []byte(newDataHolderJSON(newDataHolder, newBrandJSON(newBrand, "New Brand"))),
		},
		"change a data holder": {
			method: http.MethodPut, path: holder, body: []byte(`{"legalEntityName": "Changed Holder Ltd"}`),
		},
		"remove a data holder": {method: http.MethodPost, path: holder + "/remove"},
		"add a brand": {
			method: http.MethodPost, path: holder + "/brands", body: []byte(newBrandJSON(newBrand, "New Brand")),
		},
		"change a brand":     {method: http.MethodPut, path: holderBrand, body: []byte(`{"brandName": "Changed"}`)},
		"activate a brand":   {method: http.MethodPost, path: holderBrand + "/activate"},
		"deactivate a brand": {method: http.MethodPost, path: holderBrand + "/deactivate"},
		"remove a brand":     {method: http.MethodPost, path: holderBrand + "/remove"},
		"unknown path":       {method: http.MethodGet, path: "/v1/unknown"},
		"create a data recipient": {
			method: http.MethodPost, path: "/v1/data-recipients", body: []byte(newRecipientJSON()),
		},
		"change a data recipient": {
			method: http.MethodPut, path: recipient, body: []byte(`{"legalEntityName": "Changed Pty Ltd"}`),
		},
		"add a data recipient brand": {
			method: http.MethodPost, path: recipient + "/brands",
			body: []byte(`{"dataRecipientBrandId": "` + newRecipientBrand + `", "brandName": "New",
				"logoUri": "https://new.example/brand.png", "status": "ACTIVE", "softwareProducts": []}`),
		},
		"change a data recipient brand": {
			method: http.MethodPut, path: recipientBrand, body: []byte(`{"brandName": "Changed"}`),
		},
		"activate a data recipient brand":   {method: http.MethodPost, path: recipientBrand + "/activate"},
		"deactivate a data recipient brand": {method: http.MethodPost, path: recipientBrand + "/deactivate"},
		"remove a data recipient brand":     {method: http.MethodPost, path: recipientBrand + "/remove"},
		"add a software product": {
			method: http.MethodPost, path: recipientBrand + "/software-products",
			body: []byte(newProductJSON(newProduct, "New App Web")),
		},
		"change a software product": {
			method: http.MethodPut, path: product, body: []byte(`{"softwareProductName": "Changed"}`),
		},
	}
	for name, req := range requests {
		for hostName, requestHost := range map[string]string{
			"DNS rebinding name": "attacker.example:8080",
			"no port":            "register.test",
		} {
			t.Run(name+" "+hostName, func(t *testing.T) {
				before, holdersBefore := observe(t, handlers), observeHolders(t, handlers)
				namesBefore := productNames(t, handlers)

				rec := serveOperatorHost(
					t, handlers, requestHost, req.method, req.path, authorisedOperator(), req.body,
				)

				assert.Equal(t, http.StatusForbidden, rec.Code)
				problemOf(t, rec)
				assert.Equal(t, before, observe(t, handlers), "a refused request changed a data recipient")
				assert.Equal(t, holdersBefore, observeHolders(t, handlers), "a refused request changed a data holder")
				assert.Equal(t, namesBefore, productNames(t, handlers), "a refused request changed a software product")
			})
		}
	}
}

// Verifies: REG-1509
func TestTheRegisterDoesNotStartWithoutTheAllowedHostsOfTheOperatorInterface(t *testing.T) {
	setLabSettings(t)
	t.Setenv("INTERNAL_ALLOWED_HOSTS", "")

	_, err := loadConfig(serveMode)

	assert.ErrorContains(t, err, "INTERNAL_ALLOWED_HOSTS")
}

// The operator interface does not check the issuer, so a self-signed certificate is sufficient.
func clientCertificatePEM(t *testing.T) []byte {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	require.NoError(t, err)

	tmpl := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: "Mock Budget Web"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// Verifies: REG-0834, REG-0838
func TestTheOperatorRegistersAClientCertificateForASoftwareProductOneOrMoreTimes(t *testing.T) {
	handlers := lab(t)
	path := productAction(mockBudgetRecipient, mockBudgetWeb, "certificates")
	certificate := clientCertificatePEM(t)

	status, p := operate(t, handlers, path, certificate)
	require.Equal(t, http.StatusNoContent, status, p.Detail)

	status, p = operate(t, handlers, path, certificate)
	assert.Equal(t, http.StatusNoContent, status, "the same certificate again is not an error: %s", p.Detail)
}

// Verifies: REG-0835, REG-0839
func TestTheOperatorCannotRegisterACertificateThatIsNotValidOrForARecordThatCannotUseIt(t *testing.T) {
	handlers := lab(t)
	certificate := clientCertificatePEM(t)

	tests := map[string]struct {
		path       string
		body       []byte
		wantStatus int
		wantDetail string
	}{
		"not a certificate": {
			productAction(mockBudgetRecipient, mockBudgetWeb, "certificates"),
			[]byte("not a certificate"),
			http.StatusBadRequest,
			"one PEM certificate",
		},
		"no body": {
			productAction(mockBudgetRecipient, mockBudgetWeb, "certificates"),
			nil,
			http.StatusBadRequest,
			"one PEM certificate",
		},
		"unknown data recipient": {
			productAction(unknownID, mockBudgetWeb, "certificates"), certificate, http.StatusNotFound, "does not exist",
		},
		"unknown software product": {
			productAction(mockBudgetRecipient, unknownID, "certificates"), certificate, http.StatusNotFound,
			"does not exist",
		},
		"removed software product": {
			productAction(mockBudgetRecipient, mockBudgetBeta, "certificates"), certificate, http.StatusConflict,
			"REMOVED is final",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			status, p := operate(t, handlers, tt.path, tt.body)

			assert.Equal(t, tt.wantStatus, status)
			assert.Contains(t, p.Detail, tt.wantDetail)
		})
	}
}

// Verifies: REG-0846, REG-0847, REG-0848
func TestTheOperatorChangesAKeySetOnlyInTheStepsOfAKeyChange(t *testing.T) {
	handlers := lab(t)

	for _, keySet := range []string{"ssa", "access-token"} {
		t.Run(keySet, func(t *testing.T) {
			path := func(action string) string { return "/v1/key-sets/" + keySet + "/" + action }

			status, p := operate(t, handlers, path("publish-next"), nil)
			require.Equal(t, http.StatusNoContent, status, p.Detail)
			refused(t, handlers, path("publish-next"), http.StatusConflict, "has a next key that is not active")

			status, p = operate(t, handlers, path("activate-next"), nil)
			require.Equal(t, http.StatusNoContent, status, p.Detail)
			refused(t, handlers, path("activate-next"), http.StatusConflict, "has no next key")
		})
	}
}

// Verifies: REG-0849
func TestTheOperatorCannotChangeAKeySetThatDoesNotExist(t *testing.T) {
	handlers := lab(t)

	refused(t, handlers, "/v1/key-sets/other/publish-next", http.StatusNotFound, "The key set does not exist")
	refused(t, handlers, "/v1/key-sets/other/activate-next", http.StatusNotFound, "The key set does not exist")
}

// Verifies: REG-0008
func TestHealthProbesReturnOK(t *testing.T) {
	handlers := lab(t)

	for _, path := range []string{"/livez", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handlers[internalPort].ServeHTTP(
				rec,
				httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil),
			)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
