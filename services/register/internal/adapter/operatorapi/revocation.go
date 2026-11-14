package operatorapi

import (
	"context"
	"crypto/x509"
	"errors"
	"net/http"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

type Revoker interface {
	Revoke(ctx context.Context, cert *x509.Certificate, reason ca.Reason) error
}

type revocationRequest struct {
	Certificate *string `json:"certificate"`
	Reason      *string `json:"reason"`
}

func revokeCertificate(revoker Revoker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body revocationRequest
		if !decodeBody(w, r, &body) {
			return
		}

		cert, reason, errs := app.ParseRevocationRequest(body.Certificate, body.Reason)
		if len(errs) > 0 {
			problems := make([]string, 0, len(errs))
			for _, e := range errs {
				problems = append(problems, e.Field+": "+e.Detail)
			}

			problem.Write(w, http.StatusBadRequest, "The request body is not valid. "+strings.Join(problems, " "))

			return
		}

		var revoked app.AlreadyRevokedError

		switch err := revoker.Revoke(r.Context(), cert, reason); {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, app.ErrNotIssuedByTheIntermediate):
			problem.Write(
				w,
				http.StatusBadRequest,
				"certificate: the mock CDR intermediate did not sign the certificate.",
			)
		case errors.As(err, &revoked):
			problem.Write(w, http.StatusConflict, "The certificate is already revoked, with the reason "+
				revoked.Reason.String()+".")
		default:
			problem.WriteUnexpected(w, r, err, notCompleted)
		}
	})
}
