// Package console serves the operator console of the Register: the embedded web UI and the console API.
package console

import (
	"crypto/x509"
	"embed"
	"io/fs"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

// The Register image copies the build of web/apps/register-console into dist, so a plain Go build has no UI.
//
//go:embed all:dist
var dist embed.FS

func UI() fs.FS {
	ui, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}

	return ui
}

type Readers interface {
	app.DataRecipientPageReader
	app.DataRecipientDetailReader
	app.DataHolderPageReader
	app.DataHolderDetailReader
	app.HistoryReader
}

type Ports struct {
	Readers              Readers
	Keys                 app.KeyReader
	KeyChanger           KeyChanger
	Operator             Operator
	Issuer               CertificateIssuer
	CertificateAuthority CertificateAuthority
	CABundle             []*x509.Certificate
}

// Register returns an error when ui has no build of the UI or allowedHosts is empty.
func Register(mux *http.ServeMux, ui fs.FS, allowedHosts []string, ports Ports) error {
	readers, keys := ports.Readers, ports.Keys

	return consolehttp.Console.Register(mux, ui, allowedHosts, func(routes *httproute.Routes) {
		routes.Handle("GET /api/data-recipients", dataRecipients(readers))
		routes.Handle("GET /api/data-recipients/{legalEntityId}", dataRecipient(readers))
		routes.Handle("GET /api/data-holders", dataHolders(readers))
		routes.Handle("GET /api/data-holders/{legalEntityId}", dataHolder(readers))
		routes.Handle("GET /api/data-recipients/{legalEntityId}/history",
			history(readers, app.DataRecipientType.StreamType, noDataRecipient))
		routes.Handle("GET /api/data-holders/{legalEntityId}/history",
			history(readers, app.DataHolderType.StreamType, noDataHolder))
		routes.Handle("GET /api/key-sets", keySets(keys))
		registerStatusActions(routes, ports.Operator)
		registerKeyActions(routes, ports.KeyChanger)
		registerCertificateActions(routes, ports.Operator)
		registerCertificateIssues(routes, ports.Issuer)
		registerCDRCertificates(routes, ports.CertificateAuthority, ports.CABundle)
		registerChanges(routes, ports.Operator)
		registerCreates(routes, ports.Operator)
	})
}
