package mySqlExt

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func (m *mySqlExt) ExecContextReturnLastId(
	ctx context.Context,
	query string,
	args ...interface{},
) (*sql.Result, error) {
	var (
		start = time.Now()

		span     trace.Span
		duration time.Duration

		sqlResults sql.Result
		err        error
	)
	ctx, span = m.OtelTracer().Start(ctx, "MySqlExt.GetContext")

	defer func(ctx context.Context, query string, tableName string, duration *time.Duration) {
		m.InstrumentMetric(ctx, query, tableName, duration)
		span.SetAttributes(
			semconv.DBSystemMySQL,
			semconv.DBNameKey.String(m.DBName()),
			semconv.DBSQLTableKey.String(tableName),
			semconv.DBStatementKey.String(query),
			attribute.Float64("duration", duration.Seconds()),
		)
		span.End()
	}(ctx, query, m.TableName(ctx), &duration)

	// Check if there is an active transaction
	// If there is, use the transaction
	// Otherwise, use the masterDB because it is a write operation
	tx := m.GetTx(ctx)
	if tx != nil {
		sqlResults, err = tx.ExecContext(ctx, query, args...)
	} else {
		sqlResults, err = m.MasterDBClient().ExecContext(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}

	duration = time.Since(start)

	return &sqlResults, err
}

func (m *mySqlExt) Rebind(query string) string {
	return m.MasterDBClient().Rebind(query)
}

func (m *mySqlExt) In(rawQuery string, args ...interface{}) (string, []interface{}, error) {
	return sqlx.In(rawQuery, args...)
}
