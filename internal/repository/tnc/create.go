package tnc

import (
	"context"

	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *TNCRepository) CreateTNCVersion(ctx context.Context, version *tncModel.TNC) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/CreateTNCVersion")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, versionsTableName)

	query := `
		INSERT INTO tncs (
			uuid, version, title, markdown_content, is_active, created_by
		) VALUES (
			:uuid, :version, :title, :markdown_content, :is_active, :created_by
		)`

	_, err := r.db.NamedExecContext(ctx, query, version)
	if err != nil {
		r.logger.Error(ctx, "error when creating tnc version", logger.Error(err), logger.String("version", version.Version))
		return err
	}

	return nil
}

func (r *TNCRepository) InsertSigningHistory(ctx context.Context, history *tncModel.MerchantTNCSigningHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/InsertSigningHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, historyTableName)

	query := `
		INSERT INTO merchant_tnc_signing_histories (
			uuid, merchant_id, tnc_id, version, signed_by, signed_by_email, signed_at, document_path, ip_address, user_agent
		) VALUES (
			:uuid, :merchant_id, :tnc_id, :version, :signed_by, :signed_by_email, :signed_at, :document_path, :ip_address, :user_agent
		)`

	_, err := r.db.NamedExecContext(ctx, query, history)
	if err != nil {
		r.logger.Error(ctx, "error when inserting tnc signing history", logger.Error(err))
		return err
	}

	return nil
}
