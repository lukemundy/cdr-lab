package root

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
)

func crlRows(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()

	var n int
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT count(*) FROM cdr_crl").Scan(&n))

	return n
}

// startRegister gives the process of one start of the Register with the lab settings, on the database of pool.
func startRegister(t *testing.T, pool *pgxpool.Pool) process {
	t.Helper()

	setLabSettings(t)
	require.NoError(t, migrate(t.Context(), pool))

	cfg, err := loadConfig(serveMode)
	require.NoError(t, err)

	proc, err := newProcess(t.Context(), pool, cfg, serveMode)
	require.NoError(t, err)

	return proc
}

func listener(t *testing.T, proc process, port int) (http.Handler, *tls.Config) {
	t.Helper()

	for _, l := range proc.listeners {
		if l.Port == port {
			return l.Handler, l.TLSConfig
		}
	}

	t.Fatalf("the process has no listener on port %d", port)

	return nil, nil
}

func intermediateAuthority(t *testing.T) ca.Authority {
	t.Helper()

	authority, err := ca.LoadAuthority(
		os.Getenv("CDR_INTERMEDIATE_CERT_FILE"), os.Getenv("CDR_INTERMEDIATE_KEY_FILE"), testCRLURL,
	)
	require.NoError(t, err)

	return authority
}

func getCRL(t *testing.T, proc process) *x509.RevocationList {
	t.Helper()

	handler, _ := listener(t, proc, internalPort)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(
		rec,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/crl/cdr-intermediate.crl", nil),
	)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/pkix-crl", rec.Header().Get("Content-Type"))

	crl, err := x509.ParseRevocationList(rec.Body.Bytes())
	require.NoError(t, err)
	require.NoError(t, crl.CheckSignatureFrom(intermediateAuthority(t).Certificate))

	return crl
}

func TestTheCRLListenerGivesACurrentCRLThatTheIntermediateSigned(t *testing.T) {
	crl := getCRL(t, startRegister(t, databasetest.New(t)))

	assert.Equal(t, big.NewInt(crl.ThisUpdate.Unix()), crl.Number, "the number is the time of the signature")
	assert.WithinDuration(t, time.Now(), crl.ThisUpdate, time.Minute)
	assert.Equal(t, 7*24*time.Hour, crl.NextUpdate.Sub(crl.ThisUpdate))
	assert.Empty(t, crl.RevokedCertificateEntries)
}

// Verifies: REG-0711
func TestEachStartSignsACRLWithAHigherNumber(t *testing.T) {
	pool := databasetest.New(t)

	previous := big.NewInt(0)

	for range 3 {
		number := getCRL(t, startRegister(t, pool)).Number
		assert.Equal(t, 1, number.Cmp(previous), "%v is not higher than %v", number, previous)

		previous = number
	}
}

func revokeRequest(t *testing.T, proc process, certificate []byte, reason string) (int, operatorProblem) {
	t.Helper()

	body, err := json.Marshal(map[string]string{"certificate": string(certificate), "reason": reason})
	require.NoError(t, err)

	handler, _ := listener(t, proc, internalPort)

	return operate(t, map[int]http.Handler{internalPort: handler}, "/v1/cdr-certificates/revoke", body)
}

// Verifies: REG-0003, REG-0725
func TestARevocationTakesEffectAtOnceAtTheRegisterAndStaysAfterARestart(t *testing.T) {
	pool := databasetest.New(t)
	proc := startRegister(t, pool)

	csr, err := ca.ParseCSR(clientCSRPEM(t, "Any Name", "Any Brand"))
	require.NoError(t, err)
	cert, err := intermediateAuthority(t).Issue(csr, ca.Client, time.Now())
	require.NoError(t, err)

	handshake := func(p process) error {
		_, cfg := listener(t, p, securePort)
		return cfg.VerifyConnection(tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}})
	}
	require.NoError(t, handshake(proc))

	before := getCRL(t, proc).Number
	chain := ca.EncodeCertificates(cert, intermediateAuthority(t).Certificate)
	status, p := revokeRequest(t, proc, chain, "affiliationChanged")
	require.Equal(t, http.StatusNoContent, status, p.Detail)

	require.Error(t, handshake(proc), "the next handshake fails")

	crl := getCRL(t, proc)
	assert.Equal(t, 1, crl.Number.Cmp(before), "the revocation signs a new CRL")
	require.Len(t, crl.RevokedCertificateEntries, 1)
	assert.Equal(t, cert.SerialNumber, crl.RevokedCertificateEntries[0].SerialNumber)
	assert.Equal(t, int(ca.AffiliationChanged), crl.RevokedCertificateEntries[0].ReasonCode)

	status, p = revokeRequest(t, proc, ca.EncodeCertificates(cert), "keyCompromise")
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, p.Detail, "affiliationChanged")
	assert.Equal(t, crl.Number, getCRL(t, proc).Number, "a refused revocation signs no CRL")

	restarted := startRegister(t, pool)
	require.Error(t, handshake(restarted))
	assert.Len(t, getCRL(t, restarted).RevokedCertificateEntries, 1)
}

// Verifies: REG-0705
func TestTheRegisterRevokesOnlyACertificateOfTheIntermediate(t *testing.T) {
	proc := startRegister(t, databasetest.New(t))

	status, p := revokeRequest(t, proc, clientCertificatePEM(t), "keyCompromise")

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, p.Detail, "certificate")
	assert.Empty(t, getCRL(t, proc).RevokedCertificateEntries)
}
