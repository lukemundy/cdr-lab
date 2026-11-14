// Package readmodel keeps the read models of the Register in PostgreSQL and queries them.
package readmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func Projections() []eventstore.Projection {
	// The ID projections are first, so that an ID that is not unique gives their error, and not a key error of a
	// different read model.
	return []eventstore.Projection{
		{
			StreamType: app.DataHolderType.StreamType,
			Project: projectParticipantIDs(app.DataHolderType.StreamType, func(
				streamID string, recorded []eventsourcing.RecordedEvent,
			) ([]string, error) {
				state, _, err := app.DataHolderType.Fold(streamID, recorded)
				if err != nil || !state.Imported {
					return nil, err
				}

				return state.DataHolder.IDs(), nil
			}),
		},
		{
			StreamType: app.DataRecipientType.StreamType,
			Project: projectParticipantIDs(app.DataRecipientType.StreamType, func(
				streamID string, recorded []eventsourcing.RecordedEvent,
			) ([]string, error) {
				state, _, err := app.DataRecipientType.Fold(streamID, recorded)
				if err != nil || !state.Imported {
					return nil, err
				}

				return state.DataRecipient.IDs(), nil
			}),
		},
		{StreamType: app.DataHolderType.StreamType, Project: projectDataHolder},
		{StreamType: app.DataRecipientType.StreamType, Project: projectDataRecipient},
		{
			StreamType: app.DataHolderType.StreamType,
			Project:    projectHistory(app.DataHolderType.StreamType, app.HistoryOfDataHolder),
		},
		{
			StreamType: app.DataRecipientType.StreamType,
			Project:    projectHistory(app.DataRecipientType.StreamType, app.HistoryOfDataRecipient),
		},
	}
}

func projectDataHolder(ctx context.Context, tx pgx.Tx, streamID string, recorded []eventsourcing.RecordedEvent) error {
	state, version, err := app.DataHolderType.Fold(streamID, recorded)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM data_holder_statuses WHERE legal_entity_id = $1", streamID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM data_holder_brands WHERE legal_entity_id = $1", streamID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM data_holders WHERE legal_entity_id = $1", streamID); err != nil {
		return err
	}

	if !state.Imported {
		return nil
	}

	dh := state.DataHolder
	lastUpdated := recorded[len(recorded)-1].RecordedAt

	// Not nil: pgx writes a nil slice as NULL, and the column is NOT NULL.
	brandNames := make([]string, 0, len(dh.Brands))
	for _, b := range dh.Brands {
		brandNames = append(brandNames, b.BrandName)
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO data_holders
			(legal_entity_id, legal_entity_name, status, brand_names, data_holder, last_updated, brand_count, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		dh.LegalEntityID,
		dh.LegalEntityName,
		dh.Status,
		brandNames,
		dh,
		lastUpdated,
		len(dh.Brands),
		version,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		ctx,
		"INSERT INTO data_holder_statuses (legal_entity_id, status, industries) VALUES ($1, $2, $3)",
		dh.LegalEntityID,
		dh.Status,
		industryStrings(dh.Industries()),
	); err != nil {
		return err
	}

	legalEntity := app.DataHolderLegalEntity{
		LegalEntityID: dh.LegalEntityID, LegalEntityName: dh.LegalEntityName, LogoURI: dh.LogoURI, Status: dh.Status,
	}

	columns := []string{
		"brand_id", "legal_entity_id", "status", "industries", "brand", "last_updated", "legal_entity",
		"data_holder_brand_id",
	}

	var brands [][]any
	for _, b := range dh.Brands {
		brands = append(
			brands,
			[]any{
				b.ID(), dh.LegalEntityID, b.Status, industryStrings(b.Industries), b, lastUpdated, legalEntity,
				registeredBrandID(b),
			},
		)
	}

	_, err = tx.CopyFrom(ctx, pgx.Identifier{"data_holder_brands"}, columns, pgx.CopyFromRows(brands))

	return err
}

func registeredBrandID(b domain.DataHolderBrand) *string {
	if b.DataHolderBrandID == "" {
		return nil
	}

	return &b.DataHolderBrandID
}

func industryStrings(industries []domain.Industry) []string {
	s := make([]string, 0, len(industries))
	for _, i := range industries {
		s = append(s, string(i))
	}

	return s
}

// projectDataRecipient replaces the rows of one data recipient from its
// state, so that it needs no code for each event type.
func projectDataRecipient(
	ctx context.Context,
	tx pgx.Tx,
	streamID string,
	recorded []eventsourcing.RecordedEvent,
) error {
	state, version, err := app.DataRecipientType.Fold(streamID, recorded)
	if err != nil {
		return err
	}
	// The software product rows cascade.
	if _, err := tx.Exec(ctx, "DELETE FROM data_recipient_statuses WHERE legal_entity_id = $1", streamID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM data_recipients WHERE legal_entity_id = $1", streamID); err != nil {
		return err
	}

	if !state.Imported {
		return nil
	}

	dr := state.DataRecipient
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO data_recipients (legal_entity_id, data_recipient, last_updated, legal_entity_name,
			accreditation_number, status, version, brand_count, software_product_count)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		dr.LegalEntityID,
		dr,
		recorded[len(recorded)-1].RecordedAt,
		dr.LegalEntityName,
		dr.AccreditationNumber,
		dr.Status,
		version,
		len(dr.Brands),
		dr.SoftwareProductCount(),
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "INSERT INTO data_recipient_statuses (legal_entity_id, status) VALUES ($1, $2)",
		dr.LegalEntityID, dr.Status); err != nil {
		return err
	}

	productColumns := []string{"software_product_id", "legal_entity_id", "status"}
	clientColumns := []string{"software_product_id", "legal_entity_id", "status", "jwks_uri", "client_certificates"}

	var products, clients [][]any

	for _, b := range dr.Brands {
		for _, sp := range b.SoftwareProducts {
			products = append(products, []any{sp.SoftwareProductID, dr.LegalEntityID, sp.Status})
			// Not nil: pgx writes a nil slice as NULL, and the column is NOT NULL.
			certificates := append([]string{}, sp.ClientCertificates...)
			clients = append(
				clients,
				[]any{sp.SoftwareProductID, dr.LegalEntityID, sp.Status, sp.JWKSURI, certificates},
			)
		}
	}

	if _, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"software_product_statuses"},
		productColumns,
		pgx.CopyFromRows(products),
	); err != nil {
		return err
	}

	_, err = tx.CopyFrom(ctx, pgx.Identifier{"software_product_clients"}, clientColumns, pgx.CopyFromRows(clients))

	return err
}

