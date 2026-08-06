package paymentService

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetReceiptByID(t *testing.T) {
	conf := &config.Config{
		MerchantPortalConfig: config.MerchantPortalConfig{
			LogoURL: "https://example.com/logo.png",
		},
	}
	paymentID := uuid.NewString()
	merchantID := uuid.NewString()
	receiptKey := fmt.Sprintf(RedisTemplatePaymentReceiptKey, merchantID, paymentID)

	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	db, r := redismock.NewClientMock()

	tests := []struct {
		name         string
		request      *paymentModel.GetPaymentReceiptRequest
		modifierMock func()
		wantErr      string
	}{
		{
			name: "SUCCESS: Got receipt from redis",
			modifierMock: func() {
				r.ExpectGet(receiptKey).SetVal("pdf-bytes-content")
			},
		},
		{
			name: "ERROR: GetPaymentReceiptData returns error",
			modifierMock: func() {
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On(
					"GetPaymentReceiptData",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: GetPaymentReceiptData by reference ID returns error",
			request: &paymentModel.GetPaymentReceiptRequest{
				ReferenceID: "reference-id",
				MerchantID:  merchantID,
			},
			modifierMock: func() {
				receiptKey := fmt.Sprintf(RedisTemplatePaymentReceiptKey, merchantID, "reference-id")
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On(
					"GetPaymentReceiptData",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Payment not found",
			modifierMock: func() {
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On(
					"GetPaymentReceiptData",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Once().Return(nil, nil)
			},
			wantErr: constant.ErrPaymentNotFound.Error(),
		},
		{
			name: "ERROR: Payment status not in success status",
			modifierMock: func() {
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On(
					"GetPaymentReceiptData",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Once().Return(&paymentModel.PaymentReceiptDTO{
					UUID:       paymentID,
					Status:     constant.UnifiedPaymentSessionStatusProcessing,
					MerchantID: merchantID,
				}, nil)
			},
			wantErr: constant.ErrPaymentNotSuccessYet.Error(),
		},
		{
			name: "ERROR: Merchant ID is not valid",
			modifierMock: func() {
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				paymentRepo.On(
					"GetPaymentReceiptData",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Once().Return(&paymentModel.PaymentReceiptDTO{
					UUID:       uuid.NewString(),
					Status:     constant.UnifiedPaymentSessionStatusPaid,
					MerchantID: uuid.NewString(), // other merchant
				}, nil)
			},
			wantErr: constant.ErrPaymentNotFound.Error(),
		},
		{
			name: "ERROR: Generate PDF failed",
			modifierMock: func() {
				r.ExpectGet(receiptKey).SetErr(constant.ErrSomeErrorForUnitTest)

				referenceID := "ref-001"
				paymentRepo.On(
					"GetPaymentReceiptData",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&paymentModel.PaymentReceiptDTO{
					UUID:          paymentID,
					MerchantID:    merchantID,
					Status:        constant.UnifiedPaymentSessionStatusPaid,
					TotalAmount:   decimal.NewFromInt(1000000),
					ReferenceID:   &referenceID,
					CreatedAt:     time.Now(),
					PaymentMethod: sql.NullString{String: "VIRTUAL_ACCOUNT", Valid: true},
					MerchantName:  sql.NullString{String: "Merchant Name", Valid: true},
				}, nil).Once()

				// Note: PDF generation will fail because template file doesn't exist in test environment
			},
			wantErr: constant.ErrFailedToGenerateReceipt.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			tc.modifierMock()

			svc := New(
				paymentRepo, logger, nil, nil, nil, nil, nil,
				WithConfig(conf),
				WithRedisClient(redisExt.WrapRedisClient(db, nil)),
			)

			req := &paymentModel.GetPaymentReceiptRequest{
				PaymentID:  paymentID,
				MerchantID: merchantID,
			}
			if tc.request != nil {
				req = tc.request
			}
			_, err := svc.GetReceiptByID(ctx, req)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
