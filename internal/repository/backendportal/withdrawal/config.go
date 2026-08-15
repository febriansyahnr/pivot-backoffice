package withdrawalRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *withdrawalRepository) GetTransactionConfig(ctx context.Context, merchantId string) (*merchant.WithdrawalConfig, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/GetTransactionConfig")
	defer segment.End()

	trxConfig := types.NullJSONText{}
	withdrawalConfig := merchant.WithdrawalConfig{}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "merchants") // NOSONAR

	// If the merchant is a non-KYC sub-merchant, it will use the parent merchant configuration.
	// Otherwise, it will use its own configuration.
	rawQuery := `
	SELECT
		IF(
			m.parent_id IS NOT NULL AND m.kyc_status = 'NOT_REQUIRED', parent.transaction_configs->>'$.withdrawal', m.transaction_configs->>'$.withdrawal'	
		) AS withdrawal
	FROM
		merchants m
	LEFT JOIN
		merchants parent ON parent.uuid = m.parent_id
	WHERE 
		m.uuid = ?;`

	if err := r.db.GetContext(ctx, &trxConfig, rawQuery, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
		}
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	// Handle as default value of environment variable
	if !trxConfig.Valid {
		return nil, nil
	}

	_ = trxConfig.Unmarshal(&withdrawalConfig)

	return &withdrawalConfig, nil
}
