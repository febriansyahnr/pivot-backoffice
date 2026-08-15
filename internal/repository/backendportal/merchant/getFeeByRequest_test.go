package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantFeeByRequest(t *testing.T) {

	testCases := []struct {
		Name      string
		Input     *merchant.GetMerchantFeeRequest
		MockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		WantErr   bool
	}{
		{
			Name: "SUCCESS:Get merchant fee for payment and method",
			Input: &merchant.GetMerchantFeeRequest{
				ID:            "id",
				MerchantID:    "3b6b2868-043a-4946-8e6e-5037dbdc8d65", // NOSONAR
				AmountType:    "AMOUNT",
				Reference:     constant.ReferencePayment,
				PaymentMethod: constant.ChannelVirtualAccount,
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS:Get merchant fee for payment method and channel",
			Input: &merchant.GetMerchantFeeRequest{
				ID:            "id",
				MerchantID:    "3b6b2868-043a-4946-8e6e-5037dbdc8d65",
				AmountType:    "AMOUNT",
				Reference:     constant.ReferencePayment,
				PaymentMethod: constant.ChannelVirtualAccount,
				Channel:       "OKE",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "Error sql no rows",
			Input: &merchant.GetMerchantFeeRequest{
				ID:         "id",
				MerchantID: "merchant-id",
				AmountType: "disbursement",
				Reference:  "reference",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)

			},
			WantErr: false,
		},

		{
			Name: "Error return other errors",
			Input: &merchant.GetMerchantFeeRequest{
				ID:         "id",
				MerchantID: "merchant-id",
				AmountType: "disbursement",
				Reference:  "reference",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("error"))

			},
			WantErr: true,
		},
		{
			Name: "SUCCESS:With ReferenceType",
			Input: &merchant.GetMerchantFeeRequest{
				ID:            "id",
				MerchantID:    "merchant-id",
				AmountType:    "disbursement",
				Reference:     "reference",
				ReferenceType: "QRIS_PAYMENT",
			},
			MockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrMerchantFeeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			WantErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			test.MockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetMerchantFeeByRequest(context.Background(), test.Input)

			if test.WantErr {
				assert.NotNil(t, err)
			}
		})
	}
}
