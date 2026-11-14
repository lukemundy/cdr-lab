package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var (
	cdsInvalidProduct = cds.Error{
		Code:  "urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct",
		Title: "Invalid Software Product",
	}
	cdsInvalidBrand = cds.Error{Code: "urn:au-cds:error:cds-register:Field/InvalidBrand", Title: "Invalid Brand"}
)

type SSAIssuer interface {
	// Issue gives app.ErrNotBrandOfSoftwareProduct or an error that wraps
	// domain.ErrNotActive when it refuses the request.
	Issue(ctx context.Context, softwareProductID, dataRecipientBrandID string) (string, error)
}

// softwareStatementAssertion writes the JWT with no JSON quotes, as the ACCC
// implementations expect, with the content type that the CDS gives.
func softwareStatementAssertion(ssa SSAIssuer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		caller := callerID(r.Context())

		product, brand := r.PathValue("softwareProductId"), r.PathValue("dataRecipientBrandId")
		if product != caller {
			cds.WriteError(w, http.StatusNotFound, cdsInvalidProduct, product)
			return
		}

		token, err := ssa.Issue(r.Context(), caller, brand)
		switch {
		case errors.Is(err, app.ErrNotBrandOfSoftwareProduct):
			cds.WriteError(w, http.StatusNotFound, cdsInvalidBrand, brand)
		case errors.Is(err, domain.ErrNotActive):
			cds.WriteError(
				w,
				http.StatusForbidden,
				cds.ADRStatusNotActive,
				"The software product or its data recipient is not ACTIVE.",
			)
		case err != nil:
			unexpected(w, r, err)
		default:
			w.Header().Set("Content-Type", "application/json")
			//nolint:gosec // G705: the Register signed this JWT, so it has only base64url characters and dots
			_, _ = io.WriteString(w, token)
		}
	})
}
