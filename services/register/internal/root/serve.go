package root

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/castore"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/crlhttp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/crljobs"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/idp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/ssa"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

func serve(ctx context.Context, pool *pgxpool.Pool, m mode) error {
	cfg, err := loadConfig(m)
	if err != nil {
		return err
	}

	if err := migrateOnStart(ctx, pool); err != nil {
		return fmt.Errorf("apply the database migrations: %w", err)
	}

	proc, err := newProcess(ctx, pool, cfg, m)
	if err != nil {
		return err
	}

	if proc.jobs == nil {
		return httpserver.ListenAndServe(ctx, proc.listeners)
	}

	return jobs.Run(ctx, proc.jobs, func(ctx context.Context) error {
		return httpserver.ListenAndServe(ctx, proc.listeners)
	})
}

type process struct {
	listeners []httpserver.Listener
	// jobs is nil when the process does not work jobs.
	jobs *river.Client[pgx.Tx]
}

func newProcess(ctx context.Context, pool *pgxpool.Pool, cfg config, m mode) (process, error) {
	authority := app.NewCertificateAuthority(castore.New(pool), cfg.intermediate, time.Now)

	probes := http.NewServeMux()
	httpserver.RegisterProbes(probes, readiness(pool, authority, m))

	proc := process{listeners: []httpserver.Listener{{Port: probePort, Handler: telemetry.NewHTTPHandler(probes)}}}

	if m.api {
		// A new CRL at each start has the signature of the intermediate key that this process read.
		if err := authority.SignCRL(ctx); err != nil {
			return process{}, fmt.Errorf("sign the CRL: %w", err)
		}

		listeners, err := apiListeners(ctx, pool, cfg, authority)
		if err != nil {
			return process{}, err
		}

		proc.listeners = append(proc.listeners, listeners...)
	}

	if m.workers {
		workers := river.NewWorkers()
		river.AddWorker(workers, crljobs.Worker{CRLs: authority})

		var err error

		proc.jobs, err = jobs.NewWorkerClient(pool, workers, crljobs.PeriodicJob(crlRenewalInterval))
		if err != nil {
			return process{}, fmt.Errorf("create the job client: %w", err)
		}
	}

	return proc, nil
}

// readiness of a process with the APIs also needs a current CRL, because each relying party refuses each CDR
// certificate without one.
func readiness(pool *pgxpool.Pool, authority app.CertificateAuthority, m mode) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := pool.Ping(ctx); err != nil {
			return err
		}

		if !m.api {
			return nil
		}

		_, err := authority.CurrentCRL(ctx)

		return err
	}
}

func apiListeners(
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg config,
	authority app.CertificateAuthority,
) ([]httpserver.Listener, error) {
	idpStore := idp.NewStore(pool)
	for _, set := range []app.KeySet{app.AccessTokenKeys, app.SSAKeys} {
		if err := idpStore.EnsureSigningKey(ctx, set); err != nil {
			return nil, fmt.Errorf("create the signing key of the key set %s: %w", set, err)
		}
	}

	queries := readmodel.NewQueries(pool)
	ssaKeys := idpStore.SigningKeys(app.SSAKeys)

	authServer, err := authserver.NewRegister(authserver.RegisterConfig{
		PublicBaseURL: cfg.publicBaseURL,
		SecureBaseURL: cfg.secureBaseURL,
		Clients:       idp.Clients{Reader: queries},
		JTIs:          idpStore,
		SigningKeys:   idpStore.SigningKeys(app.AccessTokenKeys),
	})
	if err != nil {
		return nil, fmt.Errorf("create the authorisation server: %w", err)
	}

	api := http.NewServeMux()
	httpapi.Register(api, queries, ssaKeys, cfg.publicBaseURL)
	authServer.RegisterPublic(api)

	operatorUseCases := app.NewOperator(newEventStore(pool))
	operator := http.NewServeMux()
	issuer := app.NewCertificateIssuer(newEventStore(pool), cfg.intermediate, time.Now)
	operatorapi.Register(operator, operatorUseCases, issuer, authority, cfg.cdrCAs, idpStore, cfg.operator)

	secure := http.NewServeMux()
	ssaIssuer := app.SSAIssuer{Owners: queries, Signer: ssa.Signer{Keys: ssaKeys}, Now: time.Now, NewID: uuid.NewString}
	httpapi.RegisterSecure(secure, queries, authServer, ssaIssuer, cfg.secureBaseURL)
	authServer.RegisterSecure(secure)

	clientCAs := x509.NewCertPool()
	for _, ca := range cfg.cdrCAs {
		clientCAs.AddCert(ca)
	}

	crl := http.NewServeMux()
	crlhttp.Register(crl, authority)

	consoles, err := consoleListeners(
		cfg.console,
		console.UI(),
		console.Ports{
			Readers: queries, Keys: idpStore, KeyChanger: idpStore, Operator: operatorUseCases, Issuer: issuer,
			CertificateAuthority: authority, CABundle: cfg.cdrCAs,
		},
	)
	if err != nil {
		return nil, err
	}

	return append([]httpserver.Listener{
		{
			Port:      apiPort,
			Handler:   httpapi.AllowAnyOrigin(cds.InteractionID(telemetry.NewHTTPHandler(api))),
			TLSConfig: cds.TLSConfig(cfg.publicCert),
		},
		{Port: operatorPort, Handler: telemetry.NewHTTPHandler(operator)},
		{
			Port:    securePort,
			Handler: cds.InteractionID(telemetry.NewHTTPHandler(secure)),
			// The CDR root signs only the intermediate, so the CRL of the intermediate covers each client certificate.
			TLSConfig: cds.MutualTLSConfig(ctx, cfg.secureCert, clientCAs, authority.CheckRevocation),
		},
		{Port: crlPort, Handler: telemetry.NewHTTPHandler(crl)},
	}, consoles...), nil
}
