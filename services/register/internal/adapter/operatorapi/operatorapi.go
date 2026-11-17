// Package operatorapi serves the operator interface of the Register.
package operatorapi

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type Operator interface {
	Execute(ctx context.Context, legalEntityID string, command domain.DataRecipientCommand) error
	ExecuteDataHolder(ctx context.Context, legalEntityID string, command domain.DataHolderCommand) error
}

// CertificateIssuer gives the certificate and then the certificate of its issuer.
type CertificateIssuer interface {
	IssueClient(
		ctx context.Context,
		legalEntityID, softwareProductID string,
		csr *x509.CertificateRequest,
	) ([]*x509.Certificate, error)
	IssueServer(
		ctx context.Context,
		legalEntityID, brandID string,
		csr *x509.CertificateRequest,
	) ([]*x509.Certificate, error)
}

type issueFunc func(ctx context.Context, legalEntityID, recordID string, csr *x509.CertificateRequest) (
	[]*x509.Certificate, error,
)

func Register(
	mux *http.ServeMux,
	operator Operator,
	issuer CertificateIssuer,
	revoker Revoker,
	caBundle []*x509.Certificate,
	keys app.KeyChanger,
	access operatorhttp.Config,
) {
	routes := httproute.New(mux)
	// The checks apply to each route, not to the mux, so that the telemetry of a refused request has its route.
	routes.Use(operatorhttp.Guard(access))

	routes.Handle("POST /v1/key-sets/{keySet}/publish-next", changeKey(keys.PublishNextKey))
	routes.Handle("POST /v1/key-sets/{keySet}/activate-next", changeKey(keys.ActivateNextKey))

	for _, a := range domain.DataRecipientActions {
		routes.Handle(
			"POST /v1/data-recipients/{legalEntityId}/"+string(a),
			execute(operator, func(*http.Request) domain.DataRecipientCommand { return a.Command() }),
		)
	}

	for _, a := range domain.RecordActions {
		routes.Handle(
			"POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/"+string(a),
			execute(operator, func(r *http.Request) domain.DataRecipientCommand {
				return a.SoftwareProductCommand(r.PathValue("softwareProductId"))
			}),
		)
	}

	routes.Handle(
		"POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates",
		registerCertificate(operator),
	)
	routes.Handle(
		"POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue",
		issueCertificate(issuer.IssueClient, "softwareProductId", "data recipient"),
	)
	routes.Handle(
		"POST /v1/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue",
		issueCertificate(issuer.IssueServer, "brandId", "data holder"),
	)
	routes.Handle("GET /v1/cdr-ca-bundle", certificateChain(ca.EncodeCertificates(caBundle...)))
	routes.Handle("POST /v1/cdr-certificates/revoke", revokeCertificate(revoker))
	routes.Handle(
		"DELETE /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/{thumbprint}",
		execute(operator, func(r *http.Request) domain.DataRecipientCommand {
			return domain.RemoveClientCertificate{
				SoftwareProductID: r.PathValue("softwareProductId"),
				Thumbprint:        r.PathValue("thumbprint"),
			}
		}),
	)
	registerDataRecipientChanges(routes, operator)
	registerDataHolders(routes, operator)
	routes.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		problem.Write(w, http.StatusNotFound, "The URL and method are not an operator action.")
	}))
}

func changeKey(step func(context.Context, app.KeySet) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		set, ok := app.ParseKeySet(r.PathValue("keySet"))
		if !ok {
			problem.Write(
				w,
				http.StatusNotFound,
				"The key set does not exist. The key sets are access-token and ssa.",
			)

			return
		}

		err := step(r.Context(), set)
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, app.ErrKeyChangeInProgress):
			problem.Write(
				w,
				http.StatusConflict,
				"The key set has a next key that is not active. Activate it before you publish a new key.",
			)
		case errors.Is(err, app.ErrNoNextKey):
			problem.Write(
				w,
				http.StatusConflict,
				"The key set has no next key. Publish a key before you activate it.",
			)
		default:
			problem.WriteUnexpected(w, r, err, notCompleted)
		}
	})
}

// maxCertificateBytes is more than the size of a PEM certificate with an RSA 4096 key and many extensions.
const (
	maxCertificateBytes = 16 << 10
	notCompleted        = "The Register could not complete the request."
)

func registerCertificate(operator Operator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		thumbprint, err := certificateThumbprint(http.MaxBytesReader(w, r.Body, maxCertificateBytes))
		if err != nil {
			problem.WriteBodyRefusal(w, err, "The request body must be one PEM certificate.")
			return
		}

		command := domain.RegisterClientCertificate{
			SoftwareProductID: r.PathValue("softwareProductId"),
			Thumbprint:        thumbprint,
		}
		executeCommand(w, r, operator, command)
	})
}

// maxCSRBytes is more than the size of a PEM CSR with an RSA 4096 key and a long subject.
const maxCSRBytes = 16 << 10

// recordID is the path value of the record, and participant names the aggregate in the detail of a version conflict.
func issueCertificate(issue issueFunc, recordID, participant string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxCSRBytes))

		var csr *x509.CertificateRequest
		if err == nil {
			csr, err = ca.ParseCSR(data)
		}

		if err != nil {
			problem.WriteBodyRefusal(w, err, "The request body must be one PEM CSR.")
			return
		}

		chain, err := issue(r.Context(), r.PathValue("legalEntityId"), r.PathValue(recordID), csr)
		writeResult(w, r, err, participant, func(w http.ResponseWriter) {
			writeCertificateChain(w, ca.EncodeCertificates(chain...))
		})
	})
}

func certificateChain(body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeCertificateChain(w, body) })
}

func writeCertificateChain(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/pem-certificate-chain")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(body)
}

func certificateThumbprint(body io.Reader) (string, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	return domain.ClientCertificateThumbprint(data)
}

func execute(operator Operator, command func(*http.Request) domain.DataRecipientCommand) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executeCommand(w, r, operator, command(r))
	})
}

func executeCommand(
	w http.ResponseWriter,
	r *http.Request,
	operator Operator,
	command domain.DataRecipientCommand,
) {
	err := operator.Execute(r.Context(), r.PathValue("legalEntityId"), command)
	writeResult(w, r, err, "data recipient", noContent)
}

func noContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// writeResult gives the response of a command. participant names the aggregate in the detail of a version conflict.
func writeResult(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	participant string,
	success func(http.ResponseWriter),
) {
	var (
		ruleErr   domain.StatusRuleError
		notUnique domain.IDNotUniqueError
	)

	switch {
	case err == nil:
		success(w)
	case errors.As(err, &ruleErr):
		problem.Write(w, http.StatusConflict, ruleErr.Error())
	case errors.As(err, &notUnique):
		problem.Write(w, http.StatusConflict, notUnique.Error())
	case errors.Is(err, domain.ErrExists):
		problem.Write(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		problem.Write(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrNotValid):
		problem.Write(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, eventsourcing.ErrVersionConflict):
		problem.Write(
			w,
			http.StatusConflict,
			fmt.Sprintf("The %s changed during the request. Send the request again.", participant),
		)
	default:
		problem.WriteUnexpected(w, r, err, notCompleted)
	}
}
