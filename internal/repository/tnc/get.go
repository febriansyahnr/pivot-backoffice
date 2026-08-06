package tnc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *TNCRepository) GetTNCVersionByID(ctx context.Context, id string) (*tncModel.TNC, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/GetTNCVersionByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, versionsTableName)

	query := fmt.Sprintf(`SELECT %s FROM %s t WHERE t.uuid = ? AND t.deleted_at IS NULL`, versionsTableColumns, versionsTableName)

	var version tncModel.TNC
	err := r.db.GetContext(ctx, &version, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when getting tnc version by id", logger.Error(err), logger.String("id", id))
		return nil, err
	}

	return &version, nil
}

func (r *TNCRepository) GetTNCVersionByVersion(ctx context.Context, version string) (*tncModel.TNC, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/GetTNCVersionByVersion")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, versionsTableName)

	query := fmt.Sprintf(`SELECT %s FROM %s t WHERE t.version = ? AND t.deleted_at IS NULL ORDER BY t.updated_at DESC LIMIT 1`, versionsTableColumns, versionsTableName)

	var tnc tncModel.TNC
	err := r.db.GetContext(ctx, &tnc, query, version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when getting tnc version by version", logger.Error(err), logger.String("version", version))
		return nil, err
	}

	return &tnc, nil
}

func (r *TNCRepository) ListTNCVersions(
	ctx context.Context,
	q *tncModel.TNCVersionQuery,
) ([]*tncModel.TNC, int, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/ListTNCVersions")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, versionsTableName)

	whereClause, args := q.BuildCondition()
	if whereClause != "" {
		whereClause = "WHERE " + whereClause
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(t.uuid) FROM %s t %s`, versionsTableName, whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when counting tnc versions", logger.Error(err))
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.PageSize
	orderBy := q.BuildOrderBy()
	dataQuery := fmt.Sprintf(`SELECT %s FROM %s t %s ORDER BY %s LIMIT ? OFFSET ?`, versionsTableColumns, versionsTableName, whereClause, orderBy)

	args = append(args, q.PageSize, offset)

	var versions []*tncModel.TNC
	err = r.db.SelectContext(ctx, &versions, dataQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when listing tnc versions", logger.Error(err))
		return nil, 0, err
	}

	return versions, total, nil
}

func (r *TNCRepository) GetActiveTNCVersion(ctx context.Context) (*tncModel.TNC, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/GetActiveTNCVersion")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, versionsTableName)

	query := fmt.Sprintf(`SELECT %s FROM %s t WHERE t.is_active = 1 AND t.deleted_at IS NULL ORDER BY t.updated_at DESC LIMIT 1`, versionsTableColumns, versionsTableName)

	var version tncModel.TNC
	err := r.db.GetContext(ctx, &version, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when getting active tnc version", logger.Error(err))
		return nil, err
	}

	return &version, nil
}

func (r *TNCRepository) GetLatestSigningByMerchant(ctx context.Context, merchantID string) (*tncModel.MerchantTNCSigningHistory, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/GetLatestSigningByMerchant")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, historyTableName)

	query := fmt.Sprintf(`SELECT %s FROM %s t WHERE t.merchant_id = ? ORDER BY t.signed_at DESC LIMIT 1`, historyTableColumns, historyTableName)

	var history tncModel.MerchantTNCSigningHistory
	err := r.db.GetContext(ctx, &history, query, merchantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when getting latest tnc signing by merchant", logger.Error(err), logger.String("merchantID", merchantID))
		return nil, err
	}

	return &history, nil
}

func (r *TNCRepository) GetSigningByMerchantAndVersion(ctx context.Context, merchantID, version string) (*tncModel.MerchantTNCSigningHistory, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/GetSigningByMerchantAndVersion")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, historyTableName)

	query := fmt.Sprintf(`SELECT %s FROM %s t WHERE t.merchant_id = ? AND t.version = ? ORDER BY t.signed_at DESC LIMIT 1`, historyTableColumns, historyTableName)

	var history tncModel.MerchantTNCSigningHistory
	err := r.db.GetContext(ctx, &history, query, merchantID, version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when getting tnc signing by merchant and version", logger.Error(err), logger.String("merchantID", merchantID), logger.String("version", version))
		return nil, err
	}

	return &history, nil
}

func (r *TNCRepository) ListSigningHistories(
	ctx context.Context,
	q *tncModel.SigningHistoryQuery,
) ([]*tncModel.MerchantTNCSigningHistory, int, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/ListSigningHistories")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, historyTableName)

	whereClause, args := q.BuildCondition()
	if whereClause != "" {
		whereClause = "WHERE " + whereClause
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(t.uuid) FROM %s t %s`, historyTableName, whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when counting tnc signing histories", logger.Error(err))
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.PageSize
	dataQuery := fmt.Sprintf(`SELECT %s FROM %s t %s ORDER BY t.signed_at DESC LIMIT ? OFFSET ?`, historyTableColumns, historyTableName, whereClause)

	args = append(args, q.PageSize, offset)

	var histories []*tncModel.MerchantTNCSigningHistory
	err = r.db.SelectContext(ctx, &histories, dataQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when listing tnc signing histories", logger.Error(err))
		return nil, 0, err
	}

	return histories, total, nil
}
