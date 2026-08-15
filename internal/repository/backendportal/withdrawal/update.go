package withdrawalRepository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/withdrawal"

	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *withdrawalRepository) UpdateMetadataById(ctx context.Context, id string, metadata *withdrawal.Metadata) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/UpdateMetadataById")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawMetadata, _ := json.Marshal(metadata)

	affected, err := r.db.ExecContext(
		ctx, `UPDATE withdrawals SET metadata = ?, updated_at = ? WHERE id = ?;`, types.JSONText(rawMetadata), time.Now().UTC(), id,
	)
	if err != nil {
		return err

	} else if !affected {
		return constant.ErrNoRowsAffected
	}
	return nil
}
