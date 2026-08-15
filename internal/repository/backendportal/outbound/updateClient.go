package outbound

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/outbound"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *repository) UpdateClient(ctx context.Context, id string, data *outbound.Client) error {
	ctx, span := tracer.Start(ctx, "internal/repository/outbound/UpdateClient")
	defer span.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "outbound")

	q := `
	UPDATE outbound
	SET client = ?
	WHERE id = ?
	`
	dataB, _ := json.Marshal(data)
	_, err := r.db.ExecContext(ctx, q, dataB, id)
	if err != nil {
		return err
	}

	return nil
}
