package revocation_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds/revocation"
)

type authority struct {
	t    *testing.T
	cert *x509.Certificate
	key  *rsa.PrivateKey
	// crl is the response body of the CRL server. Nil gives 503.
	crl      atomic.Pointer[[]byte]
	requests atomic.Int64
	url      string
}

func newAuthority(t *testing.T) *authority {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Test Intermediate"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	a := &authority{t: t, cert: cert, key: key}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		a.requests.Add(1)

		body := a.crl.Load()
		if body == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		_, _ = w.Write(*body)
	}))
	t.Cleanup(srv.Close)
	a.url = srv.URL + "/ca.crl"

	return a
}

func (a *authority) issue(serial int64, crlURLs ...string) *x509.Certificate {
	a.t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(a.t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "client"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		CRLDistributionPoints: crlURLs,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, a.cert, &key.PublicKey, a.key)
	require.NoError(a.t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(a.t, err)

	return cert
}

func (a *authority) signCRL(signer *authority, nextUpdate time.Time, revoked ...int64) []byte {
	a.t.Helper()

	tmpl := &x509.RevocationList{Number: big.NewInt(1), ThisUpdate: time.Now().Add(-time.Hour), NextUpdate: nextUpdate}
	for _, serial := range revoked {
		tmpl.RevokedCertificateEntries = append(tmpl.RevokedCertificateEntries,
			x509.RevocationListEntry{SerialNumber: big.NewInt(serial), RevocationTime: time.Now().Add(-time.Minute)})
	}

	der, err := x509.CreateRevocationList(rand.Reader, tmpl, signer.cert, signer.key)
	require.NoError(a.t, err)

	return der
}

func (a *authority) publish(revoked ...int64) {
	a.t.Helper()
	der := a.signCRL(a, time.Now().Add(time.Hour), revoked...)
	a.crl.Store(&der)
}

func (a *authority) checker() *revocation.Checker {
	return revocation.NewChecker([]*x509.Certificate{a.cert})
}

func TestACertificateThatIsNotOnTheCRLIsAccepted(t *testing.T) {
	a := newAuthority(t)
	a.publish(99)

	err := a.checker().Check(t.Context(), a.issue(2, a.url))

	assert.NoError(t, err)
}

func TestACertificateThatIsOnTheCRLIsRevoked(t *testing.T) {
	a := newAuthority(t)
	a.publish(2)

	err := a.checker().Check(t.Context(), a.issue(2, a.url))

	assert.ErrorIs(t, err, revocation.ErrRevoked)
}

func TestTheCheckFailsWhenTheRevocationStatusIsNotKnown(t *testing.T) {
	other := newAuthority(t)

	tests := map[string]func(a *authority) *x509.Certificate{
		"the certificate gives no CRL URL": func(a *authority) *x509.Certificate {
			a.publish()
			return a.issue(2)
		},
		"the CRL URL is not HTTP": func(a *authority) *x509.Certificate {
			a.publish()
			return a.issue(2, "ldap://directory.example/ca.crl")
		},
		"the CRL server gives an error": func(a *authority) *x509.Certificate {
			return a.issue(2, a.url)
		},
		"the response is not a CRL": func(a *authority) *x509.Certificate {
			body := []byte("not a CRL")
			a.crl.Store(&body)

			return a.issue(2, a.url)
		},
		"a different CA signed the CRL": func(a *authority) *x509.Certificate {
			der := a.signCRL(other, time.Now().Add(time.Hour))
			a.crl.Store(&der)

			return a.issue(2, a.url)
		},
		"the CRL is after its next update": func(a *authority) *x509.Certificate {
			der := a.signCRL(a, time.Now().Add(-time.Minute))
			a.crl.Store(&der)

			return a.issue(2, a.url)
		},
		"the issuer is not a trusted issuer": func(a *authority) *x509.Certificate {
			other.publish()
			return other.issue(2, other.url)
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			a := newAuthority(t)
			cert := setup(a)

			err := a.checker().Check(t.Context(), cert)

			require.ErrorIs(t, err, revocation.ErrStatusUnknown)
			assert.NotErrorIs(t, err, revocation.ErrRevoked)
		})
	}
}

func TestTheCRLIsCachedForAShortTime(t *testing.T) {
	a := newAuthority(t)
	a.publish()
	checker := a.checker()
	now := time.Now()
	checker.Now = func() time.Time { return now }
	cert := a.issue(2, a.url)

	require.NoError(t, checker.Check(t.Context(), cert))
	a.publish(2)
	require.NoError(t, checker.Check(t.Context(), cert), "the cached CRL is used")
	assert.EqualValues(t, 1, a.requests.Load())

	now = now.Add(revocation.DefaultMaxAge + time.Second)

	require.ErrorIs(
		t,
		checker.Check(t.Context(), cert),
		revocation.ErrRevoked,
		"a new CRL is loaded after the maximum age",
	)
	assert.EqualValues(t, 2, a.requests.Load())
}

// checkerAt gives a checker with a controlled clock that has loaded the current CRL.
func (a *authority) checkerAt(now *time.Time, cert *x509.Certificate) *revocation.Checker {
	a.t.Helper()

	checker := a.checker()
	checker.Now = func() time.Time { return *now }
	require.NoError(a.t, checker.Check(a.t.Context(), cert))

	return checker
}

