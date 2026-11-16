package app_test

import (
	"context"
	"crypto/x509"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// caStore keeps the state of the CA in memory, as the database does across a restart.
type caStore struct {
	mu    sync.Mutex
	state app.CAState
}

func (s *caStore) ChangeCA(_ context.Context, change func(app.CAState) (app.CAChange, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, err := change(app.CAState{CRL: s.state.CRL, Revoked: append([]app.Revocation{}, s.state.Revoked...)})
	if err != nil {
		return err
	}

	if c.Revocation != nil {
		s.state.Revoked = append(s.state.Revoked, *c.Revocation)
	}

	if c.CRL != nil {
		s.state.CRL = *c.CRL
	}

	return nil
}

func (s *caStore) NewestCRL(context.Context) (app.CRL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state.CRL, nil
}

func (s *caStore) Revocation(_ context.Context, serial *big.Int) (app.Revocation, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.state.Revoked {
		if r.Serial.Cmp(serial) == 0 {
			return r, true, nil
		}
	}

	return app.Revocation{}, false, nil
}

// Revocations gives the revocations in the sequence of the store, which is enough for the tests of the CA.
func (s *caStore) Revocations(_ context.Context, query app.RevocationPageQuery) ([]app.Revocation, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := min((query.Page-1)*query.PageSize, len(s.state.Revoked))
	end := min(start+query.PageSize, len(s.state.Revoked))

	return append([]app.Revocation{}, s.state.Revoked[start:end]...), len(s.state.Revoked), nil
}

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func newCA(store *caStore, c *clock) app.CertificateAuthority {
	return app.NewCertificateAuthority(store, testLab().CDRIntermediate, c.Now)
}

func parsedCRL(t *testing.T, crl app.CRL) *x509.RevocationList {
	t.Helper()

	list, err := x509.ParseRevocationList(crl.DER)
	require.NoError(t, err)
	require.NoError(t, list.CheckSignatureFrom(testLab().CDRIntermediate.Certificate))

	return list
}

func issuedClient(t *testing.T) *x509.Certificate {
	t.Helper()

	cert, err := testLab().CDRIntermediate.Issue(clientCSR(t, nil), ca.Client, issueTime)
	require.NoError(t, err)

	return cert
}

func TestTheCASignsACRLThatIsValidForSevenDays(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}

	require.NoError(t, newCA(store, c).SignCRL(t.Context()))

	crl, err := newCA(store, c).CurrentCRL(t.Context())
	require.NoError(t, err)

	list := parsedCRL(t, crl)
	second := issueTime.Truncate(time.Second)

	assert.Equal(t, int64(1), crl.Number)
	assert.Equal(t, big.NewInt(1), list.Number)
	assert.Equal(t, second, crl.ThisUpdate)
	assert.Equal(t, second.AddDate(0, 0, 7), crl.NextUpdate)
	assert.Equal(t, crl.NextUpdate, list.NextUpdate.UTC())
	assert.Empty(t, list.RevokedCertificateEntries)
}

// Verifies: REG-0003
func TestTheCRLNumbersIncreaseAcrossARestart(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}

	for want := int64(1); want <= 3; want++ {
		// Each start of the Register makes a new CA on the same store, and signs a CRL.
		require.NoError(t, newCA(store, c).SignCRL(t.Context()))

		crl, err := newCA(store, c).CurrentCRL(t.Context())
		require.NoError(t, err)
		assert.Equal(t, want, crl.Number)
		assert.Equal(t, big.NewInt(want), parsedCRL(t, crl).Number)
	}
}

// Verifies: REG-0712
func TestARenewalSignsANewCRLNoLaterThan24HoursAfterTheCurrentOne(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	require.NoError(t, authority.SignCRL(t.Context()))

	renewAt := func(after time.Duration) int64 {
		c.now = issueTime.Add(after)

		require.NoError(t, authority.RenewCRL(t.Context()))

		crl, err := authority.CurrentCRL(t.Context())
		require.NoError(t, err)

		return crl.Number
	}

	assert.Equal(t, int64(1), renewAt(time.Hour), "a CRL of 1 hour is not renewed")
	assert.Equal(t, int64(1), renewAt(app.CRLRenewalAge-time.Second))
	assert.Equal(t, int64(2), renewAt(app.CRLRenewalAge), "also with no revocation")
	assert.Equal(t, int64(2), renewAt(app.CRLRenewalAge+time.Minute), "the new CRL is not renewed at once")
	assert.Less(t, app.CRLRenewalAge, 24*time.Hour)
}

