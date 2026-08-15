package paymentMethodRepository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetListPaymentMethodMerchant(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get list payment method merchant",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error on get list payment method merchant",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get list with Status Active",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					data := args.Get(1).(*[]*paymentModel.PaymentMethodWithPivot)
					*data = []*paymentModel.PaymentMethodWithPivot{
						{
							PaymentMethod: paymentModel.PaymentMethod{
								UUID:     uuid.NewString(),
								Category: paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
								Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
								Name:     "VA Permata",
								Acquirer: constant.BANK_ACQUIRER_PERMATA,
								RequiredDocuments: types.NullJSONText{
									JSONText: []byte(`[{"name":"document1"}]`),
									Valid:    true,
								},
								CreatedAt: util.TimeNow,
								UpdatedAt: util.TimeNow,
							},
							MerchantConfig: types.NullJSONText{
								JSONText: []byte(`{"channelConfig":{"key":"value"}}`),
								Valid:    true,
							},
							IsActive: true,
						},
					}
				})
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get list with Status Inactive",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get list with derived merchant context",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					data := args.Get(1).(*[]*paymentModel.PaymentMethodWithPivot)
					*data = []*paymentModel.PaymentMethodWithPivot{
						{
							PaymentMethod: paymentModel.PaymentMethod{
								UUID:      uuid.NewString(),
								Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
								Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
								Name:      "VA Permata",
								Acquirer:  constant.BANK_ACQUIRER_PERMATA,
								CreatedAt: util.TimeNow,
								UpdatedAt: util.TimeNow,
							},
							IsActive: true,
						},
					}
				})
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tablePaymentMethodMerchant)

			// Add derived merchant context for specific test case
			if tc.name == "SUCCESS: Get list with derived merchant context" {
				ctx = context.WithValue(ctx, constant.CtxDerivedMerchantID, "derived-merchant-123")
			}

			filter := &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: uuid.NewString(),
				Category:   paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
				Type:       paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				Acquirer:   constant.BANK_ACQUIRER_PERMATA,
				Subtype:    "SUBTYPE",
				InstallmentPlan: paymentModel.InstallmentPlanFilterRequest{
					InstallmentPlanID: uuid.NewString(),
				},
			}

			// Set status for specific test cases
			if tc.name == "SUCCESS: Get list with Status Active" {
				filter.Status = constant.PaymentMethodGeneralStatusActive
			} else if tc.name == "SUCCESS: Get list with Status Inactive" {
				filter.Status = "INACTIVE"
			}

			_, err := repo.GetListPaymentMethodMerchant(ctx, filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestFindPaymentMethodByIdAndMerchant(t *testing.T) {
	paymentMethodID := uuid.NewString()
	merchantID := uuid.NewString()
	paymentMethod := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			UUID:      paymentMethodID,
			Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
			Name:      "VA Permata",
			Acquirer:  constant.BANK_ACQUIRER_PERMATA,
			CreatedAt: util.TimeNow,
			UpdatedAt: util.TimeNow,
		},
		IsActive:   true,
		MerchantID: merchantID,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find payment method by ID and merchant",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentMethodPtr := args.Get(1).(*paymentModel.PaymentMethodWithPivot)
					*paymentMethodPtr = *paymentMethod
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Find payment method by ID and merchant not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
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
					mock.AnythingOfType("*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Find payment method with valid MerchantConfig and RequiredDocuments",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentMethodWithPivot"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentMethodPtr := args.Get(1).(*paymentModel.PaymentMethodWithPivot)
					*paymentMethodPtr = paymentModel.PaymentMethodWithPivot{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID:     paymentMethodID,
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							Category: paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
							Name:     "VA Permata",
							Acquirer: constant.BANK_ACQUIRER_PERMATA,
							RequiredDocuments: types.NullJSONText{
								JSONText: []byte(`[{"name":"document1","required":true}]`),
								Valid:    true,
							},
							CreatedAt: util.TimeNow,
							UpdatedAt: util.TimeNow,
						},
						MerchantConfig: types.NullJSONText{
							JSONText: []byte(`{"channelConfig":{"key":"value"}}`),
							Valid:    true,
						},
						IsActive:   true,
						MerchantID: merchantID,
					}
				})
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tablePaymentMethodMerchant)
			_, err := repo.FindPaymentMethodByIdAndMerchant(ctx, paymentMethodID, merchantID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
