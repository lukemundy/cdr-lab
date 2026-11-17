package httpapi

import (
	"context"
	"crypto/x509"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const registerScope = "cdr-register:read"

type TokenVerifier interface {
	// VerifyAccessToken gives an error that wraps authserver.ErrInvalidToken when the caller must get a new token.
	VerifyAccessToken(ctx context.Context, token string, clientCert *x509.Certificate) (authserver.AccessToken, error)
}

type SecureReaders interface {
	app.RegisteredDataHolderBrandReader
	app.SoftwareProductClientReader
}

// RegisterSecure checks the token and the caller status before the version headers,
// so that a caller with no valid token learns nothing about the API.
func RegisterSecure(mux *http.ServeMux, readers SecureReaders, tokens TokenVerifier, ssa SSAIssuer, baseURL string) {
	baseURL = strings.TrimSuffix(baseURL, "/")

	routes := httproute.New(mux)
	routes.Use(
		bearerToken(tokens),
		activeCaller(readers),
	)

	routes.With(cds.Versioned(3, 3)).
		Handle("GET /cdr-register/v1/{industry}/data-holders/brands", dataHolderBrands(baseURL, readers))
	routes.With(cds.Versioned(4, 4), recipientIndustry).
		Handle(
			"GET /cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}"+
				"/software-products/{softwareProductId}/ssa",
			softwareStatementAssertion(ssa),
		)

	cds.RegisterFallback(mux, notFoundDetail)
}

type ctxKeyCaller struct{}

// bearerToken gives the RFC 6750 section 3 responses with no body, as the CDS specifies.
func bearerToken(tokens TokenVerifier) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme, credentials, _ := strings.Cut(r.Header.Get("Authorization"), " ")
			if !strings.EqualFold(scheme, "Bearer") {
				w.Header().Set("WWW-Authenticate", "Bearer")
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			var cert *x509.Certificate
			if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
				cert = r.TLS.PeerCertificates[0]
			}

			token, err := tokens.VerifyAccessToken(r.Context(), credentials, cert)
			switch {
			case errors.Is(err, authserver.ErrInvalidToken):
				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
				w.WriteHeader(http.StatusUnauthorized)
			case err != nil:
				unexpected(w, r, err)
			case !slices.Contains(token.Scopes, registerScope):
				w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope", scope="`+registerScope+`"`)
				w.WriteHeader(http.StatusForbidden)
			default:
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyCaller{}, token.ClientID)))
			}
		})
	}
}

// activeCaller reads the status on each call, because a token stays
// valid after its software product stops being ACTIVE.
func activeCaller(clients app.SoftwareProductClientReader) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			status, err := inactiveStatus(r.Context(), clients, callerID(r.Context()))
			if err != nil {
				unexpected(w, r, err)
				return
			}

			if status != "" {
				cds.WriteError(w, http.StatusForbidden, cds.ADRStatusNotActive, status)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// inactiveStatus gives the status that refuses the software product, or "" when nothing refuses it.
func inactiveStatus(
	ctx context.Context,
	clients app.SoftwareProductClientReader,
	softwareProductID string,
) (string, error) {
	client, err := clients.SoftwareProductClient(ctx, softwareProductID)

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return string(domain.Removed), nil
	case err != nil:
		return "", err
	case client.Status != domain.Active:
		return string(client.Status), nil
	case client.DataRecipientStatus != domain.DataRecipientActive:
		return string(client.DataRecipientStatus), nil
	default:
		return "", nil
	}
}

// callerID gives the software product ID of the access token, which bearerToken puts in the context.
func callerID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyCaller{}).(string)

	return id
}
