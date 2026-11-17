package console

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

type CertificateAuthority interface {
	CheckRevocable(ctx context.Context, cert *x509.Certificate) (app.RevocableCertificate, error)
	Revoke(ctx context.Context, cert *x509.Certificate, reason ca.Reason) error
	Revocations(ctx context.Context, query app.RevocationPageQuery) ([]app.Revocation, int, error)
}

const badRevocationBody = "The request body must be a JSON object with only the PEM certificate and the reason, " +
	"as strings."

type revocableCertificate struct {
	SerialNumber string    `json:"serialNumber"`
	Subject      string    `json:"subject"`
	Thumbprint   string    `json:"thumbprint"`
	NotAfter     time.Time `json:"notAfter"`
}

type revocationPage struct {
	Data         []revocation `json:"data"`
	TotalRecords int          `json:"totalRecords"`
}

type revocation struct {
	SerialNumber string    `json:"serialNumber"`
	Subject      string    `json:"subject"`
	Thumbprint   string    `json:"thumbprint"`
	RevokedAt    time.Time `json:"revokedAt"`
	Reason       string    `json:"reason"`
}

func registerCDRCertificates(routes *httproute.Routes, authority CertificateAuthority, caBundle []*x509.Certificate) {
	bundle := struct {
		CABundle string `json:"caBundle"`
	}{CABundle: string(ca.EncodeCertificates(caBundle...))}

	routes.Handle("GET /api/cdr-ca-bundle", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonhttp.Write(w, http.StatusOK, bundle)
	}))
	routes.Handle("GET /api/cdr-certificates/revocations", revocations(authority))
	routes.Handle("POST /api/cdr-certificates/check-revocation", revocationAction(
		func(w http.ResponseWriter, r *http.Request, cert *x509.Certificate, _ ca.Reason) error {
			c, err := authority.CheckRevocable(r.Context(), cert)
			if err != nil {
				return err
			}

			jsonhttp.Write(w, http.StatusOK, revocableCertificate{
				SerialNumber: serialNumber(c.Serial),
				Subject:      c.Subject,
				Thumbprint:   c.Thumbprint,
				NotAfter:     c.NotAfter.UTC(),
			})

			return nil
		}))
	routes.Handle("POST /api/cdr-certificates/revoke", revocationAction(
		func(w http.ResponseWriter, r *http.Request, cert *x509.Certificate, reason ca.Reason) error {
			if err := authority.Revoke(r.Context(), cert, reason); err != nil {
				return err
			}

			w.WriteHeader(http.StatusNoContent)

			return nil
		}))
}

// serialNumber has the format of `openssl x509 -serial`.
func serialNumber(serial *big.Int) string {
	return fmt.Sprintf("%X", serial.Bytes())
}

// revocationAction writes each refusal, and the action writes its success.
func revocationAction(
	action func(w http.ResponseWriter, r *http.Request, cert *x509.Certificate, reason ca.Reason) error,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Certificate *string `json:"certificate"`
			Reason      *string `json:"reason"`
		}

		if err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxRegistrationBodyBytes), &request); err != nil {
			problem.WriteBodyRefusal(w, err, badRevocationBody)
			return
		}

		cert, reason, errs := app.ParseRevocationRequest(request.Certificate, request.Reason)
		if len(errs) > 0 {
			fields := make([]problem.FieldError, 0, len(errs))
			for _, e := range errs {
				fields = append(fields, problem.FieldError{Pointer: problem.FieldPointer(e.Field), Detail: e.Detail})
			}

			problem.WriteFieldErrors(w, http.StatusBadRequest, "", fields)

			return
		}

		err := action(w, r, cert, reason)
		if err == nil {
			return
		}

		var revoked app.AlreadyRevokedError

		switch {
		case errors.Is(err, app.ErrNotIssuedByTheIntermediate):
			problem.WriteFieldErrors(w, http.StatusBadRequest, "", []problem.FieldError{{
				Pointer: "#/certificate", Detail: "The mock CDR intermediate did not sign the certificate.",
			}})
		case errors.As(err, &revoked):
			problem.Write(w, http.StatusConflict, "The certificate is already revoked, with the reason "+
				revoked.Reason.String()+". The Register changed nothing.")
		default:
			problem.WriteUnexpected(w, r, err, notCompleted)
		}
	})
}

func revocations(authority CertificateAuthority) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, pageSize, err := consolehttp.ParsePage(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		items, total, err := authority.Revocations(r.Context(), app.RevocationPageQuery{Page: page, PageSize: pageSize})
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		body := revocationPage{Data: make([]revocation, 0, len(items)), TotalRecords: total}
		for _, item := range items {
			body.Data = append(body.Data, revocation{
				SerialNumber: serialNumber(item.Serial),
				Subject:      item.Subject,
				Thumbprint:   item.Thumbprint,
				RevokedAt:    item.Time.UTC(),
				Reason:       item.Reason.String(),
			})
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}
