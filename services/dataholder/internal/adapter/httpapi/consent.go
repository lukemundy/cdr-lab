package httpapi

import (
	"context"
	"crypto/x509"
	"errors"
	"log/slog"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	scopeOpenID        = "openid"
	scopeProfile       = "profile"
	scopeCustomerBasic = "common:customer.basic:read"
)

type ConsentTokenVerifier interface {
	// VerifyConsentToken gives an error that wraps authserver.ErrInvalidToken when the caller must get a new token.
	VerifyConsentToken(
		ctx context.Context,
		token string,
		clientCert *x509.Certificate,
	) (authserver.ConsentToken, error)
}

type ArrangementChecker interface {
	// RequireCurrent gives app.ErrArrangementEnded for an arrangement that ended, and app.ErrArrangementNotFound.
	RequireCurrent(ctx context.Context, arrangementID string) error
}

type ClientStatusReader interface {
	Status(ctx context.Context, clientID string) (domain.SoftwareProductStatus, error)
}

type CustomerReader interface {
	Customer(ctx context.Context, arrangementID string) (domain.Customer, error)
}

// notImplemented has the APIs of the pinned CDS specifications on the resource base URI that the data holder does not
// serve.
var notImplemented = []string{
	"GET /cds-au/v1/common/customer/detail",
	"GET /cds-au/v1/banking/accounts/{accountId}/direct-debits",
	"GET /cds-au/v1/banking/accounts/direct-debits",
	"POST /cds-au/v1/banking/accounts/direct-debits",
	"GET /cds-au/v1/banking/accounts/{accountId}/payments/scheduled",
	"GET /cds-au/v1/banking/payments/scheduled",
	"POST /cds-au/v1/banking/payments/scheduled",
	"GET /cds-au/v1/banking/accounts/{accountId}/payments/plans",
	"GET /cds-au/v1/banking/accounts/payments/plans",
	"GET /cds-au/v1/banking/payees",
	"GET /cds-au/v1/banking/payees/{payeeId}",
}

func registerConsentData(routes *httproute.Routes, s Secure) {
	baseURL := strings.TrimSuffix(s.ResourceBaseURL, "/")

	logged := routes.With(logInteraction)

	userInfo := logged.With(
		consentToken(s, scopeOpenID, userInfoNotActive),
		recordDisclosure(s.Disclosures, userInfoScope),
	)
	userInfo.Handle("GET "+s.UserInfoPath, userInfoClaims(s.Customers))
	userInfo.Handle("POST "+s.UserInfoPath, userInfoClaims(s.Customers))

	resource := func(scope string, minV, maxV int) *httproute.Routes {
		return logged.With(
			consentToken(s, scope, writeNotActive),
			recordDisclosure(s.Disclosures, func(authserver.ConsentToken) string { return scope }),
			cds.Versioned(minV, maxV),
			resourceHeaders,
		)
	}
	resource(scopeCustomerBasic, 1, 1).Handle("GET /cds-au/v1/common/customer", customer(baseURL, s.Customers))
	resource(scopeAccountsBasic, 3, 3).Handle("GET /cds-au/v1/banking/accounts", accounts(baseURL, s.Accounts))
	resource(scopeAccountsDetail, 5, 5).
		Handle("GET /cds-au/v1/banking/accounts/{accountId}", accountDetail(baseURL, s.Accounts))
	resource(scopeAccountsBasic, 1, 1).
		Handle("GET /cds-au/v1/banking/accounts/{accountId}/balance", accountBalance(baseURL, s.Balances))
	resource(scopeAccountsBasic, 2, 2).
		Handle("GET /cds-au/v1/banking/accounts/balances", bulkBalances(baseURL, s.Balances))
	resource(scopeAccountsBasic, 1, 1).
		Handle("POST /cds-au/v1/banking/accounts/balances", specificBalances(baseURL, s.Balances))
	resource(scopeTransactions, 2, 2).Handle(
		"GET /cds-au/v1/banking/accounts/{accountId}/transactions", transactions(baseURL, s.Transactions),
	)
	resource(scopeTransactions, 3, 3).Handle(
		"GET /cds-au/v1/banking/accounts/{accountId}/transactions/{transactionId}",
		transactionDetail(baseURL, s.Transactions),
	)

	// A path that is not implemented has no scope to check, so it needs no token.
	for _, pattern := range notImplemented {
		routes.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			cds.WriteError(
				w, http.StatusNotFound, cds.ResourceNotImplemented, "The data holder does not implement this API.",
			)
		}))
	}
}

type ctxKeyConsentToken struct{}

func consentTokenOf(r *http.Request) authserver.ConsentToken {
	token, _ := r.Context().Value(ctxKeyConsentToken{}).(authserver.ConsentToken)

	return token
}

// notActiveRefusal is different on the UserInfo endpoint and on the resource base URI.
type notActiveRefusal func(w http.ResponseWriter, status domain.SoftwareProductStatus)

func userInfoNotActive(w http.ResponseWriter, _ domain.SoftwareProductStatus) { invalidToken(w) }

