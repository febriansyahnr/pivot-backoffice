package merchant_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/merchant"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMerchantFeeXB(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	merchantID := uuid.NewString()
	reference := "PAYMENT"
	channel := "QRIS"

	settlementJSON := types.NullJSONText{
		JSONText: []byte(`{"type":"daily"}`),
		Valid:    true,
	}

	ptrMerchantFeeMockType := mock.AnythingOfType("*merchant.MerchantFee")

	tests := []struct {
		name       string
		query      *merchant.MerchantFeeXBQuery
		setupMock  func(db *mySqlExtMock.IMySqlExt)
		wantErr    bool
		wantResult *merchant.MerchantFee
	}{
		{
			name: "SUCCESS: Data found",
			query: &merchant.MerchantFeeXBQuery{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    channel,
			},
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrMerchantFeeMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					fee := args.Get(1).(*merchant.MerchantFee)
					fee.UUID = uuid.NewString()
					fee.MerchantID = merchantID
					fee.Reference = reference
					fee.AmountType = "FLAT"
					fee.Amount = 1000.0
					fee.Percentage = 0.0
					fee.DeductionType = "DEDUCT_FROM_PAYMENT"
					fee.TaxType = "INCLUSIVE"
					fee.TaxPercentage = 11.0
					fee.SettlementConfigs = settlementJSON
				}).Return(nil)
			},
			wantErr: false,
			wantResult: &merchant.MerchantFee{
				MerchantID:        merchantID,
				Reference:         reference,
				AmountType:        "FLAT",
				Amount:            1000.0,
				Percentage:        0.0,
				DeductionType:     "DEDUCT_FROM_PAYMENT",
				TaxType:           "INCLUSIVE",
				TaxPercentage:     11.0,
				SettlementConfigs: settlementJSON,
			},
		},
		{
			name: "SUCCESS: Data not found (sql.ErrNoRows)",
			query: &merchant.MerchantFeeXBQuery{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    channel,
			},
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrMerchantFeeMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			wantErr:    false,
			wantResult: nil,
		},
		{
			name: "ERROR: Database error",
			query: &merchant.MerchantFeeXBQuery{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    channel,
			},
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrMerchantFeeMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:    true,
			wantResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := mySqlExtMock.NewIMySqlExt(t)
			test.setupMock(db)

			repo := New(db, mockLogger)
			result, err := repo.GetMerchantFeeXB(context.Background(), test.query)

			if test.wantErr {
				require.Error(t, err)
				assert.NotNil(t, result)
			} else {
				require.NoError(t, err)
				if test.wantResult == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, test.wantResult.MerchantID, result.MerchantID)
					assert.Equal(t, test.wantResult.Reference, result.Reference)
					assert.Equal(t, test.wantResult.AmountType, result.AmountType)
					assert.Equal(t, test.wantResult.Amount, result.Amount)
				}
			}
		})
	}
}
