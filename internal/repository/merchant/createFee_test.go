package merchant

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRepositoryCreateMerchantFee(t *testing.T) {
	merchantId := uuid.NewString()

	merchant := &merchantModel.MerchantFee{
		UUID:       uuid.NewString(),
		MerchantID: merchantId,
		Amount:     1000,
		Reference:  "DISBURSEMENT",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	testCases := []struct {
		name      string
		merchant  *merchantModel.MerchantFee
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name:     "SUCCESS: create merchant fee",
			merchant: merchant,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.PtrMerchantFeeMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:     "FAIL: Failure Insert to Database",
			merchant: merchant,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.PtrMerchantFeeMockType(),
				).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.PtrMerchantFeeMockType(),
				).Return(false, nil)
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
			err := repo.CreateMerchantFee(context.Background(), tc.merchant)

			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.CreateMerchantFee() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
