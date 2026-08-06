package merchant

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantRepository) GetDepositSetting(ctx context.Context, merchantId string) (result *merchant.DepositSettingResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetDepositSetting")
	defer segment.End()

	result = &merchant.DepositSettingResponse{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `SELECT 
			name AS merchant_name, IFNULL(transaction_configs->>'$.autoWithdrawal', 'OFF') AS auto_withdrawal 
		FROM merchants WHERE uuid = ?;`

	if err = r.db.GetContext(ctx, result, rawQuery, merchantId); err != nil {
		return nil, err
	}
	return
}

func (r *MerchantRepository) SetAutoWithdrawal(ctx context.Context, request *merchant.AutoWithdrawalSettingRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/SetAutoWithdrawal")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantsTable)

	rawQuery := `UPDATE
			merchants
		SET transaction_configs = JSON_SET(transaction_configs, '$.autoWithdrawal', ?), updated_at = ? WHERE uuid = ?;`

	_, err := r.db.ExecContext(ctx, rawQuery, request.Status, time.Now().UTC(), request.MerchantId)
	return err
}
