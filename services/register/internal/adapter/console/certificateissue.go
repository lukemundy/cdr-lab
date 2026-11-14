package console

import (
	"context"
	"crypto/x509"
	"errors"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// CertificateIssuer gives the certificate and then the certificate of its issuer.
type CertificateIssuer interface {
	IssueClientAt(
		ctx context.Context,
		legalEntityID string,
		version int64,
		softwareProductID string,
		csr *x509.CertificateRequest,
	) ([]*x509.Certificate, error)
	IssueServerAt(
		ctx context.Context,
		legalEntityID string,
		version int64,
		brandID string,
		csr *x509.CertificateRequest,
	) ([]*x509.Certificate, error)
}

type issueFunc func(
	ctx context.Context,
	legalEntityID string,
	version int64,
	recordID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error)

// maxIssueBodyBytes is twice the CSR limit of the operator interface, because JSON escapes each line break.
const (
	maxIssueBodyBytes = 32 << 10
	badIssueBody      = "The request body must be a JSON object with only the version of the participant, " +
		"an integer of 1 or more, and the PEM CSR as a string."
	notOneCSR = "The CSR must be one PEM CSR, with no other text."
)

type issueResponse struct {
	CertificateChain string `json:"certificateChain"`
	Thumbprint       string `json:"thumbprint"`
}

func registerCertificateIssues(routes *httproute.Routes, issuer CertificateIssuer) {
	// The handlers call the issuer through a function literal, because a method value of a nil issuer panics.
	routes.Handle("POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue",
		issueCertificate("softwareProductId", noSoftwareProduct, func(
			ctx context.Context, legalEntityID string, version int64, id string, csr *x509.CertificateRequest,
		) ([]*x509.Certificate, error) {
			return issuer.IssueClientAt(ctx, legalEntityID, version, id, csr)
		}))
	routes.Handle("POST /api/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue",
		issueCertificate("brandId", noDataHolderBrand, func(
			ctx context.Context, legalEntityID string, version int64, id string, csr *x509.CertificateRequest,
		) ([]*x509.Certificate, error) {
			return issuer.IssueServerAt(ctx, legalEntityID, version, id, csr)
		}))
}

func issueCertificate(recordID, notFound string, issue issueFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Version *int64  `json:"version"`
			CSR     *string `json:"csr"`
		}

		err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxIssueBodyBytes), &request)
		if err != nil || request.Version == nil || *request.Version < 1 || request.CSR == nil {
			problem.Write(w, http.StatusBadRequest, badIssueBody)
			return
		}

		csr, err := ca.ParseCSR([]byte(*request.CSR))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, notOneCSR)
			return
		}

		chain, err := issue(r.Context(), r.PathValue("legalEntityId"), *request.Version, r.PathValue(recordID), csr)
		writeIssueResult(w, r, notFound, chain, err)
	})
}

func writeIssueResult(w http.ResponseWriter, r *http.Request, notFound string, chain []*x509.Certificate, err error) {
	var (
		rulesErr domain.CSRRulesError
		ruleErr  domain.StatusRuleError
	)

	switch {
	case err == nil:
		jsonhttp.Write(w, http.StatusOK, issueResponse{
			CertificateChain: string(ca.EncodeCertificates(chain...)),
			Thumbprint:       domain.CertificateThumbprint(chain[0]),
		})
	case errors.As(err, &rulesErr):
		problem.Write(w, http.StatusBadRequest, rulesErr.Error())
	case errors.Is(err, eventsourcing.ErrVersionConflict):
		problem.Write(w, http.StatusConflict, staleVersion)
	case errors.As(err, &ruleErr):
		problem.Write(w, http.StatusConflict, ruleErr.Error())
	case errors.Is(err, domain.ErrNotFound):
		problem.Write(w, http.StatusNotFound, notFound)
	default:
		problem.WriteUnexpected(w, r, err, notCompleted)
	}
}
