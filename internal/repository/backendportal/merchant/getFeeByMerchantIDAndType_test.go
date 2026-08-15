package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRepositoryGetMerchantFeeByMerchantIDAndType(t *testing.T) {
	merchantId := uuid.NewString()
	feeType := "disbursement"

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get merchant fee by merchant ID and type",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					merchantFee := args.Get(1).(*merchant.MerchantFee)
					*merchantFee = merchant.MerchantFee{
						UUID:       "merchant-id",
						MerchantID: merchantId,
						Amount:     1000,
						Reference:  feeType,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Merchant fee not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("invalid-query"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetMerchantFeeByMerchantIDAndType(context.Background(), merchantId, feeType)
			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.GetMerchantFeeByMerchantIDAndType() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
