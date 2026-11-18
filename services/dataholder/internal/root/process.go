package root

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/eligibilityjobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/grantstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/noticejobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/recipientnotice"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registerstatus"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/removaljobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/signingkeystore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusjobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

type process struct {
	listeners []httpserver.Listener
	// jobs is nil when the process does not work jobs.
	jobs *river.Client[pgx.Tx]
}

func newProcess(ctx context.Context, pool *pgxpool.Pool, cfg config, m mode) (process, error) {
	signingKeys := signingkeystore.New(pool)
	if err := signingKeys.EnsureSigningKey(ctx); err != nil {
		return process{}, fmt.Errorf("create the signing key: %w", err)
	}

	inserter, err := jobs.NewInsertOnlyClient(pool)
	if err != nil {
		return process{}, fmt.Errorf("create the job inserter: %w", err)
	}

	events := newEventStore(pool, inserter)
	customers := customerstore.New(pool, eligibilityjobs.Hook(inserter), transactionstore.Insert)
	grants := grantstore.New(pool)
	parts := shared{
		events:        events,
		registrations: registrationstore.NewQueries(pool),
		signingKeys:   signingKeys,
		customers:     customers,
		grants:        grants,
		statuses:      statusstore.New(pool, removaljobs.Hook(inserter)),
		arrangements: app.Arrangements{
			Store:     events,
			Grants:    grants,
			Customers: customers,
			Reader:    arrangementstore.NewQueries(pool),
			Now:       time.Now,
		},
	}

	// The groups of routes on the internal listener each apply their own policy, so the mux has no middleware.
	internal := http.NewServeMux()
	httpserver.RegisterProbes(internal, readiness(pool, parts.statuses, m))

	proc := process{listeners: []httpserver.Listener{{Port: internalPort, Handler: telemetry.NewHTTPHandler(internal)}}}

	if m.api {
		listeners, err := apiListeners(ctx, pool, cfg.api, parts, internal)
		if err != nil {
			return process{}, err
		}

		proc.listeners = append(proc.listeners, listeners...)
	}

	if m.workers {
		workers := river.NewWorkers()
		river.AddWorker(workers, noticejobs.Worker{Notices: app.Notices{
			Store:         events,
			Registrations: parts.registrations,
			Sender: recipientnotice.Sender{
				BrandID: cfg.workers.brandID,
				Keys:    signingKeys.SigningKeys,
				Client:  recipientnotice.NewClient(http.DefaultTransport),
				NewID:   uuid.NewString,
				Now:     time.Now,
			},
			Now: time.Now,
		}})
		river.AddWorker(workers, eligibilityjobs.Worker{Arrangements: parts.arrangements})
		river.AddWorker(workers, statusjobs.Worker{Statuses: app.Statuses{
			Source: registerstatus.NewClient(cfg.workers.registerPublicBaseURL, http.DefaultTransport),
			Store:  parts.statuses,
		}})
		river.AddWorker(workers, removaljobs.Worker{Products: parts.arrangements})

		proc.jobs, err = jobs.NewWorkerClient(pool, workers, statusjobs.PeriodicJob(cfg.workers.statusInterval))
		if err != nil {
			return process{}, fmt.Errorf("create the job client: %w", err)
		}
	}

	return proc, nil
}

func readiness(pool *pgxpool.Pool, statuses *statusstore.Store, m mode) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := pool.Ping(ctx); err != nil {
			return err
		}

		if !m.api {
			return nil
		}

		found, err := statuses.HasStatuses(ctx)
		if err != nil {
			return err
		}

		if !found {
			return errors.New("the data holder has not got the statuses of the software products")
		}

		return nil
	}
}

// shared are the parts of the data holder that the APIs and the workers use.
type shared struct {
	events        *eventstore.Store
	registrations registrationstore.Queries
	signingKeys   *signingkeystore.Store
	customers     *customerstore.Store
	grants        *grantstore.Store
	statuses      *statusstore.Store
	arrangements  app.Arrangements
}