// Verifies: REG-0712
func TestARenewalAfterAStopOfMoreThanSevenDaysSignsANewCRL(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	require.NoError(t, authority.SignCRL(t.Context()))

	c.now = issueTime.AddDate(0, 0, 8)
	_, err := authority.CurrentCRL(t.Context())
	require.ErrorIs(t, err, app.ErrNoCurrentCRL)

	require.NoError(t, authority.RenewCRL(t.Context()))

	crl, err := authority.CurrentCRL(t.Context())
	require.NoError(t, err)
	assert.Equal(t, int64(2), crl.Number)
}

func TestARenewalOfACAWithNoCRLSignsOne(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}

	require.NoError(t, newCA(store, c).RenewCRL(t.Context()))

	crl, err := newCA(store, c).CurrentCRL(t.Context())
	require.NoError(t, err)
	assert.Equal(t, int64(1), crl.Number)
}

func TestTheCAHasNoCurrentCRLBeforeItsFirstSignature(t *testing.T) {
	_, err := newCA(&caStore{}, &clock{now: issueTime}).CurrentCRL(t.Context())

	require.ErrorIs(t, err, app.ErrNoCurrentCRL)
}

func TestARevocationSignsANewCRLThatListsTheCertificateWithItsReason(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	require.NoError(t, authority.SignCRL(t.Context()))
	cert := issuedClient(t)

	c.now = issueTime.Add(time.Hour + 500*time.Millisecond)

	require.NoError(t, authority.Revoke(t.Context(), cert, ca.Superseded))

	crl, err := authority.CurrentCRL(t.Context())
	require.NoError(t, err)

	list := parsedCRL(t, crl)
	second := issueTime.Add(time.Hour).Truncate(time.Second)

	assert.Equal(t, int64(2), crl.Number)
	assert.Equal(t, second, crl.ThisUpdate)
	require.Len(t, list.RevokedCertificateEntries, 1)
	assert.Equal(t, cert.SerialNumber, list.RevokedCertificateEntries[0].SerialNumber)
	assert.Equal(t, second, list.RevokedCertificateEntries[0].RevocationTime.UTC())
	assert.Equal(t, int(ca.Superseded), list.RevokedCertificateEntries[0].ReasonCode)

	assert.Equal(t, []app.Revocation{{
		Serial:     cert.SerialNumber,
		Subject:    cert.Subject.String(),
		Thumbprint: domain.CertificateThumbprint(cert),
		Time:       second,
		Reason:     ca.Superseded,
	}}, store.state.Revoked)
}

// Verifies: REG-0003
func TestTheCRLKeepsEachRevocationAcrossARestart(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	first, second := issuedClient(t), issuedClient(t)
	require.NoError(t, newCA(store, c).Revoke(t.Context(), first, ca.KeyCompromise))
	require.NoError(t, newCA(store, c).Revoke(t.Context(), second, ca.Unspecified))

	require.NoError(t, newCA(store, c).SignCRL(t.Context()))

	crl, err := newCA(store, c).CurrentCRL(t.Context())
	require.NoError(t, err)

	entries := parsedCRL(t, crl).RevokedCertificateEntries
	require.Len(t, entries, 2)
	assert.Equal(t, first.SerialNumber, entries[0].SerialNumber)
	assert.Equal(t, second.SerialNumber, entries[1].SerialNumber)
	assert.Equal(t, int64(3), crl.Number)
}

// Verifies: REG-0706
func TestASecondRevocationOfACertificateChangesNothing(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	cert := issuedClient(t)
	require.NoError(t, authority.Revoke(t.Context(), cert, ca.KeyCompromise))

	before := store.state

	err := authority.Revoke(t.Context(), cert, ca.Superseded)

	var revoked app.AlreadyRevokedError
	require.ErrorAs(t, err, &revoked)
	assert.Equal(t, ca.KeyCompromise, revoked.Reason)
	require.EqualError(t, err, "the certificate is already revoked, with the reason keyCompromise")
	assert.Equal(t, before, store.state)
}

// Verifies: REG-0705
func TestTheCARevokesOnlyACertificateThatTheIntermediateSigned(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	lab := testLab()

	for name, cert := range map[string]*x509.Certificate{
		"the intermediate": lab.CDRIntermediate.Certificate,
		"the root":         lab.CDRRoot.Certificate,
		"a public certificate": func() *x509.Certificate {
			cert, err := lab.PublicRoot.IssuePublicServer(&csrKey().PublicKey, []string{"register"}, issueTime)
			require.NoError(t, err)

			return cert
		}(),
	} {
		err := newCA(store, c).Revoke(t.Context(), cert, ca.KeyCompromise)

		require.ErrorIs(t, err, app.ErrNotIssuedByTheIntermediate, name)
	}

	assert.Empty(t, store.state.Revoked)
	assert.Zero(t, store.state.CRL)
}

