package app

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// CRLRenewalAge is less than the 24 hours of the ACCC CP, section 4.9.7, by more than the interval of the renewal
// job, so that each CRL has a successor no later than 24 hours after its thisUpdate.
const CRLRenewalAge = 23 * time.Hour

var (
	ErrNotIssuedByTheIntermediate = errors.New("the mock CDR intermediate did not sign the certificate")
	ErrNoCurrentCRL               = errors.New("the Register has no current CRL of the mock CDR intermediate")
	ErrCertificateRevoked         = errors.New("the certificate is revoked")
)

type AlreadyRevokedError struct{ Reason ca.Reason }

func (e AlreadyRevokedError) Error() string {
	return "the certificate is already revoked, with the reason " + e.Reason.String()
}

type Revocation struct {
	Serial *big.Int
	// Subject is in the RFC 4514 format.
	Subject    string
	Thumbprint string
	Time       time.Time
	Reason     ca.Reason
}

// CRL is a CRL of the mock CDR intermediate, in DER. The zero CRL is no CRL.
type CRL struct {
	Number     int64
	ThisUpdate time.Time
	NextUpdate time.Time
	DER        []byte
}

type CAState struct {
	CRL     CRL
	Revoked []Revocation
}

// CAChange adds Revocation and replaces the CRL with CRL, when each is not nil.
type CAChange struct {
	Revocation *Revocation
	CRL        *CRL
}

type RevocationPageQuery struct {
	// Page starts at 1.
	Page     int
	PageSize int
}

type RevocableCertificate struct {
	Serial *big.Int
	// Subject is in the RFC 4514 format.
	Subject    string
	Thumbprint string
	NotAfter   time.Time
}

type CAStore interface {
	// ChangeCA calls change with the state of the CA, one change at a time, and stores the change that it gives in
	// one transaction. It stores nothing when change gives an error.
	ChangeCA(ctx context.Context, change func(CAState) (CAChange, error)) error
	NewestCRL(ctx context.Context) (CRL, error)
	Revocation(ctx context.Context, serial *big.Int) (r Revocation, found bool, err error)
	// Revocations gives the revocations newest first, then by serial number.
	Revocations(ctx context.Context, query RevocationPageQuery) (page []Revocation, total int, err error)
}

type CertificateAuthority struct {
	store     CAStore
	authority ca.Authority
	now       func() time.Time
}

func NewCertificateAuthority(store CAStore, authority ca.Authority, now func() time.Time) CertificateAuthority {
	return CertificateAuthority{store: store, authority: authority, now: now}
}

func (a CertificateAuthority) SignCRL(ctx context.Context) error {
	return a.store.ChangeCA(ctx, func(state CAState) (CAChange, error) {
		crl, err := a.sign(state.CRL.Number, state.Revoked)
		return CAChange{CRL: &crl}, err
	})
}

// RenewCRL signs a new CRL when the newest CRL is CRLRenewalAge or older, or when the CA has no CRL.
func (a CertificateAuthority) RenewCRL(ctx context.Context) error {
	return a.store.ChangeCA(ctx, func(state CAState) (CAChange, error) {
		if state.CRL.Number != 0 && a.now().Before(state.CRL.ThisUpdate.Add(CRLRenewalAge)) {
			return CAChange{}, nil
		}

		crl, err := a.sign(state.CRL.Number, state.Revoked)

		return CAChange{CRL: &crl}, err
	})
}

// Revoke also signs a new CRL that lists the certificate.
func (a CertificateAuthority) Revoke(ctx context.Context, cert *x509.Certificate, reason ca.Reason) error {
	if cert.CheckSignatureFrom(a.authority.Certificate) != nil {
		return ErrNotIssuedByTheIntermediate
	}

	return a.store.ChangeCA(ctx, func(state CAState) (CAChange, error) {
		for _, r := range state.Revoked {
			if r.Serial.Cmp(cert.SerialNumber) == 0 {
				return CAChange{}, AlreadyRevokedError{Reason: r.Reason}
			}
		}

		revocation := Revocation{
			Serial:     cert.SerialNumber,
			Subject:    cert.Subject.String(),
			Thumbprint: domain.CertificateThumbprint(cert),
			Time:       a.time(),
			Reason:     reason,
		}

		crl, err := a.sign(state.CRL.Number, append(state.Revoked, revocation))

		return CAChange{Revocation: &revocation, CRL: &crl}, err
	})
}

// CheckRevocable gives the errors of Revoke, and changes nothing.
func (a CertificateAuthority) CheckRevocable(
	ctx context.Context,
	cert *x509.Certificate,
) (RevocableCertificate, error) {
	if cert.CheckSignatureFrom(a.authority.Certificate) != nil {
		return RevocableCertificate{}, ErrNotIssuedByTheIntermediate
	}

	r, found, err := a.store.Revocation(ctx, cert.SerialNumber)
	if err != nil {
		return RevocableCertificate{}, err
	}

	if found {
		return RevocableCertificate{}, AlreadyRevokedError{Reason: r.Reason}
	}

	return RevocableCertificate{
		Serial:     cert.SerialNumber,
		Subject:    cert.Subject.String(),
		Thumbprint: domain.CertificateThumbprint(cert),
		NotAfter:   cert.NotAfter.UTC(),
	}, nil
}

func (a CertificateAuthority) Revocations(ctx context.Context, query RevocationPageQuery) ([]Revocation, int, error) {
	return a.store.Revocations(ctx, query)
}

func (a CertificateAuthority) CurrentCRL(ctx context.Context) (CRL, error) {
	crl, err := a.store.NewestCRL(ctx)
	if err != nil {
		return CRL{}, err
	}

	if crl.Number == 0 || !a.now().Before(crl.NextUpdate) {
		return CRL{}, ErrNoCurrentCRL
	}

	return crl, nil
}

// CheckRevocation uses the newest CRL, so that a revocation takes effect at once.
func (a CertificateAuthority) CheckRevocation(ctx context.Context, cert *x509.Certificate) error {
	if cert.CheckSignatureFrom(a.authority.Certificate) != nil {
		return ErrNotIssuedByTheIntermediate
	}

	crl, err := a.CurrentCRL(ctx)
	if err != nil {
		return err
	}

	list, err := x509.ParseRevocationList(crl.DER)
	if err != nil {
		return fmt.Errorf("parse the CRL: %w", err)
	}

	for _, entry := range list.RevokedCertificateEntries {
		if entry.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			return ErrCertificateRevoked
		}
	}

	return nil
}

func (a CertificateAuthority) sign(lastNumber int64, revoked []Revocation) (CRL, error) {
	entries := make([]ca.Revocation, 0, len(revoked))
	for _, r := range revoked {
		entries = append(entries, ca.Revocation{Serial: r.Serial, Time: r.Time, Reason: r.Reason})
	}

	// A number from the time is higher than each number of an earlier lab, which has the same issuer DN.
	thisUpdate := a.time()
	number := max(lastNumber+1, thisUpdate.Unix())

	der, err := a.authority.CRL(entries, big.NewInt(number), thisUpdate)
	if err != nil {
		return CRL{}, err
	}

	list, err := x509.ParseRevocationList(der)
	if err != nil {
		return CRL{}, err
	}

	return CRL{Number: number, ThisUpdate: list.ThisUpdate.UTC(), NextUpdate: list.NextUpdate.UTC(), DER: der}, nil
}

// time has no fractional seconds, because X.509 times do not keep them.
func (a CertificateAuthority) time() time.Time {
	return a.now().UTC().Truncate(time.Second)
}