func TestTheLastCRLIsUsedWhenAReloadFails(t *testing.T) {
	other := newAuthority(t)

	tests := map[string]func(a *authority){
		"the CRL server gives an error": func(a *authority) { a.crl.Store(nil) },
		"the response is not a CRL": func(a *authority) {
			body := []byte("not a CRL")
			a.crl.Store(&body)
		},
		"a different CA signed the CRL": func(a *authority) {
			der := a.signCRL(other, time.Now().Add(time.Hour), 2)
			a.crl.Store(&der)
		},
		"the CRL is after its next update": func(a *authority) {
			der := a.signCRL(a, time.Now().Add(-time.Minute), 2)
			a.crl.Store(&der)
		},
	}
	for name, fail := range tests {
		t.Run(name, func(t *testing.T) {
			a := newAuthority(t)
			a.publish(3)

			now := time.Now()
			valid, revoked := a.issue(2, a.url), a.issue(3, a.url)
			checker := a.checkerAt(&now, valid)

			fail(a)

			now = now.Add(revocation.RetryInterval + time.Second)

			require.NoError(t, checker.Check(t.Context(), valid))
			require.ErrorIs(t, checker.Check(t.Context(), revoked), revocation.ErrRevoked)
			assert.EqualValues(t, 2, a.requests.Load(), "the reload is tried")
		})
	}
}

func TestTheLastCRLIsNotUsedAfterItsNextUpdate(t *testing.T) {
	a := newAuthority(t)
	a.publish()

	now := time.Now()
	cert := a.issue(2, a.url)
	checker := a.checkerAt(&now, cert)

	a.crl.Store(nil)

	now = now.Add(time.Hour)

	assert.ErrorIs(t, checker.Check(t.Context(), cert), revocation.ErrStatusUnknown)
}

func TestAFailedReloadIsTriedAgainAfterTheRetryInterval(t *testing.T) {
	a := newAuthority(t)
	a.publish()

	now := time.Now()
	cert := a.issue(2, a.url)
	checker := a.checkerAt(&now, cert)

	a.crl.Store(nil)

	now = now.Add(revocation.RetryInterval + time.Second)

	require.NoError(t, checker.Check(t.Context(), cert))
	require.NoError(t, checker.Check(t.Context(), cert))
	assert.EqualValues(t, 2, a.requests.Load(), "one attempt in each retry interval")

	a.publish(2)

	now = now.Add(revocation.RetryInterval)

	require.ErrorIs(t, checker.Check(t.Context(), cert), revocation.ErrRevoked)
	assert.EqualValues(t, 3, a.requests.Load())
}

func TestTheFirstLoadIsTriedAgainAtEachCheckUntilItSucceeds(t *testing.T) {
	a := newAuthority(t)
	cert := a.issue(2, a.url)
	checker := a.checker()

	require.ErrorIs(t, checker.Check(t.Context(), cert), revocation.ErrStatusUnknown)

	a.publish()

	require.NoError(t, checker.Check(t.Context(), cert))
	assert.EqualValues(t, 2, a.requests.Load())
}

//nolint:paralleltest // the test changes the default logger
func TestAFailedReloadThatUsesTheLastCRLGivesAWarning(t *testing.T) {
	var logs bytes.Buffer

	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	a := newAuthority(t)
	a.publish()

	now := time.Now()
	cert := a.issue(2, a.url)
	checker := a.checkerAt(&now, cert)

	a.crl.Store(nil)

	now = now.Add(revocation.RetryInterval + time.Second)

	require.NoError(t, checker.Check(t.Context(), cert))

	var entry map[string]any
	require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, a.url, entry["url.full"])
	assert.Contains(t, entry["error"], "503")
}

func TestACheckerWithAMaximumAgeOfZeroLoadsTheCRLAtEachCheck(t *testing.T) {
	a := newAuthority(t)
	a.publish()
	checker := a.checker()
	checker.MaxAge = 0
	cert := a.issue(2, a.url)

	require.NoError(t, checker.Check(t.Context(), cert))
	a.publish(2)

	require.ErrorIs(t, checker.Check(t.Context(), cert), revocation.ErrRevoked)
	assert.EqualValues(t, 2, a.requests.Load())
}

func TestAFailedReloadIsTriedAgainAfterTheRetryIntervalAlsoWithAMaximumAgeOfZero(t *testing.T) {
	a := newAuthority(t)
	a.publish()

	now := time.Now()
	cert := a.issue(2, a.url)
	checker := a.checkerAt(&now, cert)
	checker.MaxAge = 0

	a.crl.Store(nil)

	require.NoError(t, checker.Check(t.Context(), cert))

	now = now.Add(time.Second)

	require.NoError(t, checker.Check(t.Context(), cert), "the last CRL is in use")
	assert.EqualValues(t, 2, a.requests.Load(), "one attempt in each retry interval")

	a.publish(2)

	now = now.Add(revocation.RetryInterval)

	require.ErrorIs(t, checker.Check(t.Context(), cert), revocation.ErrRevoked)
	assert.EqualValues(t, 3, a.requests.Load())
}
