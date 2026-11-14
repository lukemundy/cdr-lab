package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var errRedacted = errors.New("database error: message removed because it can contain query arguments")

// redactingTracer removes query error messages from spans, because they can contain argument values.
type redactingTracer struct{ *otelpgx.Tracer }

func (t redactingTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	data.Err = redact(data.Err)
	t.Tracer.TraceQueryEnd(ctx, conn, data)
}

func (t redactingTracer) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	data.Err = redact(data.Err)
	t.Tracer.TraceBatchQuery(ctx, conn, data)
}

func (t redactingTracer) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {
	data.Err = redact(data.Err)
	t.Tracer.TraceBatchEnd(ctx, conn, data)
}

func (t redactingTracer) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
	data.Err = redact(data.Err)
	t.Tracer.TraceCopyFromEnd(ctx, conn, data)
}

// redact keeps only the error fields that contain no data.
func redact(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil, errors.Is(err, sql.ErrNoRows):
		return err
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	case errors.As(err, &pgErr):
		return &pgconn.PgError{
			Severity:            pgErr.Severity,
			SeverityUnlocalized: pgErr.SeverityUnlocalized,
			Code:                pgErr.Code,
			Message:             errRedacted.Error(),
			SchemaName:          pgErr.SchemaName,
			TableName:           pgErr.TableName,
			ColumnName:          pgErr.ColumnName,
			DataTypeName:        pgErr.DataTypeName,
			ConstraintName:      pgErr.ConstraintName,
		}
	default:
		return errRedacted
	}
}
