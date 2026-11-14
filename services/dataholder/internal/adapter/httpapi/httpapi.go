// Package httpapi serves the CDR APIs of the data holder.
package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

// Register adds the public APIs to mux; baseURL is the public and product base URI of the brand, for the links.
func Register(mux *http.ServeMux, baseURL string, reader app.ProductReader, now func() time.Time) {
	baseURL = strings.TrimSuffix(baseURL, "/")

	routes := httproute.New(mux)

	routes.With(cds.Versioned(5, 5)).Handle("GET /cds-au/v1/banking/products", products(baseURL, reader, now))
	routes.With(cds.Versioned(7, 7)).
		Handle("GET /cds-au/v1/banking/products/{productId}", productDetail(baseURL, reader))
	routes.With(cds.Versioned(1, 1)).Handle("GET /cds-au/v1/discovery/status", status(baseURL, now))
	routes.With(cds.Versioned(1, 1)).Handle("GET /cds-au/v1/discovery/outages", outages(baseURL))

	mux.HandleFunc("/", notFound)
}

type Secure struct {
	// ResourceBaseURL is the resource base URI of the brand, for the links.
	ResourceBaseURL  string
	RegistrationPath string
	UserInfoPath     string
	Registrar        Registrar
	Tokens           TokenVerifier
	Registrations    app.RegistrationReader
	ConsentTokens    ConsentTokenVerifier
	Customers        CustomerReader
	Accounts         AccountReader
	Transactions     TransactionReader
	Balances         BalanceReader
	ClientStatuses   ClientStatusReader
	Disclosures      DisclosureRecorder
}

// RegisterSecure needs a listener that verifies the client certificate.
func RegisterSecure(mux *http.ServeMux, s Secure) {
	routes := httproute.New(mux)
	routes.Handle("POST "+s.RegistrationPath, register(s.Registrar))

	own := routes.With(ownRegistration(s.Tokens, s.Registrations))
	own.Handle("GET "+s.RegistrationPath+"/{ClientId}", http.HandlerFunc(readRegistration))
	own.Handle("PUT "+s.RegistrationPath+"/{ClientId}", updateRegistration(s.Registrar))
	own.Handle("DELETE "+s.RegistrationPath+"/{ClientId}", deleteRegistration(s.Registrar))

	registerConsentData(routes, s)

	mux.HandleFunc("/", notFound)
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	cds.WriteError(w, http.StatusNotFound, cds.ResourceNotFound, "The URL is not a data holder API endpoint.")
}

// AllowAnyOrigin does not apply to pagesPath and the paths below it, which serve pages to a consumer and not an API.
func AllowAnyOrigin(next http.Handler, pagesPath string) http.Handler {
	allowed := cds.AllowAnyOrigin(next, "x-v", "Retry-After", "x-fapi-interaction-id")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pagesPath || strings.HasPrefix(r.URL.Path, pagesPath+"/") {
			next.ServeHTTP(w, r)
			return
		}

		allowed.ServeHTTP(w, r)
	})
}

type statusJSON struct {
	Status     string `json:"status"`
	UpdateTime string `json:"updateTime"`
}

func status(baseURL string, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := statusJSON{Status: "OK", UpdateTime: now().UTC().Format(time.RFC3339)}
		jsonhttp.Write(w, http.StatusOK, cds.NewResponse(r, baseURL, data))
	})
}

type outagesJSON struct {
	Outages []struct{} `json:"outages"`
}

func outages(baseURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonhttp.Write(w, http.StatusOK, cds.NewResponse(r, baseURL, outagesJSON{Outages: []struct{}{}}))
	})
}
