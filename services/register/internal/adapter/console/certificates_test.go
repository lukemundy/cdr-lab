package console_test

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
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	certificates = "/api/data-recipients/dr-1/software-products/sp-1/certificates"
	notOnePEM    = "The certificate must be one PEM certificate, with no other text."
)

func certificatePEM(t *testing.T) (pemText, thumbprint string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Product sp-1"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	sum := sha256.Sum256(der)

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		base64.RawURLEncoding.EncodeToString(sum[:])
}

func registration(t *testing.T, version int64, certificate string) string {
	t.Helper()

	body, err := json.Marshal(map[string]any{"version": version, "certificate": certificate})
	require.NoError(t, err)

	return string(body)
}

// The seed data recipient has an ACTIVE and a REMOVED software product.
func seedForCertificates(t *testing.T) console.Ports {
	t.Helper()

	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Removed)}),
	}})

	return console.Ports{Readers: queries, Operator: app.NewOperator(store)}
}

type certificatesBody struct {
	Version int64 `json:"version"`
	Brands  []struct {
		SoftwareProducts []struct {
			Thumbprints []string `json:"clientCertificateThumbprints"`
			Actions     []string `json:"clientCertificateActions"`
		} `json:"softwareProducts"`
	} `json:"dataRecipientBrands"`
}

func certificatesOf(t *testing.T, ports console.Ports) certificatesBody {
	t.Helper()

	resp := get(t, ports.Readers, "/api/data-recipients/dr-1")
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var body certificatesBody
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	return body
}

func problemDetail(t *testing.T, resp response) string {
	t.Helper()

	var p struct {
		Detail string `json:"detail"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &p), resp.body)

	return p.Detail
}

func TestADataRecipientDetailGivesThePermittedCertificateActionsOfEachSoftwareProduct(t *testing.T) {
	products := certificatesOf(t, seedForCertificates(t)).Brands[0].SoftwareProducts

	assert.Equal(t, []string{"register", "issue", "remove"}, products[0].Actions)
	assert.Equal(t, []string{"remove"}, products[1].Actions, "a REMOVED software product cannot get a certificate")
}

// Verifies: REG-1310, REG-1332
func TestAPEMCertificateIsRegisteredWithItsThumbprintAndThenRemoved(t *testing.T) {
	ports := seedForCertificates(t)
	pemText, thumbprint := certificatePEM(t)

	resp := post(t, ports, certificates, registration(t, 1, pemText))

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)
	after := certificatesOf(t, ports)
	assert.Equal(t, []string{thumbprint}, after.Brands[0].SoftwareProducts[0].Thumbprints)

	resp = post(t, ports, certificates+"/"+thumbprint+"/remove", fmt.Sprintf(`{"version": %d}`, after.Version))

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, certificatesOf(t, ports).Brands[0].SoftwareProducts[0].Thumbprints)
}

func TestARemovalOfAThumbprintThatIsNotRegisteredChangesNothing(t *testing.T) {
	ports := seedForCertificates(t)
	_, thumbprint := certificatePEM(t)

	resp := post(t, ports, certificates+"/"+thumbprint+"/remove", `{"version": 1}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Equal(t, int64(1), certificatesOf(t, ports).Version)
}

func TestTextThatIsNotOnePEMCertificateIsRefusedWithTheRuleOfTheField(t *testing.T) {
	ports := seedForCertificates(t)
	pemText, _ := certificatePEM(t)

	for name, text := range map[string]string{
		"empty":             "",
		"not PEM":           "not a certificate",
		"not a certificate": string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")})),
		"two certificates":  pemText + pemText,
	} {
		resp := post(t, ports, certificates, registration(t, 1, text))

		require.Equal(t, http.StatusBadRequest, resp.status, name)
		assert.Equal(t, notOnePEM, problemDetail(t, resp), name)
	}

	assert.Equal(t, int64(1), certificatesOf(t, ports).Version)
}

func TestACertificateRequestRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForCertificates(t)
	pemText, thumbprint := certificatePEM(t)

	for _, body := range []string{
		``, `{}`, `{"version": 1}`, `{"certificate": "x"}`, `{"version": 0, "certificate": "x"}`,
		`{"version": 1, "certificate": 1}`,
		`{"version": 1, "certificate": "x", "thumbprint": "x"}`,
	} {
		resp := post(t, ports, certificates, body)

		assert.Equal(t, http.StatusBadRequest, resp.status, "body %.40s", body)
		assert.NotEqual(t, notOnePEM, problemDetail(t, resp), "body %.40s", body)
	}

	resp := post(t, ports, certificates+"/"+thumbprint+"/remove", registration(t, 1, pemText))
	assert.Equal(t, http.StatusBadRequest, resp.status, "a removal has only the version")
	assert.Equal(t, int64(1), certificatesOf(t, ports).Version)
}

func TestACertificateActionThatTheRulesDoNotPermitGivesConflict(t *testing.T) {
	ports := seedForCertificates(t)
	pemText, thumbprint := certificatePEM(t)

	tests := map[string]struct{ target, body, detail string }{
		"registration for a REMOVED software product": {
			"/api/data-recipients/dr-1/software-products/sp-2/certificates", registration(t, 1, pemText),
			"cannot get a client certificate: REMOVED is final",
		},
		"registration at a version that is not current": {
			certificates, registration(t, 2, pemText), "The participant changed after the version in the request.",
		},
		"removal at a version that is not current": {
			certificates + "/" + thumbprint + "/remove", `{"version": 2}`,
			"The participant changed after the version in the request.",
		},
	}
	for name, tt := range tests {
		resp := post(t, ports, tt.target, tt.body)

		require.Equal(t, http.StatusConflict, resp.status, name)
		assert.Contains(t, problemDetail(t, resp), tt.detail, name)
	}

	assert.Equal(t, int64(1), certificatesOf(t, ports).Version)
}

func TestACertificateActionForARecordThatDoesNotExistOrAThumbprintThatIsNotValidGivesNoIdentifier(t *testing.T) {
	ports := seedForCertificates(t)
	pemText, thumbprint := certificatePEM(t)

	tests := map[string]struct {
		target, body string
		status       int
	}{
		"unknown data recipient": {
			"/api/data-recipients/dr-unknown/software-products/sp-1/certificates", registration(t, 1, pemText),
			http.StatusNotFound,
		},
		"unknown software product": {
			"/api/data-recipients/dr-1/software-products/sp-unknown/certificates/" + thumbprint + "/remove",
			`{"version": 1}`, http.StatusNotFound,
		},
		"thumbprint that is not valid": {
			certificates + "/unknown-thumbprint/remove", `{"version": 1}`, http.StatusBadRequest,
		},
	}
	for name, tt := range tests {
		resp := post(t, ports, tt.target, tt.body)

		require.Equal(t, tt.status, resp.status, name)
		assert.NotContains(t, resp.body, "unknown", name)
	}
}
