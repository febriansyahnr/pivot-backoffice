package paymentMethodRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByPaymentMethodId(t *testing.T) {
	newUuid := uuid.NewString()

	paymentMethod := &paymentModel.PaymentMethod{
		UUID:      newUuid,
		Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
		Name:      "VA Permata",
		Acquirer:  constant.BANK_ACQUIRER_PERMATA,
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	paymentMethodId := newUuid

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment method by uuid",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentMethodPtr := args.Get(1).(*paymentModel.PaymentMethod)
					*paymentMethodPtr = *paymentMethod
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Method Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payment_methods")
			_, err := repo.GetPaymentMethodById(ctx, paymentMethodId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetActivePaymentMethodByRequest(t *testing.T) {
	paymentMethod := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			UUID:      uuid.NewString(),
			Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
			Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			Name:      "VA Permata",
			Acquirer:  constant.BANK_ACQUIRER_PERMATA,
			CreatedAt: util.TimeNow,
			UpdatedAt: util.TimeNow,
		},
	}

	methodType := paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT
	bankCode := constant.BANK_ACQUIRER_PERMATA

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment method by type and bankCode",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentMethodWithPivot(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentMethodPtr := args.Get(1).(*paymentModel.PaymentMethodWithPivot)
					*paymentMethodPtr = *paymentMethod
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Method Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentMethodWithPivot(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentMethodWithPivot(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get payment method with valid MerchantConfig and RequiredDocuments",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentMethodWithPivot(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentMethodPtr := args.Get(1).(*paymentModel.PaymentMethodWithPivot)
					*paymentMethodPtr = paymentModel.PaymentMethodWithPivot{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID:     uuid.NewString(),
							Category: paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payment_methods")
			_, err := repo.GetActivePaymentMethodByRequest(ctx, &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: uuid.NewString(),
				Category:   paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
				Type:       methodType,
				Acquirer:   bankCode,
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetAllPaymentMethodByCategory(t *testing.T) {
	paymentMethod := &paymentModel.PaymentMethod{
		UUID:      uuid.NewString(),
		Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
		Name:      "VA Permata",
		Acquirer:  constant.BANK_ACQUIRER_PERMATA,
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	methodCategory := paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment method by type and bankCode",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentMethodPtr := args.Get(1).(*[]*paymentModel.PaymentMethod)
					*paymentMethodPtr = []*paymentModel.PaymentMethod{paymentMethod}
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Method Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payment_methods")
			_, err := repo.GetAllPaymentMethodByCategory(ctx, methodCategory)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepository_GetPaymentMethodByType(t *testing.T) {
	paymentMethod := &paymentModel.PaymentMethod{
		UUID:      uuid.NewString(),
		Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
		Name:      "VA Permata",
		Acquirer:  constant.BANK_ACQUIRER_PERMATA,
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	methodCategory := paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment method by type and bankCode",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentMethodPtr := args.Get(1).(*[]*paymentModel.PaymentMethod)
					*paymentMethodPtr = []*paymentModel.PaymentMethod{paymentMethod}
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Method Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*paymentModel.PaymentMethod"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payment_methods")
			_, err := repo.GetPaymentMethodByType(ctx, methodCategory)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetPaymentMethodByCategoryTypeAndAcquirer(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	paymentMethod := paymentModel.PaymentMethod{
		UUID: "74735bb6-77a2-49f8-9e7d-45ac359e8fb8",
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *paymentModel.PaymentMethod
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, constant.ReferencePayment, constant.ChannelVirtualAccount, "BRI", // NOSONAR
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR: Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, constant.ReferencePayment, constant.ChannelVirtualAccount, "BRI", // NOSONAR
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, constant.ReferencePayment, constant.ChannelVirtualAccount, "BRI", // NOSONAR
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*paymentModel.PaymentMethod) = paymentMethod
				}).Return(nil)
			},
			wantResult: &paymentMethod,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetPaymentMethodByCategoryTypeAndAcquirer(context.Background(), constant.ReferencePayment, constant.ChannelVirtualAccount, "BRI")

			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
