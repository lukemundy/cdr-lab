// Package banksite serves the online bank site of the data holder: the embedded web UI and the site API. The site is
// lab-only.
package banksite

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The data holder image copies the build of web/apps/dataholder-bank into dist, so a plain Go build has no UI.
//
//go:embed all:dist
var dist embed.FS

const notCompleted = "The data holder could not complete the request."

var Settings = consolehttp.Settings{
	Name:            "online bank site",
	EnabledEnv:      "BANK_SITE_ENABLED",
	AllowedHostsEnv: "BANK_SITE_ALLOWED_HOSTS",
	ExampleHost:     "localhost:8185",
}

func UI() fs.FS {
	ui, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}

	return ui
}

type Ports struct {
	// BrandName is the brandName of the brand on the Register.
	BrandName string
	// BrandLayout is the layout of the site, and BrandPhotos the file name of each photo, from the brand directory.
	BrandLayout string
	BrandPhotos []string
	// BrandFiles adds the stylesheet, the logo, the mark and the photos of the brand; a test can leave it nil.
	BrandFiles func(*httproute.Routes)
	// Products is the source of Get Products, so that the site gives the same products.
	Products app.ProductReader
	// BrandID is the dataHolderBrandId of the brand, which names the session cookie.
	BrandID    string
	Customers  Customers
	Accounts   AccountReader
	Opener     AccountOpener
	Consents   ConsentReader
	Withdrawer ConsentWithdrawer
	Now        func() time.Time
}

// Register returns an error when ui has no build of the UI or allowedHosts is empty.
func Register(mux *http.ServeMux, ui fs.FS, allowedHosts []string, ports Ports) error {
	return Settings.Register(mux, ui, allowedHosts, func(routes *httproute.Routes) {
		if ports.BrandFiles != nil {
			ports.BrandFiles(routes)
		}

		routes.Handle("GET /api/brand", brand(ports))
		routes.Handle("GET /api/products", listProducts(ports.Products, ports.Now))
		routes.Handle("GET /api/products/{productId}", readProduct(ports.Products, ports.Now))

		s := newSessions(newSessionStore(maxSessions, ports.Now), ports.BrandID)
		signedIn := routes.With(s.require)
		routes.Handle("POST /api/customers", signUp(ports.Customers, s))
		routes.Handle("POST /api/session", signIn(ports.Customers, s))
		routes.Handle("DELETE /api/session", signOut(s))
		signedIn.Handle("GET /api/session", readSession(ports.Customers, s))
		signedIn.Handle("GET /api/accounts", listAccounts(ports.Accounts, s, ports.Now))
		signedIn.Handle("POST /api/accounts", openAccount(ports.Opener, ports.Accounts, s, ports.Now))
		signedIn.Handle("GET /api/accounts/{accountId}/balance", readBalance(ports.Accounts, s))
		signedIn.Handle("GET /api/accounts/{accountId}/transactions", listTransactions(ports.Accounts, s))
		signedIn.Handle("POST /api/accounts/{accountId}/close", closeAccount(ports.Customers, s))
		signedIn.Handle("GET /api/consents", listConsents(ports.Consents, s, ports.Now))
		signedIn.Handle("POST /api/consents/{arrangementId}/withdraw", withdrawConsent(ports.Withdrawer, s))
	})
}

type brandBody struct {
	BrandName           string   `json:"brandName"`
	DemonstrationNotice string   `json:"demonstrationNotice"`
	Layout              string   `json:"layout"`
	Photos              []string `json:"photos"`
}

func brand(ports Ports) http.Handler {
	body := brandBody{
		BrandName:           ports.BrandName,
		DemonstrationNotice: domain.DemonstrationNotice(ports.BrandName),
		Layout:              ports.BrandLayout,
		Photos:              append([]string{}, ports.BrandPhotos...),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonhttp.Write(w, http.StatusOK, body)
	})
}

// product has the fields of BankingProductV6 that the data holder holds, because the list of Get Products has only
// these.
type product struct {
	ProductID       string     `json:"productId"`
	EffectiveFrom   *time.Time `json:"effectiveFrom,omitempty"`
	EffectiveTo     *time.Time `json:"effectiveTo,omitempty"`
	LastUpdated     time.Time  `json:"lastUpdated"`
	ProductCategory string     `json:"productCategory"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Brand           string     `json:"brand"`
	IsTailored      bool       `json:"isTailored"`
}

type productList struct {
	Products []product `json:"products"`
}

func listProducts(reader app.ProductReader, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		found, _, err := reader.Products(r.Context(), app.CurrentProducts(now()))
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		body := productList{Products: make([]product, 0, len(found))}
		for _, p := range found {
			body.Products = append(body.Products, product{
				ProductID:       p.ProductID,
				EffectiveFrom:   p.EffectiveFrom,
				EffectiveTo:     p.EffectiveTo,
				LastUpdated:     p.LastUpdated,
				ProductCategory: string(p.ProductCategory),
				Name:            p.Name,
				Description:     p.Description,
				Brand:           p.Brand,
				IsTailored:      p.IsTailored,
			})
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}

// readProduct uses the query of the list, so that the site and Get Products have one rule for a current product.
func readProduct(reader app.ProductReader, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := app.CurrentProducts(now())
		query.ProductID = r.PathValue("productId")

		found, _, err := reader.Products(r.Context(), query)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		if len(found) != 1 {
			problem.Write(w, http.StatusNotFound, "The bank has no current product with this ID.")
			return
		}

		jsonhttp.Write(w, http.StatusOK, found[0])
	})
}
