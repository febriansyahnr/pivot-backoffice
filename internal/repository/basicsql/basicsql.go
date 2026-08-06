package basicsql

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PropertiesRepository")

type Properties struct {
	db mySqlExt.IMySqlExt
}

func NewBasicSQLProperties(db mySqlExt.IMySqlExt) *Properties {
	return &Properties{db}
}

func (r *Properties) BeginTransaction(ctx context.Context) (context.Context, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/basicsql/BeginTransaction")
	defer segment.End()

	ctxTx, err := r.db.BeginTxx(ctx)
	if err != nil {
		return nil, err
	}
	return ctxTx, nil
}

func (r *Properties) CommitTransaction(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/basicsql/CommitTransaction")
	defer segment.End()

	return r.db.Commit(ctx)
}

func (r *Properties) RollbackTransaction(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/basicsql/RollbackTransaction")
	defer segment.End()

	return r.db.Rollback(ctx)
}
