package app

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// CertificateIssuer is the mock CDR intermediate CA of a running lab.
type CertificateIssuer struct {
	store     eventsourcing.Store
	authority ca.Authority
	now       func() time.Time
}

func NewCertificateIssuer(store eventsourcing.Store, authority ca.Authority, now func() time.Time) CertificateIssuer {
	return CertificateIssuer{store: store, authority: authority, now: now}
}

// IssueClient gives the certificate and then the intermediate, or no certificate when the record changes during the
// request.
func (i CertificateIssuer) IssueClient(
	ctx context.Context,
	legalEntityID, softwareProductID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	return i.issueClient(ctx, legalEntityID, nil, softwareProductID, csr)
}

// IssueClientAt signs nothing and gives eventsourcing.ErrVersionConflict when the data recipient has a different
// version.
func (i CertificateIssuer) IssueClientAt(
	ctx context.Context,
	legalEntityID string,
	version int64,
	softwareProductID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	return i.issueClient(ctx, legalEntityID, &version, softwareProductID, csr)
}

func (i CertificateIssuer) issueClient(
	ctx context.Context,
	legalEntityID string,
	version *int64,
	softwareProductID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	return issue(ctx, i, legalEntityID, version, csr, certificateRecord[
		domain.DataRecipientState, domain.DataRecipientCommand, domain.DataRecipientEvent,
	]{
		aggregate: DataRecipientType,
		kind:      ca.Client,
		subjectRules: func(state domain.DataRecipientState, subject pkix.Name) ([]string, error) {
			return domain.ClientCertificateSubjectRules(state, softwareProductID, subject)
		},
		command: func(thumbprint string) domain.DataRecipientCommand {
			return domain.IssueClientCertificate{SoftwareProductID: softwareProductID, Thumbprint: thumbprint}
		},
	})
}

// IssueServer gives the certificate and then the intermediate, or no certificate when the record changes during the
// request.
func (i CertificateIssuer) IssueServer(
	ctx context.Context,
	legalEntityID, brandID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	return i.issueServer(ctx, legalEntityID, nil, brandID, csr)
}

// IssueServerAt signs nothing and gives eventsourcing.ErrVersionConflict when the data holder has a different version.
func (i CertificateIssuer) IssueServerAt(
	ctx context.Context,
	legalEntityID string,
	version int64,
	brandID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	return i.issueServer(ctx, legalEntityID, &version, brandID, csr)
}

func (i CertificateIssuer) issueServer(
	ctx context.Context,
	legalEntityID string,
	version *int64,
	brandID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	return issue(ctx, i, legalEntityID, version, csr, certificateRecord[
		domain.DataHolderState, domain.DataHolderCommand, domain.DataHolderEvent,
	]{
		aggregate: DataHolderType,
		kind:      ca.Server,
		subjectRules: func(state domain.DataHolderState, subject pkix.Name) ([]string, error) {
			return domain.ServerCertificateSubjectRules(state, brandID, subject)
		},
		command: func(thumbprint string) domain.DataHolderCommand {
			return domain.IssueServerCertificate{BrandID: brandID, Thumbprint: thumbprint}
		},
	})
}

type certificateRecord[S, C, E any] struct {
	aggregate    eventsourcing.AggregateType[S, C, E]
	kind         ca.Kind
	subjectRules func(state S, subject pkix.Name) ([]string, error)
	command      func(thumbprint string) C
}

func issue[S, C, E any](
	ctx context.Context,
	i CertificateIssuer,
	legalEntityID string,
	expectedVersion *int64,
	csr *x509.CertificateRequest,
	r certificateRecord[S, C, E],
) ([]*x509.Certificate, error) {
	state, version, err := r.aggregate.Load(ctx, i.store, legalEntityID)
	if err != nil {
		return nil, err
	}

	if expectedVersion != nil && version != 0 && version != *expectedVersion {
		return nil, eventsourcing.ErrVersionConflict
	}

	recordRules, err := r.subjectRules(state, csr.Subject)
	if err != nil {
		return nil, err
	}

	if rules := append(profileRules(csr, r.kind), recordRules...); len(rules) > 0 {
		return nil, domain.CSRRulesError{Rules: rules}
	}

	// X.509 times have no fractional seconds.
	cert, err := i.authority.Issue(csr, r.kind, i.now().UTC().Truncate(time.Second))
	if err != nil {
		return nil, err
	}

	command := r.command(domain.CertificateThumbprint(cert))
	if err := r.aggregate.ExecuteAt(ctx, i.store, legalEntityID, version, command); err != nil {
		return nil, err
	}

	return []*x509.Certificate{cert, i.authority.Certificate}, nil
}

func profileRules(csr *x509.CertificateRequest, kind ca.Kind) []string {
	err := ca.ValidateCSR(csr, kind)
	if err == nil {
		return nil
	}

	errs := []error{err}

	var joined interface{ Unwrap() []error }
	if errors.As(err, &joined) {
		errs = joined.Unwrap()
	}

	rules := make([]string, 0, len(errs))
	for _, e := range errs {
		rules = append(rules, e.Error())
	}

	return rules
}
