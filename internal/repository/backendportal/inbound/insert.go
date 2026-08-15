package inboundRepository

import (
	"context"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/inbound"
)

// This process does not use the “New Relic Span” because it run behind the scenes (Goroutine)
func (r *repository) Insert(ctx context.Context, data *inboundModel.InboundRequest) error {
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "inbound")

	rawQuery := `INSERT INTO inbound
			(id, ip, client, method, url, headers, body, status_code, response_time_ms, response_body, metadata, snap_compatibility, created_at, updated_at)
		VALUES(:id, :ip, :client, :method, :url, :headers, :body, :status_code, :response_time_ms, :response_body, :metadata, :snap_compatibility, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, rawQuery, data.ToInbound())
	return err
}
