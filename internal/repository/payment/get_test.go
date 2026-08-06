package paymentRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentRepository_GetPaymentById(t *testing.T) {
	paymentId := uuid.NewString()

	payment := &paymentModel.PaymentWithPaymentMethodDTO{
		UUID:                     paymentId,
		MerchantID:               uuid.NewString(),
		CustomerID:               uuid.NewString(),
		PaymentMethodID:          uuid.NewString(),
		ProcessorReferenceNumber: nil,
		Currency:                 "IDR",
		Amount:                   decimal.NewFromInt(1000000),
		TotalAmount:              decimal.NewFromInt(1000000),
		Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
		CreatedAt:                util.TimeNow,
		UpdatedAt:                util.TimeNow,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment by id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentWithPaymentMethodDTOMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentPtr := args.Get(1).(*paymentModel.PaymentWithPaymentMethodDTO)
					*paymentPtr = *payment
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentWithPaymentMethodDTOMockType(),
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
					constant.ValueCtxMockType(),
					constant.PtrPaymentWithPaymentMethodDTOMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			_, err := repo.GetPaymentById(ctx, paymentId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepository_GetPaymentItemsByPaymentId(t *testing.T) {
	paymentId := uuid.NewString()

	paymentItemDTO := &paymentModel.PaymentItemDTO{
		UUID:        uuid.NewString(),
		PaymentID:   uuid.NewString(),
		Name:        "Bill A",
		Qty:         1,
		Currency:    "IDR",
		Amount:      decimal.NewFromInt(1000000),
		TotalAmount: decimal.NewFromInt(1000000),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	paymentItemDTOs := []*paymentModel.PaymentItemDTO{paymentItemDTO}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment items by payment id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentItemDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentPtr := args.Get(1).(*[]*paymentModel.PaymentItemDTO)
					*paymentPtr = paymentItemDTOs
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*paymentModel.PaymentItemDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
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
					mock.AnythingOfType("*[]*paymentModel.PaymentItemDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			_, err := repo.GetPaymentItemsByPaymentId(ctx, paymentId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepository_GetActivePaymentByProcessorReferenceNumber(t *testing.T) {
	processorReferenceNumber := "1234123412341234"
	payment := &paymentModel.PaymentWithPaymentMethodDTO{
		UUID:                     uuid.NewString(),
		MerchantID:               uuid.NewString(),
		CustomerID:               uuid.NewString(),
		PaymentMethodID:          uuid.NewString(),
		ProcessorReferenceNumber: &processorReferenceNumber,
		Currency:                 "IDR",
		Amount:                   decimal.NewFromInt(1000000),
		TotalAmount:              decimal.NewFromInt(1000000),
		Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
		CreatedAt:                util.TimeNow,
		UpdatedAt:                util.TimeNow,
	}

	expiredCardPayment := &paymentModel.PaymentWithPaymentMethodDTO{
		UUID:                     uuid.NewString(),
		MerchantID:               uuid.NewString(),
		CustomerID:               uuid.NewString(),
		PaymentMethodID:          uuid.NewString(),
		ProcessorReferenceNumber: &processorReferenceNumber,
		Currency:                 "IDR",
		Amount:                   decimal.NewFromInt(1000000),
		TotalAmount:              decimal.NewFromInt(1000000),
		Status:                   paymentConstant.UnifiedPaymentStatusExpired,
		CreatedAt:                util.TimeNow,
		UpdatedAt:                util.TimeNow,
		PaymentMethodType: sql.NullString{
			Valid:  true,
			String: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
		},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get active payment by processor reference number",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentWithPaymentMethodDTOMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentPtr := args.Get(1).(*paymentModel.PaymentWithPaymentMethodDTO)
					*paymentPtr = *payment
				})
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get expired payment by processor reference number",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentWithPaymentMethodDTOMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentPtr := args.Get(1).(*paymentModel.PaymentWithPaymentMethodDTO)
					*paymentPtr = *expiredCardPayment
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrPaymentWithPaymentMethodDTOMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
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
					constant.ValueCtxMockType(),
					constant.PtrPaymentWithPaymentMethodDTOMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			_, err := repo.GetActivePaymentByProcessorReferenceNumber(ctx, &paymentModel.GetActivePaymentByProcessorReferenceNumberRequest{
				ProcessorReferenceNumber: processorReferenceNumber,
				Amount:                   decimal.NewFromInt(1000000),
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

func TestPaymentRepository_GetPaymentByMerchantAndReferenceId(t *testing.T) {
	paymentId := uuid.NewString()
	merchantId := uuid.NewString()
	referenceId := "some-reference-id"

	payment := &paymentModel.PaymentWithPaymentMethodDTO{
		UUID:                     paymentId,
		MerchantID:               uuid.NewString(),
		CustomerID:               uuid.NewString(),
		PaymentMethodID:          uuid.NewString(),
		ProcessorReferenceNumber: nil,
		Currency:                 "IDR",
		Amount:                   decimal.NewFromInt(1000000),
		TotalAmount:              decimal.NewFromInt(1000000),
		Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
		CreatedAt:                util.TimeNow,
		UpdatedAt:                util.TimeNow,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment by id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentPtr := args.Get(1).(*paymentModel.PaymentWithPaymentMethodDTO)
					*paymentPtr = *payment
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
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
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			_, err := repo.GetPaymentByMerchantAndReferenceId(ctx, merchantId, referenceId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepository_GetPaymentByIdAndMerchantId(t *testing.T) {
	paymentId := uuid.NewString()
	merchantId := uuid.NewString()

	payment := &paymentModel.PaymentWithPaymentMethodDTO{
		UUID:                     paymentId,
		MerchantID:               uuid.NewString(),
		CustomerID:               uuid.NewString(),
		PaymentMethodID:          uuid.NewString(),
		ProcessorReferenceNumber: nil,
		Currency:                 "IDR",
		Amount:                   decimal.NewFromInt(1000000),
		TotalAmount:              decimal.NewFromInt(1000000),
		Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
		CreatedAt:                util.TimeNow,
		UpdatedAt:                util.TimeNow,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment by id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentPtr := args.Get(1).(*paymentModel.PaymentWithPaymentMethodDTO)
					*paymentPtr = *payment
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
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
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO"),
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			_, err := repo.GetPaymentByIdAndMerchantId(ctx, paymentId, merchantId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetPaymentQrStaticByMerchantId(t *testing.T) {
	merchantId := uuid.NewString()
	paymentMethodId := uuid.NewString()
	subMerchantId := uuid.NewString()

	payment := &paymentModel.PaymentDTO{
		UUID:                     uuid.NewString(),
		MerchantID:               merchantId,
		PaymentMethodID:          paymentMethodId,
		ProcessorReferenceNumber: nil,
		Currency:                 "IDR",
		Amount:                   decimal.NewFromInt(1000000),
		TotalAmount:              decimal.NewFromInt(1000000),
		Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
		CreatedAt:                util.TimeNow,
		UpdatedAt:                util.TimeNow,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get payment qr static by merchant id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					paymentPtr := args.Get(1).(*paymentModel.PaymentDTO)
					*paymentPtr = *payment
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentDTO"),
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
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentDTO"),
					mock.AnythingOfType("string"),
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			_, err := repo.GetPaymentQrStaticByMerchantId(ctx, merchantId, subMerchantId, paymentMethodId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepository_GetPaymentReceiptData(t *testing.T) {
	paymentID := uuid.NewString()
	referenceID := "ref-123"
	merchantID := uuid.NewString()

	receiptDTO := &paymentModel.PaymentReceiptDTO{
		UUID:          paymentID,
		ReferenceID:   &referenceID,
		MerchantID:    merchantID,
		TotalAmount:   decimal.NewFromInt(1000000),
		Status:        constant.UnifiedPaymentSessionStatusPaid,
		CreatedAt:     util.TimeNow,
		MerchantName:  sql.NullString{Valid: true, String: "Test Merchant"},
		PaymentMethod: sql.NullString{Valid: true, String: "VIRTUAL_ACCOUNT"},
	}

	testCases := []struct {
		name        string
		paymentID   string
		referenceID string
		merchantID  string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr     bool
		wantNil     bool
	}{
		{
			name:        "SUCCESS: Get by paymentID only",
			paymentID:   paymentID,
			referenceID: "",
			merchantID:  "",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentReceiptDTO"),
					mock.AnythingOfType("string"),
					paymentID,
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*paymentModel.PaymentReceiptDTO)
					*ptr = *receiptDTO
				})
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:        "SUCCESS: Get by referenceID only",
			paymentID:   "",
			referenceID: referenceID,
			merchantID:  "",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentReceiptDTO"),
					mock.AnythingOfType("string"),
					referenceID,
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*paymentModel.PaymentReceiptDTO)
					*ptr = *receiptDTO
				})
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:        "SUCCESS: Get by paymentID and merchantID",
			paymentID:   paymentID,
			referenceID: "",
			merchantID:  merchantID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentReceiptDTO"),
					mock.AnythingOfType("string"),
					paymentID,
					merchantID,
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*paymentModel.PaymentReceiptDTO)
					*ptr = *receiptDTO
				})
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:        "SUCCESS: Get by referenceID and merchantID",
			paymentID:   "",
			referenceID: referenceID,
			merchantID:  merchantID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentReceiptDTO"),
					mock.AnythingOfType("string"),
					referenceID,
					merchantID,
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*paymentModel.PaymentReceiptDTO)
					*ptr = *receiptDTO
				})
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:        "SUCCESS: Get by all parameters",
			paymentID:   paymentID,
			referenceID: referenceID,
			merchantID:  merchantID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentReceiptDTO"),
					mock.AnythingOfType("string"),
					paymentID,
					referenceID,
					merchantID,
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*paymentModel.PaymentReceiptDTO)
					*ptr = *receiptDTO
				})
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:        "NOT_FOUND: No rows returned",
			paymentID:   paymentID,
			referenceID: "",
			merchantID:  "",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentReceiptDTO"),
					mock.AnythingOfType("string"),
					paymentID,
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name:        "ERROR: Database error",
			paymentID:   paymentID,
			referenceID: "",
			merchantID:  "",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.PaymentReceiptDTO"),
					mock.AnythingOfType("string"),
					paymentID,
				).Return(errors.New("database error"))
			},
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			result, err := repo.GetPaymentReceiptData(ctx, tc.paymentID, tc.referenceID, tc.merchantID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
