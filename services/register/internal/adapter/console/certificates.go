package console

import (
	"io"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// maxRegistrationBodyBytes is more than twice the size of a PEM certificate with an RSA 4096 key, because JSON escapes
// each line break.
const (
	maxRegistrationBodyBytes = 32 << 10
	notOnePEM                = "The certificate must be one PEM certificate, with no other text."
	badRegistrationBody      = "The request body must be a JSON object with only the version of the data recipient, " +
		"an integer of 1 or more, and the PEM certificate as a string."
	thumbprintNotValid = "The thumbprint must be a base64url SHA-256 value (RFC 8705 x5t#S256)."
)

func registerCertificateActions(routes *httproute.Routes, operator Operator) {
	const product = "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}"

	routes.Handle("POST "+product+"/certificates", registerCertificate(operator))
	routes.Handle("POST "+product+"/certificates/{thumbprint}/remove",
		executeAction(dataRecipientsAt(operator), noSoftwareProduct, func(r *http.Request) domain.DataRecipientCommand {
			return domain.RemoveClientCertificate{
				SoftwareProductID: r.PathValue("softwareProductId"),
				Thumbprint:        r.PathValue("thumbprint"),
			}
		}))
}

func registerCertificate(operator Operator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version, certificate, ok := decodeRegistration(http.MaxBytesReader(w, r.Body, maxRegistrationBodyBytes))
		if !ok {
			problem.Write(w, http.StatusBadRequest, badRegistrationBody)
			return
		}

		thumbprint, err := domain.ClientCertificateThumbprint([]byte(certificate))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, notOnePEM)
			return
		}

		command := domain.RegisterClientCertificate{
			SoftwareProductID: r.PathValue("softwareProductId"),
			Thumbprint:        thumbprint,
		}
		writeResult(w, r, noSoftwareProduct,
			operator.ExecuteAt(r.Context(), r.PathValue("legalEntityId"), version, command))
	})
}

func decodeRegistration(body io.Reader) (version int64, certificate string, ok bool) {
	var request struct {
		Version     *int64  `json:"version"`
		Certificate *string `json:"certificate"`
	}

	if jsonhttp.Decode(body, &request) != nil || request.Version == nil || *request.Version < 1 ||
		request.Certificate == nil {
		return 0, "", false
	}

	return *request.Version, *request.Certificate, true
}
