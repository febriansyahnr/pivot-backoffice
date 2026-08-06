package vccSettlement

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *VccSettlementRepository) Delete(ctx context.Context, rcnId string, postingDate time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/vccSettlement/Delete")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := `
		UPDATE ` + tableName + `
		SET deleted_at = ?
		WHERE posting_date = ? AND rcn_id = ? AND deleted_at IS NULL
		`
	_, err := r.db.ExecContext(ctx, query, time.Now(), postingDate, rcnId)
	if err != nil {
		r.logger.Error(ctx, "error when delete vcc settlements records", logger.Error(err), logger.String("rcnId", rcnId), logger.String("postingDate", postingDate.String()))
		return err
	}
	return nil
}
