package merchant

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

type MerchantWithSubMerchantList struct {
	ID              string         `db:"id"`
	CreatedAt       time.Time      `db:"created_at"`
	RawSubMerchants types.JSONText `db:"sub_merchants"`
	SubMerchants    []string       `db:"-"`
}

type MerchantFeeForBalanceDeduction struct {
	MerchantId     string     `db:"merchant_id"`
	Reference      string     `db:"reference"`
	Method         string     `db:"method"`
	DeductionDay   int        `db:"deduction_day"`
	LastDeductDate *time.Time `db:"deduction_last_date"`
	CreatedAt      time.Time  `db:"created_at"`
}

type MerchantFeeThatUseTier struct {
	Id                string             `db:"uuid"`
	MerchantId        string             `db:"merchant_id"`
	Reference         string             `db:"reference"`
	PaymentMethod     *string            `db:"payment_method"`
	TieringType       string             `db:"tiering_type"`
	RawTieringConfigs types.JSONText     `db:"tiering_configs"`
	TieringConfigs    []FeeTieringConfig `db:"-"`
}
