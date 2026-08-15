package paymentRepository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
)

func TestGetList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *paymentModel.GetListFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter:  &paymentModel.GetListFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &paymentModel.GetListFilterRequest{
				UUID:           uuid.NewString(),
				MerchantID:     uuid.NewString(),
				Status:         constant.UnifiedPaymentSessionStatusPaid,
				ReferenceID:    "ref-001",
				PaymentMethod:  constant.UnifiedPaymentMethodQris,
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("no rows data"))
			},
			filter:  &paymentModel.GetListFilterRequest{},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("some-error"))

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

			},
			filter:  &paymentModel.GetListFilterRequest{},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get List with Sort and SortBy",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &paymentModel.GetListFilterRequest{
				Sort:   "ASC",
				SortBy: "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with data to trigger BuildRespData loop",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					// Populate the data slice to trigger the loop in BuildRespData
					dataPtr := args.Get(1).(*[]*paymentModel.PaymentWithPaymentMethodDTO)
					*dataPtr = []*paymentModel.PaymentWithPaymentMethodDTO{
						{
							UUID:       uuid.NewString(),
							MerchantID: uuid.NewString(),
						},
					}
				})

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter:  &paymentModel.GetListFilterRequest{},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetList(ctx, tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
