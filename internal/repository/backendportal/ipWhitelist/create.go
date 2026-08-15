package ipWhitelistRepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ipWhitelist"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *IPWhitelistRepository) Create(ctx context.Context, configuration *ipwhitelistModel.IPWhitelistConfiguration) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ipWhitelist/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, IPWhitelistTable)
	query := `
		INSERT INTO ` + IPWhitelistTable + `
			(id, merchant_id, ip, subnet, priority, action, status, description, created_at, updated_at) 
		VALUES (:id, :merchant_id, :ip, :subnet, :priority, :action, :status, :description, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, configuration)
	if err != nil {
		r.logger.Error(ctx, "error when inserting ip whitelist configuration", logger.Error(err), logger.Any("configuration", configuration))
		return err
	}
	return nil

}
