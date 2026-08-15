package merchant

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"

	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) UpdateTransactionConfig(ctx context.Context, merchantId string, config *merchant.TransactionConfigs) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/UpdateTransactionConfigs")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	args := make([]any, 2)
	args[0], _ = json.Marshal(config.Disbursement)
	args[1], _ = json.Marshal(config.Withdrawal)

	rawQueryDailyDisbursement := ""
	if config.DailyDisbursement != nil {
		raw, _ := json.Marshal(config.DailyDisbursement)

		args = append(args, raw)
		rawQueryDailyDisbursement = ` transaction_configs = JSON_SET(transaction_configs, '$.dailyDisbursement', CAST(? AS JSON)),`
	}

	rawQuery := fmt.Sprintf(`UPDATE `+merchantsTable+` 
		SET 
			transaction_configs = JSON_SET(transaction_configs, '$.disbursement', CAST(? AS JSON), '$.withdrawal', CAST(? AS JSON)), %s updated_at = ?
		WHERE uuid = ? AND deleted_at IS NULL;`, rawQueryDailyDisbursement)

	args = append(args, time.Now().UTC(), merchantId)
	if _, err := r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		r.logger.Error(ctx, "Update transaction config", logger.Error(err))
		return err
	}
	return nil
}

func (r *MerchantRepository) UpdateFDSConfig(ctx context.Context, merchantID string, config merchant.FDSConfigRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/UpdateFDSConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	raw, _ := json.Marshal(config)

	rawQuery := `UPDATE merchants SET fds_configs = ?, updated_at = ? WHERE uuid = ?;`
	if affected, err := r.db.ExecContext(ctx, rawQuery, types.JSONText(raw), time.Now().UTC(), merchantID); err != nil {
		return err

	} else if !affected {
		return constant.ErrNoRowsAffected
	}
	return nil
}

func (r *MerchantRepository) UpdatePaymentInvestigationConfig(ctx context.Context, merchantID string, config merchant.PaymentInvestigationConfigRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/UpdatePaymentInvestigationConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	raw, _ := json.Marshal(config)

	rawQuery := `UPDATE
		merchants
	SET 
		transaction_configs = JSON_SET(transaction_configs, '$.paymentInvestigation', CAST(? AS JSON)), updated_at = ? 
	WHERE uuid = ?;`

	if _, err := r.db.ExecContext(ctx, rawQuery, types.JSONText(raw), time.Now().UTC(), merchantID); err != nil {
		return err
	}
	return nil
}

