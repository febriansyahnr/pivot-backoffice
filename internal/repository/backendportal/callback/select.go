package callbackRepository

import (
	"context"
	"database/sql"
	"errors"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *CallbackRepository) FindCallbackMasterByName(
	ctx context.Context,
	name string,
) (*callbackModel.CallbackMaster, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/FindCallbackMasterByName")
	defer segment.End()

	var callbackMaster callbackModel.CallbackMaster

	query := `
		SELECT 
			uuid, 
			name, 
			description, 
			created_at, 
			updated_at, 
			deleted_at
		FROM callback_masters
		WHERE name = ?
		ORDER BY created_at DESC LIMIT 1`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "callback_masters")

	if err := r.db.GetContext(ctx, &callbackMaster, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn(ctx, "callback_masters not found", logger.String("name", name))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding callback_masters", logger.Error(err))
		return &callbackMaster, err
	}

	return &callbackMaster, nil
}

func (r *CallbackRepository) FindCallbackByNameAndMerchantID(
	ctx context.Context,
	name string,
	merchantId uuid.UUID,
) (*callbackModel.Callback, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/FindCallbackByNameAndMerchantID")
	defer segment.End()

	var callback callbackModel.Callback

	query := `
		SELECT 
			cb.uuid,
			cb.callback_master_id,
			cb.merchant_id,
			cb.base_url,
			cb.url,
			cb.description,
			cb.created_at,
			cb.updated_at,
			cb.deleted_at
		FROM callbacks cb
		INNER JOIN callback_masters cm ON cb.callback_master_id = cm.uuid
		WHERE cm.name = ?
		AND cb.merchant_id = ?
		ORDER BY cb.created_at DESC LIMIT 1`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "callbacks")

	if err := r.db.GetContext(ctx, &callback, query, name, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding callbacks", logger.Error(err))
		return &callback, err
	}

	return &callback, nil
}

func (r *CallbackRepository) GetCallbackURLByMerchantId(ctx context.Context, merchantID string, masterName string) (resp []callbackModel.CallbackURLSettingResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/GetCallbackURLByMerchantId")
	defer segment.End()
	resp = []callbackModel.CallbackURLSettingResp{}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "callbacks, callback_masters")

	rawQuery := `
		SELECT 
			cm.uuid AS master_id, cm.name AS master_name, c.uuid AS callback_id, c.base_url AS callback_base_url, c.url AS callback_url, c.updated_at 
		FROM callback_masters cm
		LEFT JOIN callbacks c ON c.callback_master_id = cm.uuid AND c.merchant_id = ? AND c.deleted_at IS NULL
		WHERE cm.deleted_at IS NULL`

	// Add the masterName filter
	var queryParams []interface{}
	queryParams = append(queryParams, merchantID)

	if masterName != "" {
		rawQuery += " AND cm.name LIKE ?"
		queryParams = append(queryParams, "%"+masterName+"%")
	} else {
		rawQuery += " AND cm.name NOT LIKE ?"
		queryParams = append(queryParams, "%SNAP%")
	}

	// Add visibility condition
	rawQuery += ` AND (
			cm.visibility = 'PUBLIC' OR
			(cm.visibility = 'RESTRICTED' AND JSON_CONTAINS(cm.whitelisted_merchant_ids, JSON_ARRAY(?)))
		)`
	queryParams = append(queryParams, merchantID)

	// Add query ordering
	rawQuery += " ORDER BY cm.name"

	err = r.db.SelectContext(ctx, &resp, rawQuery, queryParams...)
	return
}

func (r *CallbackRepository) GetCallbackAPIKeyByMerchantId(ctx context.Context, merchantID string) (result *callbackModel.CallbackAPIKeyResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/GetCallbackAPIKeyByMerchantId")
	defer segment.End()

	ctx, result = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "merchants"), &callbackModel.CallbackAPIKeyResp{}

	rawQuery := "SELECT callback_api_key, callback_api_key_version FROM merchants WHERE uuid = ? AND deleted_at IS NULL;"

	err = r.db.GetContext(ctx, result, rawQuery, merchantID)
	return
}

func (r *CallbackRepository) GetCallbackIdByMerchantAndMasterCallbackId(ctx context.Context, merchantID, masterID string) (id string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/callback/GetCallbackIdByMerchantAndMasterCallbackId")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, TableCallback)
	rawQuery := "SELECT IFNULL(MAX(uuid), '') FROM " + TableCallback + " WHERE merchant_id = ? AND callback_master_id = ?"

	err = r.db.GetContext(ctx, &id, rawQuery, merchantID, masterID)
	return
}
