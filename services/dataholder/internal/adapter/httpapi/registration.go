package httpapi

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"slices"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// An SSA and the client metadata are a few kilobytes.
const maxRegistrationRequestBytes = 64 << 10

const registrationMediaType = "application/jwt"

// Registrar gives a domain.RegistrationError when it refuses a request, and app.ErrRegistrationNotFound when no
// registration has the client ID.
type Registrar interface {
	Register(ctx context.Context, requestJWT string) (domain.Registration, error)
	Update(ctx context.Context, clientID, requestJWT string) (domain.Registration, error)
	Delete(ctx context.Context, clientID string) error
}

// registrationErrorJSON has the format of RFC 7591, section 3.2.2.
type registrationErrorJSON struct {
	Error       string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

func register(registrar Registrar) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestJWT, ok := registrationRequestJWT(w, r)
		if !ok {
			return
		}

		registration, err := registrar.Register(r.Context(), requestJWT)
		writeRegistration(w, r, http.StatusCreated, registration, err)
	})
}

// registrationRequestJWT writes the response when the request has no body that the registrar can read.
func registrationRequestJWT(w http.ResponseWriter, r *http.Request) (requestJWT string, ok bool) {
	if mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mediaType != registrationMediaType {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return "", false
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRegistrationRequestBytes))
	if err != nil {
		writeRegistrationRefusal(w, r, domain.RegistrationError{
			Code:        domain.InvalidClientMetadata,
			Description: "the request body cannot be read, or it is too large",
		})

		return "", false
	}

	return strings.TrimSpace(string(body)), true
}

// writeRegistration writes the registration with successStatus, or the response for err.
func writeRegistration(
	w http.ResponseWriter,
	r *http.Request,
	successStatus int,
	registration domain.Registration,
	err error,
) {
	w.Header().Set("Cache-Control", "no-store")

	if refusal, ok := errors.AsType[domain.RegistrationError](err); ok {
		writeRegistrationRefusal(w, r, refusal)
		return
	}
	// A different request deleted the registration after the token check.
	if errors.Is(err, app.ErrRegistrationNotFound) {
		invalidToken(w)
		return
	}

	if err != nil {
		registrationFailure(w, r, err)
		return
	}

	jsonhttp.Write(w, successStatus, registration)
}

func writeRegistrationRefusal(w http.ResponseWriter, r *http.Request, refusal domain.RegistrationError) {
	slog.InfoContext(
		r.Context(), "registration refused",
		"error.type", string(refusal.Code), "error.message", refusal.Description,
	)
	jsonhttp.Write(
		w,
		http.StatusBadRequest,
		registrationErrorJSON{Error: string(refusal.Code), Description: refusal.Description},
	)
}

type TokenVerifier interface {
	// VerifyAccessToken gives an error that wraps authserver.ErrInvalidToken when the caller must get a new token.
	VerifyAccessToken(ctx context.Context, token string, clientCert *x509.Certificate) (authserver.AccessToken, error)
}

type ctxKeyRegistration struct{}

// ownRegistration lets a request pass only with a token of the client in the path, and gives the handler the
// registration of that client. It gives the RFC 6750 section 3 responses with no body, as the CDS specification has
// them.
func ownRegistration(tokens TokenVerifier, registrations app.RegistrationReader) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			credentials, ok := bearerToken(r)
			if !ok {
				noToken(w)
				return
			}

			token, err := tokens.VerifyAccessToken(r.Context(), credentials, clientCertificate(r))
			if errors.Is(err, authserver.ErrInvalidToken) {
				invalidToken(w)
				return
			}

			if err != nil {
				registrationFailure(w, r, err)
				return
			}

			registration, err := registrations.Registration(r.Context(), r.PathValue("ClientId"))
			if errors.Is(err, app.ErrRegistrationNotFound) {
				invalidToken(w)
				return
			}

			if err != nil {
				registrationFailure(w, r, err)
				return
			}

			if !slices.Contains(token.Scopes, authserver.RegistrationScope) {
				w.Header().Set(
					"WWW-Authenticate",
					`Bearer error="insufficient_scope", scope="`+authserver.RegistrationScope+`"`,
				)
				w.WriteHeader(http.StatusForbidden)

				return
			}

			if token.ClientID != registration.ClientID {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRegistration{}, registration)))
		})
	}
}

func readRegistration(w http.ResponseWriter, r *http.Request) {
	registration, _ := r.Context().Value(ctxKeyRegistration{}).(domain.Registration)
	writeRegistration(w, r, http.StatusOK, registration, nil)
}

func updateRegistration(registrar Registrar) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestJWT, ok := registrationRequestJWT(w, r)
		if !ok {
			return
		}

		registration, err := registrar.Update(r.Context(), r.PathValue("ClientId"), requestJWT)
		writeRegistration(w, r, http.StatusOK, registration, err)
	})
}

func deleteRegistration(registrar Registrar) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := registrar.Delete(r.Context(), r.PathValue("ClientId"))
		// A different request deleted the registration after the token check.
		if errors.Is(err, app.ErrRegistrationNotFound) {
			invalidToken(w)
			return
		}

		if err != nil {
			registrationFailure(w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// bearerToken reads only the Authorization header: a token in the query or the body does not count.
func bearerToken(r *http.Request) (string, bool) {
	scheme, credentials, _ := strings.Cut(r.Header.Get("Authorization"), " ")

	return credentials, strings.EqualFold(scheme, "Bearer")
}

func clientCertificate(r *http.Request) *x509.Certificate {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return nil
	}

	return r.TLS.PeerCertificates[0]
}

func noToken(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)
}

func invalidToken(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
	w.WriteHeader(http.StatusUnauthorized)
}

// registrationFailure does not log the message of err, because it can contain data of the request.
func registrationFailure(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "registration request failed", "error.type", fmt.Sprintf("%T", err))
	jsonhttp.Write(w, http.StatusInternalServerError, registrationErrorJSON{Error: "server_error"})
}