func (r *MerchantRepository) GetTransactionConfig(ctx context.Context, merchantId string) (*merchant.TransactionConfigResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetTransactionConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `
		SELECT
			m.name AS merchant_name,
			IF(
				m.parent_id IS NOT NULL AND m.kyc_status = 'NOT_REQUIRED', parent.transaction_configs->>'$.disbursement', m.transaction_configs->>'$.disbursement'	
			) AS disbursement,
			IF(
				m.parent_id IS NOT NULL AND m.kyc_status = 'NOT_REQUIRED', parent.transaction_configs->>'$.withdrawal', m.transaction_configs->>'$.withdrawal'	
			) AS withdrawal,
			m.transaction_configs->>'$.dailyDisbursement' AS daily_disbursement,
			CASE
				WHEN m.parent_id IS NOT NULL AND p.name IS NOT NULL THEN 'SUB_MERCHANT_PLATFORM'
				WHEN m.parent_id IS NOT NULL THEN CONCAT( IF( IFNULL(m.kyc_status, 'APPROVED') = 'NOT_REQUIRED', 'NON_KYC_', 'KYC_'), 'SUB_MERCHANT')
				WHEN p.name IS NOT NULL THEN 'MERCHANT_PLATFORM'
				ELSE 'MERCHANT'
			END AS merchant_type
		FROM merchants m
		LEFT JOIN merchants parent ON parent.uuid = m.parent_id
		LEFT JOIN merchant_selected_products msp ON msp.merchant_id = m.uuid AND msp.active = 1
		LEFT JOIN products p ON p.uuid = msp.product_id AND p.name = 'PLATFORM'
		WHERE m.uuid = ? AND m.deleted_at IS NULL LIMIT 1;`

	config := &merchant.RawTransactionConfig{}
	if err := r.db.GetContext(ctx, config, rawQuery, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := &merchant.TransactionConfigResp{
		MerchantId:   merchantId,
		MerchantName: config.MerchantName,
		MerchantType: config.MerchantType,
		TransactionConfigs: merchant.TransactionConfigs{
			Disbursement: merchant.DisbursementConfig{
				MinAmount: r.config.DisbursementConfig.MinAmount,
				MaxAmount: r.config.DisbursementConfig.MaxAmount,
			},
			Withdrawal: merchant.WithdrawalConfig{
				MinAmount: r.config.WithdrawalConfig.MinAmount,
				MaxAmount: r.config.WithdrawalConfig.MaxAmount,
			},
		},
	}
	if config.MerchantType != "NON_KYC_SUB_MERCHANT" {
		result.TransactionConfigs.DailyDisbursement = &merchant.DailyDisbursementConfig{
			Merchant: r.config.DisbursementConfig.DailyLimitMerchant,
		}
	}
	if strings.HasSuffix(config.MerchantType, "MERCHANT_PLATFORM") {
		result.TransactionConfigs.DailyDisbursement.MerchantPlatform = &r.config.DisbursementConfig.DailyLimitMerchantPlatform
	}
	_ = config.Withdrawal.Unmarshal(&result.TransactionConfigs.Withdrawal)
	_ = config.Disbursement.Unmarshal(&result.TransactionConfigs.Disbursement)
	_ = config.DailyDisbursement.Unmarshal(result.TransactionConfigs.DailyDisbursement)

	return result, nil
}

func (r *MerchantRepository) GetFDSConfig(ctx context.Context, merchantID string) (*merchant.GetFDSConfigResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetFDSConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "merchants,products,merchant_selected_products")

	rawQuery := `SELECT
		m.uuid AS merchant_id, m.name AS merchant_name,
		IF(m.parent_id IS NOT NULL AND m.kyc_status = 'NOT_REQUIRED', parent.fds_configs, m.fds_configs) AS fds_configs,
		CASE
			WHEN m.parent_id IS NOT NULL AND p.name IS NOT NULL THEN 'SUB_MERCHANT_PLATFORM'
			WHEN m.parent_id IS NOT NULL THEN CONCAT( IF( IFNULL(m.kyc_status, 'APPROVED') = 'NOT_REQUIRED', 'NON_KYC_', 'KYC_'), 'SUB_MERCHANT')
			WHEN p.name IS NOT NULL THEN 'MERCHANT_PLATFORM'
			ELSE 'MERCHANT'
		END AS merchant_type
	FROM merchants m
	LEFT JOIN merchants parent ON parent.uuid = m.parent_id
	LEFT JOIN merchant_selected_products msp ON msp.merchant_id = m.uuid AND msp.active = 1
	LEFT JOIN products p ON p.uuid = msp.product_id AND p.name = 'PLATFORM'
	WHERE m.uuid = ? AND m.deleted_at IS NULL LIMIT 1;`

	result := merchant.GetFDSConfigResponse{}
	if err := r.db.GetContext(ctx, &result, rawQuery, merchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if result.RawFDSConfig.Valid {
		_ = result.RawFDSConfig.Unmarshal(&result.FDSConfig)
	} else {
		result.FDSConfig = merchant.FDSConfig{}
	}

	// Set POP default if empty
	if result.FDSConfig.ProofOfPayment == nil {
		popConfig := r.config.FdsConfig.Features.ProofOfPayment
		result.FDSConfig.ProofOfPayment = &merchant.FDSFeatureProofOfPayment{
			Velocity: merchant.FDSRuleVelocityConfig{
				Enabled: popConfig.Velocity.Enabled,
				Window: merchant.FDSWindowConfig{
					Interval: popConfig.Velocity.Window.Interval,
					Unit:     popConfig.Velocity.Window.Unit,
				},
				Threshold: merchant.FDSThresholdConfig{
					Count: popConfig.Velocity.Threshold.Count,
				},
				Action: popConfig.Velocity.Action,
			},
		}
	}

	// Reset data because it is no longer used
	result.RawFDSConfig.Valid = false
	result.RawFDSConfig.JSONText = nil

	return &result, nil
}

func (r *MerchantRepository) GetDisbursementMerchantConfig(ctx context.Context, merchantId string) (*merchant.DisbursementMerchantConfig, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetDisbursementMerchantConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `SELECT
		IF(( parent_id IS NOT NULL AND kyc_status = 'APPROVED') OR parent_id IS NULL, uuid, parent_id) AS daily_limit_merchant_id, 
		IF(( parent_id IS NOT NULL AND kyc_status = 'APPROVED') OR parent_id IS NULL, 'merchant', 'merchant-platform') AS daily_limit_merchant_type
	FROM merchants 
	WHERE uuid = ? AND deleted_at IS NULL;`

	result := &merchant.DisbursementMerchantConfig{}
	return result, r.db.GetContext(ctx, result, rawQuery, merchantId)
}

func (r *MerchantRepository) IsInvestigationFlowEnabled(ctx context.Context, merchantID string) (enabled bool, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/IsInvestigationFlowEnabled")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `SELECT
		IFNULL(IF(
			m.parent_id IS NOT NULL AND m.kyc_status = 'NOT_REQUIRED', 
			parent.transaction_configs->>'$.paymentInvestigation.enabled', m.transaction_configs->>'$.paymentInvestigation.enabled'
		), FALSE) AS payment_investigation_enabled
	FROM merchants m
	LEFT JOIN merchants parent ON parent.uuid = m.parent_id
	WHERE m.uuid = ? AND m.deleted_at IS NULL LIMIT 1;`

	if err = r.db.GetContext(ctx, &enabled, rawQuery, merchantID); err != nil {
		return
	}
	return
}
