// Command datarecipient runs one brand of the mock data recipient.
//
// Usage:
//
//	datarecipient [serve]   serve the web app, and the health probes on the internal listener
//	datarecipient migrate   apply the database migrations, then stop
//
// Each subcommand reads the database connection from the libpq variables (PG*), and the telemetry settings from the
// OpenTelemetry variables (OTEL_*). DATABASE_MIGRATE_ON_START is optional.
//
// serve needs exactly one of these settings, which selects the brand:
//
//	DATA_RECIPIENT_BRAND       the name of a preset brand of the data recipient
//	DATA_RECIPIENT_BRAND_DIR   the path of a brand directory (services/datarecipient/brands/README.md)
//
// serve also needs these settings:
//
//	WEB_APP_ALLOWED_HOSTS                         the Host header values of the web app
//	WEB_APP_TLS_CERT_FILE, WEB_APP_TLS_KEY_FILE   the certificate and key of the web app, from the public CA
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lukemundy/cdr-lab/pkg/platform/lifecycle"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/root"
)

func main() {
	if err := lifecycle.Run(
		context.Background(),
		func(ctx context.Context) error { return root.Run(ctx, os.Args[1:]) },
	); err != nil {
		slog.Error("data recipient stopped with error", "error", err)
		os.Exit(1)
	}
}