func TestAnErrorOfTheStoreIsTheErrorOfTheRevocation(t *testing.T) {
	failure := errors.New("database failure")

	err := app.NewCertificateAuthority(failingStore{failure}, testLab().CDRIntermediate, time.Now).
		Revoke(t.Context(), issuedClient(t), ca.KeyCompromise)

	require.ErrorIs(t, err, failure)
}

type failingStore struct{ err error }

func (s failingStore) ChangeCA(context.Context, func(app.CAState) (app.CAChange, error)) error {
	return s.err
}

func (s failingStore) NewestCRL(context.Context) (app.CRL, error) { return app.CRL{}, s.err }

func (s failingStore) Revocation(context.Context, *big.Int) (app.Revocation, bool, error) {
	return app.Revocation{}, false, s.err
}

func (s failingStore) Revocations(context.Context, app.RevocationPageQuery) ([]app.Revocation, int, error) {
	return nil, 0, s.err
}

// Verifies: REG-0728
func TestTheCheckRefusesARevokedCertificateAtOnce(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	require.NoError(t, authority.SignCRL(t.Context()))
	cert := issuedClient(t)

	require.NoError(t, authority.CheckRevocation(t.Context(), cert))
	require.NoError(t, authority.Revoke(t.Context(), cert, ca.KeyCompromise))

	require.ErrorIs(t, authority.CheckRevocation(t.Context(), cert), app.ErrCertificateRevoked)
	require.NoError(t, authority.CheckRevocation(t.Context(), issuedClient(t)), "only the revoked certificate")
}

// Verifies: REG-0729
func TestTheCheckRefusesEachCertificateWhenTheCAHasNoCurrentCRL(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	cert := issuedClient(t)

	require.ErrorIs(t, authority.CheckRevocation(t.Context(), cert), app.ErrNoCurrentCRL)

	require.NoError(t, authority.SignCRL(t.Context()))

	c.now = issueTime.AddDate(0, 0, 7)

	require.ErrorIs(t, authority.CheckRevocation(t.Context(), cert), app.ErrNoCurrentCRL)
}

// Verifies: REG-0727
func TestTheCheckRefusesACertificateThatTheIntermediateDidNotSign(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	require.NoError(t, authority.SignCRL(t.Context()))

	other, err := testLab().CDRRoot.Issue(clientCSR(t, nil), ca.Client, issueTime)
	require.NoError(t, err)

	require.ErrorIs(t, authority.CheckRevocation(t.Context(), other), app.ErrNotIssuedByTheIntermediate)
}

func TestACheckOfARevocationNamesTheCertificateAndChangesNothing(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	require.NoError(t, authority.SignCRL(t.Context()))
	cert := issuedClient(t)
	before := store.state

	got, err := authority.CheckRevocable(t.Context(), cert)

	require.NoError(t, err)
	assert.Equal(t, app.RevocableCertificate{
		Serial:     cert.SerialNumber,
		Subject:    cert.Subject.String(),
		Thumbprint: domain.CertificateThumbprint(cert),
		NotAfter:   cert.NotAfter.UTC(),
	}, got)
	assert.Equal(t, before, store.state)
}

func TestACheckOfARevocationRefusesARevokedCertificateWithItsReason(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	cert := issuedClient(t)
	require.NoError(t, authority.Revoke(t.Context(), cert, ca.AffiliationChanged))

	_, err := authority.CheckRevocable(t.Context(), cert)

	var revoked app.AlreadyRevokedError
	require.ErrorAs(t, err, &revoked)
	assert.Equal(t, ca.AffiliationChanged, revoked.Reason)
}

func TestACheckOfARevocationRefusesACertificateThatTheIntermediateDidNotSign(t *testing.T) {
	_, err := newCA(&caStore{}, &clock{now: issueTime}).CheckRevocable(t.Context(), testLab().CDRRoot.Certificate)

	require.ErrorIs(t, err, app.ErrNotIssuedByTheIntermediate)
}

func TestACheckOfARevocationGivesTheErrorOfTheStore(t *testing.T) {
	failure := errors.New("database failure")

	_, err := app.NewCertificateAuthority(failingStore{failure}, testLab().CDRIntermediate, time.Now).
		CheckRevocable(t.Context(), issuedClient(t))

	require.ErrorIs(t, err, failure)
}

func TestTheRevocationsAreThoseOfTheStore(t *testing.T) {
	store, c := &caStore{}, &clock{now: issueTime}
	authority := newCA(store, c)
	first, second := issuedClient(t), issuedClient(t)
	require.NoError(t, authority.Revoke(t.Context(), first, ca.KeyCompromise))
	require.NoError(t, authority.Revoke(t.Context(), second, ca.Superseded))

	page, total, err := authority.Revocations(t.Context(), app.RevocationPageQuery{Page: 2, PageSize: 1})

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, page, 1)
	assert.Equal(t, second.SerialNumber, page[0].Serial)
}
