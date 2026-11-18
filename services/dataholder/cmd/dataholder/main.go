// Command dataholder runs one brand of the mock CDR data holder.
//
// Usage:
//
//	dataholder [serve]                        serve the APIs and work the background jobs
//	dataholder api                            serve only the APIs
//	dataholder worker                         work only the background jobs, and serve only the health probes
//	dataholder migrate                        apply the database migrations, then stop
//	dataholder seed [-if-empty] [<file.json>] load a full set of products into a data holder that has none, then stop:
//	                                          the product catalogue of the brand, or the products of the file
//
// Each subcommand reads the database connection from the libpq variables (PG*), and the telemetry settings from the
// OpenTelemetry variables (OTEL_*). DATABASE_MIGRATE_ON_START is optional.
//
// The APIs (api and serve) and seed need exactly one of these settings, which selects the brand:
//
//	DATA_HOLDER_BRAND       the name of a preset brand of the data holder
//	DATA_HOLDER_BRAND_DIR   the path of a brand directory (services/dataholder/brands/README.md)
//
// The APIs also need these settings:
//
//	PUBLIC_BASE_URL, PUBLIC_TLS_CERT_FILE, PUBLIC_TLS_KEY_FILE
//	SECURE_BASE_URL, SECURE_TLS_CERT_FILE, SECURE_TLS_KEY_FILE, CDR_CA_BUNDLE_FILE
//	REGISTER_JWKS_URL, OPERATOR_TOKEN, INTERNAL_ALLOWED_HOSTS, SMTP_ADDRESS, MAIL_FROM
//	CRL_MAX_AGE (optional, default 1m)
//	FIXED_OTPS (optional)
//	CONSOLE_ENABLED (optional)
//	BANK_SITE_ENABLED (optional), and BANK_SITE_ALLOWED_HOSTS and DATA_HOLDER_BRAND_ID when the online bank site
//	is on (lab-only)
//
// The workers (worker and serve) need these settings:
//
//	DATA_HOLDER_BRAND_ID, REGISTER_PUBLIC_BASE_URL, REGISTER_STATUS_INTERVAL
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lukemundy/cdr-lab/pkg/platform/lifecycle"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/root"
)

func main() {
	if err := lifecycle.Run(
		context.Background(),
		func(ctx context.Context) error { return root.Run(ctx, os.Args[1:]) },
	); err != nil {
		slog.Error("data holder stopped with error", "error", err)
		os.Exit(1)
	}
}
