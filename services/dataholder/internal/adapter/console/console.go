// Package console serves the operator console of the data holder: the embedded web UI and the console API.
package console

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

// The data holder image copies the build of web/apps/dataholder-console into dist, so a plain Go build has no UI.
//
//go:embed all:dist
var dist embed.FS

const notCompleted = "The data holder could not complete the request."

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
	// BrandFiles adds the brand files of the console; a test can leave it nil.
	BrandFiles         func(*httproute.Routes)
	Customers          app.CustomerPageReader
	Creates            Creator
	Details            CustomerDetails
	Accounts           AccountCloser
	Arrangements       ArrangementWithdrawer
	Registrations      app.RegistrationRecords
	ClientArrangements ClientArrangements
	Statuses           app.StatusListReader
	Overview           app.OverviewReader
	Now                func() time.Time
}

// Register returns an error when ui has no build of the UI or allowedHosts is empty.
func Register(mux *http.ServeMux, ui fs.FS, allowedHosts []string, ports Ports) error {
	return consolehttp.Console.Register(mux, ui, allowedHosts, func(routes *httproute.Routes) {
		if ports.BrandFiles != nil {
			ports.BrandFiles(routes)
		}

		routes.Handle("GET /api/brand", brand(ports.BrandName))
		routes.Handle("GET /api/overview", overview(ports.Overview, ports.Now))
		routes.Handle("GET /api/customers", customers(ports.Customers, ports.Now))
		routes.Handle("POST /api/customers", createCustomer(ports.Creates))
		routes.Handle("GET /api/customers/{customerNumber}", readCustomer(ports.Details, ports.Now))
		routes.Handle("POST /api/customers/{customerNumber}/accounts", createAccount(ports.Creates))
		routes.Handle("POST /api/customers/{customerNumber}/accounts/{accountId}/close", closeAccount(ports.Accounts))
		routes.Handle(
			"POST /api/customers/{customerNumber}/arrangements/{arrangementId}/withdraw",
			withdrawArrangement(ports.Arrangements),
		)
		routes.Handle(
			"GET /api/customers/{customerNumber}/accounts/{accountId}/transactions",
			readTransactions(ports.Details),
		)
		routes.Handle(
			"POST /api/customers/{customerNumber}/accounts/{accountId}/transactions",
			createTransaction(ports.Creates),
		)
		routes.Handle("GET /api/registrations", registrations(ports.Registrations, ports.Now))
		routes.Handle("GET /api/registrations/{clientId}", readRegistration(ports.Registrations, ports.Now))
		routes.Handle(
			"GET /api/registrations/{clientId}/arrangements",
			readClientArrangements(ports.Registrations, ports.ClientArrangements, ports.Now),
		)
		routes.Handle("GET /api/software-product-statuses", softwareProductStatuses(ports.Statuses))
	})
}

type brandBody struct {
	BrandName string `json:"brandName"`
}

func brand(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonhttp.Write(w, http.StatusOK, brandBody{BrandName: name})
	})
}
