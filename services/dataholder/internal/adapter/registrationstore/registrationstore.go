// Package registrationstore keeps the read model of the client registrations in PostgreSQL and queries it.
package registrationstore

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func Projections() []eventstore.Projection {
	return []eventstore.Projection{{StreamType: app.ClientRegistrationType.StreamType, Project: project}}
}

// project gets the stream of one software product, whose stream ID is its software_id.
func project(ctx context.Context, tx pgx.Tx, softwareID string, recorded []eventsourcing.RecordedEvent) error {
	state, _, err := app.ClientRegistrationType.Fold(softwareID, recorded)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM registrations WHERE software_id = $1", softwareID); err != nil {
		return err
	}

	registeredAt, err := writeChanges(ctx, tx, softwareID, recorded)
	if err != nil || !state.Registered {
		return err
	}

	// The arrangement projection also writes this count, so the sequence of the projections has no effect.
	_, err = tx.Exec(
		ctx,
		`INSERT INTO registrations (client_id, software_id, registration, registered_at, updated_at, arrangement_count)
			VALUES ($1, $2, $3, $4, $5, (SELECT count(*) FROM arrangements WHERE client_id = $1))`,
		state.Registration.ClientID, softwareID, state.Registration, registeredAt, recorded[len(recorded)-1].RecordedAt,
	)

	return err
}

// writeChanges writes one change for each event, with the client_name of the registration at that event, and gives the
// time of the last registration.
func writeChanges(
	ctx context.Context,
	tx pgx.Tx,
	softwareID string,
	recorded []eventsourcing.RecordedEvent,
) (time.Time, error) {
	if _, err := tx.Exec(ctx, "DELETE FROM registration_changes WHERE software_id = $1", softwareID); err != nil {
		return time.Time{}, err
	}

	var (
		registeredAt time.Time
		current      domain.Registration
	)

	for _, r := range recorded {
		event, err := app.ClientRegistrationType.Codec.Decode(r.EventData)
		if err != nil {
			return time.Time{}, err
		}

		var kind app.ChangeKind

		switch e := event.(type) {
		case domain.ClientRegistered:
			kind, current, registeredAt = app.ChangeClientRegistered, e.Registration, r.RecordedAt
		case domain.ClientUpdated:
			kind, current = app.ChangeClientUpdated, e.Registration
		case domain.ClientDeleted:
			kind = app.ChangeClientDeleted
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO registration_changes (software_id, stream_version, client_id, client_name, kind, changed_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			softwareID, r.StreamVersion, current.ClientID, current.ClientName, kind, r.RecordedAt,
		); err != nil {
			return time.Time{}, err
		}
	}

	return registeredAt, nil
}

type Queries struct {
	pool *pgxpool.Pool
}

var (
	_ app.RegistrationReader  = Queries{}
	_ app.RegistrationRecords = Queries{}
)

func NewQueries(pool *pgxpool.Pool) Queries {
	return Queries{pool: pool}
}

func (q Queries) Registration(ctx context.Context, clientID string) (domain.Registration, error) {
	var registration domain.Registration

	err := q.pool.QueryRow(ctx, "SELECT registration FROM registrations WHERE client_id = $1", clientID).
		Scan(&registration)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Registration{}, app.ErrRegistrationNotFound
	}

	return registration, err
}

// The count of the arrangements that did not end uses the rule of the customer list.
const records = `SELECT r.registration, r.registered_at, r.updated_at, coalesce(s.status, ''), r.arrangement_count,
		(SELECT count(*) FROM arrangements a WHERE a.client_id = r.client_id AND a.ends_at > $1)
	FROM registrations r LEFT JOIN software_product_statuses s ON s.software_product_id = r.software_id`

var registrationSorts = map[app.RegistrationSort]func(a, b app.RegistrationRecord) int{
	// The byte order of the lower case is the order of lower(client_name) COLLATE "C" in the other lists.
	app.SortRegistrationsByClientName: func(a, b app.RegistrationRecord) int {
		return strings.Compare(strings.ToLower(a.Registration.ClientName), strings.ToLower(b.Registration.ClientName))
	},
	app.SortRegistrationsByRegisteredAt: func(a, b app.RegistrationRecord) int {
		return a.RegisteredAt.Compare(b.RegisteredAt)
	},
	app.SortRegistrationsByCurrentArrangements: func(a, b app.RegistrationRecord) int {
		return cmp.Compare(a.CurrentArrangements, b.CurrentArrangements)
	},
	app.SortRegistrationsByArrangements: func(a, b app.RegistrationRecord) int {
		return cmp.Compare(a.Arrangements, b.Arrangements)
	},
}

func (q Queries) RegistrationList(
	ctx context.Context,
	query app.RegistrationListQuery,
) ([]app.RegistrationRecord, error) {
	compare, ok := registrationSorts[query.Sort]
	if !ok {
		return nil, fmt.Errorf("unknown sort %q for the registrations", query.Sort)
	}

	rows, err := q.pool.Query(ctx, records, query.At)
	if err != nil {
		return nil, err
	}

	list, err := pgx.CollectRows(rows, scanRecord)
	// The sort is here, because no query reads the content of the registration column. The lab has few registrations.
	slices.SortFunc(list, func(a, b app.RegistrationRecord) int {
		order := compare(a, b)
		if query.Descending {
			order = -order
		}

		return cmp.Or(order, cmp.Compare(a.Registration.ClientID, b.Registration.ClientID))
	})

	return list, err
}

func (q Queries) RegistrationRecord(
	ctx context.Context,
	clientID string,
	at time.Time,
) (app.RegistrationRecord, error) {
	rows, err := q.pool.Query(ctx, records+" WHERE r.client_id = $2", at, clientID)
	if err != nil {
		return app.RegistrationRecord{}, err
	}

	record, err := pgx.CollectExactlyOneRow(rows, scanRecord)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.RegistrationRecord{}, app.ErrRegistrationNotFound
	}

	return record, err
}

func scanRecord(row pgx.CollectableRow) (app.RegistrationRecord, error) {
	var record app.RegistrationRecord

	err := row.Scan(
		&record.Registration, &record.RegisteredAt, &record.UpdatedAt, &record.Status,
		&record.Arrangements, &record.CurrentArrangements,
	)
	record.RegisteredAt, record.UpdatedAt = record.RegisteredAt.UTC(), record.UpdatedAt.UTC()

	return record, err
}

// Clients gives the authorisation server each registration as a client.
type Clients struct {
	Reader app.RegistrationReader
}

var _ authserver.DataHolderClientStore = Clients{}

func (c Clients) Client(ctx context.Context, id string) (authserver.DataHolderClient, error) {
	registration, err := c.Reader.Registration(ctx, id)
	if errors.Is(err, app.ErrRegistrationNotFound) {
		return authserver.DataHolderClient{}, authserver.ErrClientNotFound
	}

	if err != nil {
		return authserver.DataHolderClient{}, err
	}

	grantTypes := make([]string, 0, len(registration.GrantTypes))
	for _, g := range registration.GrantTypes {
		grantTypes = append(grantTypes, string(g))
	}

	return authserver.DataHolderClient{
		ID:                      registration.ClientID,
		JWKSURI:                 registration.JWKSURI,
		Scopes:                  strings.Fields(registration.Scope),
		GrantTypes:              grantTypes,
		RedirectURIs:            registration.RedirectURIs,
		RequestObjectSigningAlg: string(registration.RequestObjectSigningAlg),
	}, nil
}
