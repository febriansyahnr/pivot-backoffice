package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *MerchantRepository) GetSettlementConfig(ctx context.Context, request model.GetSettlementConfigRequest) (*model.SettlementConfig, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetSettlementConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, merchantFeeTable)

	var whereClause1, whereClause2 string
	args := []any{
		request.MerchantId, request.Reference, request.Method, request.Channel,
	}
	if request.SettlementMethod != "" {
		whereClause1 = "merchant_id = ? AND reference = ? AND payment_method = ? AND channel = ? AND settlement_method = ? AND settlement_configs IS NOT NULL"
		whereClause2 = "merchant_id = ? AND reference = ? AND payment_method = ? AND channel IS NULL AND settlement_method = ? AND settlement_configs IS NOT NULL"

		args = append(args, request.SettlementMethod, request.MerchantId, request.Reference, request.Method, request.SettlementMethod)
	} else {
		whereClause1 = "merchant_id = ? AND reference = ? AND payment_method = ? AND channel = ? AND settlement_configs IS NOT NULL"
		whereClause2 = "merchant_id = ? AND reference = ? AND payment_method = ? AND channel IS NULL AND settlement_configs IS NOT NULL"

		args = append(args, request.MerchantId, request.Reference, request.Method)
	}

	rawQuery := `
		SELECT 
			settlement_configs
		FROM (
			SELECT
				0 AS priority, settlement_configs
			FROM
				merchant_fees
			WHERE ` + whereClause1 + `
			UNION ALL
			SELECT 
				1 AS priority, settlement_configs
			FROM
				merchant_fees
			WHERE ` + whereClause2 + `
		) foo
		ORDER BY priority LIMIT 1;`

	dst := types.JSONText{}
	if err := r.db.GetContext(ctx, &dst, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get settlement config: %v", err)
	}

	config := model.SettlementConfig{}
	if err := dst.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshal json: %v", err)
	}
	return &config, nil
}
