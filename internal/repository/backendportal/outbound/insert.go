package outbound

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

// This process does not use the “New Relic Span” because it run behind the scenes (Goroutine)
func (r *repository) Insert(ctx context.Context, data *outbound.OutboundRequest) error {

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "outbound")

	rawQuery := `INSERT INTO outbound
			(id, client, date, method, url, headers, body, status_code, response_time, response_body, error_message)
		VALUES(:id, :client, :date, :method, :url, :headers, :body, :status_code, :response_time, :response_body, :error_message)`

	_, err := r.db.NamedExecContext(ctx, rawQuery, data.ToOutbound())
	return err
}
