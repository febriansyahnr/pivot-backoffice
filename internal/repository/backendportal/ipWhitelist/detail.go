package ipWhitelistRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ipWhitelist"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *IPWhitelistRepository) Detail(ctx context.Context, uuid string) (*ipwhitelistModel.IPWhitelistConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ipWhitelist/Detail")
	defer segment.End()

	var configuration ipwhitelistModel.IPWhitelistConfiguration
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, IPWhitelistTable)
	query := `
		SELECT id, merchant_id, ip, subnet, priority, action, status, description, created_at, updated_at, deleted_at
		FROM ` + IPWhitelistTable + `
		WHERE 
			id = ? and deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &configuration, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding ip whitelist", logger.Error(err), logger.Any("uuid", uuid))
		return nil, err
	}
	return &configuration, err
}