// consentToken also refuses the token of an arrangement that ended, in case the end did not revoke the token.
func consentToken(s Secure, scope string, notActive notActiveRefusal) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			credentials, ok := bearerToken(r)
			if !ok {
				noToken(w)
				return
			}

			token, err := s.ConsentTokens.VerifyConsentToken(r.Context(), credentials, clientCertificate(r))
			if errors.Is(err, authserver.ErrInvalidToken) {
				invalidToken(w)
				return
			}

			if err != nil {
				cds.WriteUnexpected(w, r, err, "The data holder cannot verify the access token.")
				return
			}

			err = s.Arrangements.RequireCurrent(r.Context(), token.ArrangementID)
			if errors.Is(err, app.ErrArrangementEnded) || errors.Is(err, app.ErrArrangementNotFound) {
				invalidToken(w)
				return
			}

			if err != nil {
				cds.WriteUnexpected(w, r, err, "The data holder cannot read the arrangement of the access token.")
				return
			}

			status, err := s.ClientStatuses.Status(r.Context(), token.ClientID)
			if err != nil {
				cds.WriteUnexpected(w, r, err, "The data holder cannot read the status of the software product.")
				return
			}

			if status != domain.SoftwareProductActive {
				notActive(w, status)
				return
			}

			if !slices.Contains(token.Scopes, scope) {
				w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope", scope="`+scope+`"`)
				writeInvalidConsent(w, scope)

				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyConsentToken{}, token)))
		})
	}
}

func writeNotActive(w http.ResponseWriter, status domain.SoftwareProductStatus) {
	cds.WriteError(w, http.StatusForbidden, cds.ADRStatusNotActive, string(status))
}

func writeInvalidConsent(w http.ResponseWriter, scope string) {
	cds.WriteError(
		w, http.StatusForbidden, cds.InvalidConsent, "The consent does not have the scope "+scope+".",
	)
}

func resourceHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const authDate, customerIP = "x-fapi-auth-date", "x-fapi-customer-ip-address"

		date := r.Header.Get(authDate)
		if date == "" {
			cds.WriteError(w, http.StatusBadRequest, cds.HeaderMissing, authDate)
			return
		}

		// http.ParseTime accepts the three formats of an HTTP-date (RFC 9110, section 5.6.7).
		if _, err := http.ParseTime(date); err != nil {
			cds.WriteError(w, http.StatusBadRequest, cds.HeaderInvalid, authDate)
			return
		}

		if address := r.Header.Get(customerIP); address != "" {
			if _, err := netip.ParseAddr(address); err != nil {
				cds.WriteError(w, http.StatusBadRequest, cds.HeaderInvalid, customerIP)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter

	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	if s.status == 0 {
		s.status = status
	}

	s.ResponseWriter.WriteHeader(status)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}

	return s.ResponseWriter.Write(b)
}

// logInteraction logs each request with its x-fapi-interaction-id (FAPI 1.0 Baseline, section 6.2.1). The attribute
// names are those of the OpenTelemetry semantic conventions for HTTP.
func logInteraction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		slog.InfoContext(
			r.Context(), "data of a consent requested",
			"http.route", r.Pattern,
			"http.response.status_code", max(recorder.status, http.StatusOK),
			"http.response.header.x-fapi-interaction-id", w.Header().Values("x-fapi-interaction-id"),
		)
	})
}

type personJSON struct {
	LastUpdateTime string   `json:"lastUpdateTime"`
	FirstName      string   `json:"firstName,omitempty"`
	LastName       string   `json:"lastName"`
	MiddleNames    []string `json:"middleNames"`
}

type customerJSON struct {
	CustomerUType string     `json:"customerUType"`
	Person        personJSON `json:"person"`
}

func customer(baseURL string, customers CustomerReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := customers.Customer(r.Context(), consentTokenOf(r).ArrangementID)
		if err != nil {
			cds.WriteUnexpected(w, r, err, "The data holder cannot read the customer.")
			return
		}

		jsonhttp.Write(w, http.StatusOK, cds.NewResponse(r, baseURL, customerJSON{
			CustomerUType: "person",
			Person: personJSON{
				LastUpdateTime: c.LastUpdated.UTC().Format(time.RFC3339),
				FirstName:      c.FirstName,
				LastName:       c.LastName,
				MiddleNames:    []string{},
			},
		}))
	})
}

// userInfoClaims gives the claims of OpenID Connect Core, section 5.1, that the CDS permits.
func userInfoClaims(customers CustomerReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := consentTokenOf(r)
		claims := map[string]any{"sub": token.Subject}

		if slices.Contains(token.Scopes, scopeProfile) {
			c, err := customers.Customer(r.Context(), token.ArrangementID)
			if err != nil {
				cds.WriteUnexpected(w, r, err, "The data holder cannot read the customer.")
				return
			}

			claims["name"] = strings.TrimSpace(c.FirstName + " " + c.LastName)
			claims["family_name"] = c.LastName
			claims["updated_at"] = c.LastUpdated.Unix()

			if c.FirstName != "" {
				claims["given_name"] = c.FirstName
			}
		}

		jsonhttp.Write(w, http.StatusOK, claims)
	})
}
