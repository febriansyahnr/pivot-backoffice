package paymentRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentRepository_CountActiveStaticPayment(t *testing.T) {
	merchantID := uuid.NewString()
	paymentMethodID := uuid.NewString()
	derivedMerchantID := uuid.NewString()

	testCases := []struct {
		name            string
		mockSetup       func(mysqlMock *mysqlMocks.IMySqlExt)
		setupContext    func() context.Context
		merchantID      string
		paymentMethodID string
		expectedCount   int
		wantErr         bool
	}{
		{
			name: "SUCCESS: Count active static payment",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*int"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					countPtr := args.Get(1).(*int)
					*countPtr = 3
				})
			},
			setupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
				return context.WithValue(ctx, constant.CtxDerivedMerchantID, derivedMerchantID)
			},
			merchantID:      merchantID,
			paymentMethodID: paymentMethodID,
			expectedCount:   3,
			wantErr:         false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*int"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("database error"))
			},
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			},
			merchantID:      merchantID,
			paymentMethodID: paymentMethodID,
			expectedCount:   0,
			wantErr:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := tc.setupContext()
			count, err := repo.CountActiveStaticPayment(ctx, tc.merchantID, tc.paymentMethodID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, 0, count)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, count)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestPaymentRepository_GetFirstActiveStaticQrisByMerchant(t *testing.T) {
	merchantID := uuid.NewString()
	partnerReferenceNo := "test-ref-123"

	testCases := []struct {
		name      string
		mockSetup func(*mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get first active static QRIS payment",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					dto := args.Get(1).(*paymentModel.PaymentWithPaymentMethodDTO)
					dto.UUID = uuid.NewString()
					dto.MerchantID = merchantID
					dto.Status = constant.StatusActive
					dto.Type = constant.UnifiedPaymentTypeMultiple
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("database connection failed"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No records found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")

			payment, err := repo.GetFirstActiveStaticQrisByMerchant(ctx, merchantID, partnerReferenceNo)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
