package ipWhitelistRepository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ipWhitelist"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *IPWhitelistRepository) Update(ctx context.Context, configuration *ipwhitelistModel.IPWhitelistConfiguration) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ipWhitelist/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, IPWhitelistTable)
	query := `
		UPDATE ` + IPWhitelistTable + `
		SET 
			ip = :ip, 
			subnet = :subnet, 
			priority = :priority, 
			action = :action, 
			status = :status, 
			description = :description, 
			updated_at = :updated_at
		WHERE 
			id = :id AND deleted_at IS NULL
	`
	affected, err := r.db.NamedExecContext(ctx, query, configuration)
	if err != nil {
		r.logger.Error(ctx, "error when update ip whitelist configuration", logger.Error(err), logger.Any("configuration", configuration))
		return err
	}

	if !affected {
		r.logger.Info(ctx, "failed when update ip whitelist configuration", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}
	return nil
}
