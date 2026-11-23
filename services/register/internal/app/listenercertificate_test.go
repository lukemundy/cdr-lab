package app_test

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

var registerNames = []string{"register", "mtls.register.example"}

type listener struct {
	current    *tls.Certificate
	replaced   int
	replaceErr error
}

func (l *listener) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return l.current, nil
}

func (l *listener) Replace(_ context.Context, cert tls.Certificate) error {
	if l.replaceErr != nil {
		return l.replaceErr
	}

	l.current = &cert
	l.replaced++

	return nil
}

func listenerCertificates(c *clock) app.ListenerCertificates {
	return app.NewListenerCertificates(testLab().CDRIntermediate, registerNames, c.Now)
}

func issuedListener(t *testing.T, certs app.ListenerCertificates) *listener {
	t.Helper()

	cert, err := certs.Issue()
	require.NoError(t, err)

	return &listener{current: &cert}
}

// Verifies: REG-1515, REG-1517, REG-1518
func TestTheListenerCertificateIsACDRServerCertificateOfTheRegister(t *testing.T) {
	cert, err := listenerCertificates(&clock{now: issueTime}).Issue()
	require.NoError(t, err)

	leaf := cert.Leaf
	require.NotNil(t, leaf)
	assert.Equal(t, []string{"Mock CDR Register"}, leaf.Subject.Organization)
	assert.Equal(t, []string{"Consumer Data Right"}, leaf.Subject.OrganizationalUnit)
	assert.Equal(t, []string{"AU"}, leaf.Subject.Country)
	assert.Equal(t, "register", leaf.Subject.CommonName)
	assert.Equal(t, registerNames, leaf.DNSNames)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, leaf.ExtKeyUsage)
	assert.Equal(t, []string{"http://ca.example/crl"}, leaf.CRLDistributionPoints)
	assert.Equal(t, issueTime.Truncate(time.Second), leaf.NotBefore)
	require.NoError(t, leaf.CheckSignatureFrom(testLab().CDRIntermediate.Certificate))

	key, ok := cert.PrivateKey.(*rsa.PrivateKey)
	require.True(t, ok)
	assert.Equal(t, ca.LeafKeyBits, key.N.BitLen())
	assert.True(t, key.PublicKey.Equal(leaf.PublicKey))
}

// Verifies: REG-1516
func TestEachListenerCertificateHasANewKey(t *testing.T) {
	certs := listenerCertificates(&clock{now: issueTime})

	first, err := certs.Issue()
	require.NoError(t, err)
	second, err := certs.Issue()
	require.NoError(t, err)

	firstKey, ok := first.PrivateKey.(*rsa.PrivateKey)
	require.True(t, ok)
	assert.False(t, firstKey.Equal(second.PrivateKey))
}

// Verifies: REG-1522
func TestTheListenerCertificateIsNotIssuedForNamesThatAreNotValid(t *testing.T) {
	for name, names := range map[string][]string{
		"no names":            nil,
		"not a host name":     {"register_1"},
		"an empty second one": {"register", ""},
	} {
		t.Run(name, func(t *testing.T) {
			certs := app.NewListenerCertificates(testLab().CDRIntermediate, names, (&clock{now: issueTime}).Now)

			_, err := certs.Issue()

			assert.Error(t, err)
		})
	}
}

// Verifies: REG-1520, REG-1526
func TestTheListenerGetsANewCertificateWhen30DaysOrLessRemain(t *testing.T) {
	c := &clock{now: issueTime}
	certs := listenerCertificates(c)
	l := issuedListener(t, certs)
	first := l.current.Leaf

	c.now = first.NotAfter.Add(-30*24*time.Hour - time.Second)

	require.NoError(t, certs.Renew(t.Context(), l))
	assert.Same(t, first, l.current.Leaf, "more than 30 days remain")

	c.now = first.NotAfter.Add(-30 * 24 * time.Hour)

	require.NoError(t, certs.Renew(t.Context(), l))
	require.Equal(t, 1, l.replaced)
	assert.NotEqual(t, first.SerialNumber, l.current.Leaf.SerialNumber)
	assert.Equal(t, c.now.Truncate(time.Second), l.current.Leaf.NotBefore)
	assert.True(t, l.current.Leaf.NotAfter.After(first.NotAfter))
}

// Verifies: REG-1522, REG-1525
func TestAFailedRenewalKeepsTheCurrentCertificateAndLogsAnError(t *testing.T) {
	logs := captureLogs(t)
	c := &clock{now: issueTime}
	certs := listenerCertificates(c)
	l := issuedListener(t, certs)
	first := l.current
	l.replaceErr = errors.New("replace failed")
	c.now = first.Leaf.NotAfter

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})

	go func() {
		defer close(done)

		certs.KeepRenewed(ctx, l, time.Millisecond)
	}()

	assert.Eventually(t, func() bool { return strings.Contains(logs.String(), `"level":"ERROR"`) },
		5*time.Second, time.Millisecond)
	cancel()
	<-done

	assert.Same(t, first, l.current)
	assert.Contains(t, logs.String(), "the certificate of the mTLS listener was not renewed")
}

// Verifies: REG-1516
func TestTheKeyOfTheListenerCertificateIsNotInTheLog(t *testing.T) {
	logs := captureLogs(t)
	c := &clock{now: issueTime}
	certs := listenerCertificates(c)
	l := issuedListener(t, certs)
	c.now = l.current.Leaf.NotAfter

	require.NoError(t, certs.Renew(t.Context(), l))

	key, err := x509.MarshalPKCS8PrivateKey(l.current.PrivateKey)
	require.NoError(t, err)

	rsaKey, ok := l.current.PrivateKey.(*rsa.PrivateKey)
	require.True(t, ok)

	assert.NotContains(t, logs.String(), base64.StdEncoding.EncodeToString(key))
	assert.NotContains(t, logs.String(), rsaKey.D.String())
}

// logBuffer is safe for a test that reads the log while a goroutine writes it.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *logBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *logBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func captureLogs(t *testing.T) *logBuffer {
	t.Helper()

	logs := &logBuffer{}

	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return logs
}