type Queries struct {
	pool *pgxpool.Pool
}

var (
	_ app.StatusReader                    = (*Queries)(nil)
	_ app.DataHolderBrandReader           = (*Queries)(nil)
	_ app.DataRecipientReader             = (*Queries)(nil)
	_ app.SoftwareProductClientReader     = (*Queries)(nil)
	_ app.SoftwareProductOwnerReader      = (*Queries)(nil)
	_ app.RegisteredDataHolderBrandReader = (*Queries)(nil)
)

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

func (q *Queries) DataHolderStatuses(
	ctx context.Context,
	industry *domain.Industry,
) ([]app.DataHolderStatusItem, error) {
	rows, err := q.pool.Query(
		ctx,
		`SELECT legal_entity_id, status FROM data_holder_statuses
			WHERE $1::text IS NULL OR $1 = ANY (industries) ORDER BY legal_entity_id`,
		(*string)(industry),
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByPos[app.DataHolderStatusItem])
}

func (q *Queries) DataHolderBrands(ctx context.Context, industry *domain.Industry) ([]app.DataHolderBrandItem, error) {
	rows, err := q.pool.Query(
		ctx,
		`SELECT brand, last_updated FROM data_holder_brands
			WHERE status <> $1 AND ($2::text IS NULL OR $2 = ANY (industries)) ORDER BY brand_id`,
		domain.Removed,
		(*string)(industry),
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByPos[app.DataHolderBrandItem])
}

func (q *Queries) DataRecipientStatuses(ctx context.Context) ([]app.DataRecipientStatusItem, error) {
	rows, err := q.pool.Query(
		ctx,
		"SELECT legal_entity_id, status FROM data_recipient_statuses ORDER BY legal_entity_id",
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByPos[app.DataRecipientStatusItem])
}

func (q *Queries) SoftwareProductStatuses(ctx context.Context) ([]app.SoftwareProductStatusItem, error) {
	rows, err := q.pool.Query(
		ctx,
		"SELECT software_product_id, status FROM software_product_statuses ORDER BY software_product_id",
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByPos[app.SoftwareProductStatusItem])
}

func (q *Queries) DataRecipients(ctx context.Context) ([]app.DataRecipientItem, error) {
	rows, err := q.pool.Query(
		ctx,
		"SELECT data_recipient, last_updated, version FROM data_recipients ORDER BY legal_entity_id",
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByPos[app.DataRecipientItem])
}

func (q *Queries) SoftwareProductClient(
	ctx context.Context,
	softwareProductID string,
) (app.SoftwareProductClient, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT c.software_product_id, c.status, r.status, c.jwks_uri, c.client_certificates
		FROM software_product_clients c JOIN data_recipient_statuses r USING (legal_entity_id)
		WHERE c.software_product_id = $1`,
		softwareProductID)
	if err != nil {
		return app.SoftwareProductClient{}, err
	}

	client, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[app.SoftwareProductClient])
	if errors.Is(err, pgx.ErrNoRows) {
		return app.SoftwareProductClient{}, fmt.Errorf("software product %w", domain.ErrNotFound)
	}

	return client, err
}

const registeredBrands = `FROM data_holder_brands
	WHERE data_holder_brand_id IS NOT NULL AND ($1::text IS NULL OR $1 = ANY (industries))
	AND ($2::timestamptz IS NULL OR last_updated >= $2)`

func (q *Queries) RegisteredDataHolderBrands(
	ctx context.Context,
	query app.DataHolderBrandQuery,
) ([]app.RegisteredDataHolderBrandItem, int, error) {
	var total int
	if err := q.pool.QueryRow(ctx, "SELECT count(*) "+registeredBrands, (*string)(query.Industry), query.UpdatedSince).
		Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := q.pool.Query(
		ctx,
		"SELECT brand, legal_entity, last_updated "+registeredBrands+" ORDER BY brand_id LIMIT $3 OFFSET $4",
		(*string)(query.Industry),
		query.UpdatedSince,
		query.PageSize,
		(query.Page-1)*query.PageSize,
	)
	if err != nil {
		return nil, 0, err
	}

	page, err := pgx.CollectRows(rows, pgx.RowToStructByPos[app.RegisteredDataHolderBrandItem])

	return page, total, err
}

func (q *Queries) DataRecipientOfSoftwareProduct(
	ctx context.Context,
	softwareProductID string,
) (domain.DataRecipient, error) {
	rows, err := q.pool.Query(
		ctx,
		`SELECT r.data_recipient FROM data_recipients r JOIN software_product_clients c USING (legal_entity_id)
			WHERE c.software_product_id = $1`,
		softwareProductID,
	)
	if err != nil {
		return domain.DataRecipient{}, err
	}

	dr, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[domain.DataRecipient])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DataRecipient{}, fmt.Errorf("software product %w", domain.ErrNotFound)
	}

	return dr, err
}
