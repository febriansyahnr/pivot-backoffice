package creditcard

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentNotification(t *testing.T) {
	now := time.Now()
	validRequest := creditcardModel.CardPaymentNotificationRequest{
		Event: "CREDIT_CARD_PAYMENTS_CALLBACK",
		Data: creditcardModel.PaymentNotificationDataRequest{
			ReferenceID:           "reference-id", // NOSONAR
			Amount:                decimal.NewFromFloat(10000),
			Currency:              "IDR",                     // NOSONAR
			AcquirerTransactionID: "acquirer-transaction-id", // NOSONAR
			PaymentStatus:         "SUCCESS",                 // NOSONAR
			CardData:              &creditcardModel.CardDataRequest{First8Digit: "12345678", Last4Digit: "4321", CardType: "credit", CardBrand: "visa", CardIssuing: "bank", CountryCode: "ID", Fingerprint: "fingerprint"},
			Device: &creditcardModel.PaymentNotificationDevice{
				IPAddress: "0.0.0.0", // NOSONAR
			},
			ResponseCode: &creditcardModel.PaymentNotificationResponseCode{
				AcquirerCode:    "00",       // NOSONAR
				AcquirerMessage: "Approved", // NOSONAR
			},
			Error:   &creditcardModel.PaymentNotificationError{},
			Updated: now,
		},
	}
	input, err := json.Marshal(validRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name       string
		input      []byte
		channel    string
		mocksSetup func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService)
		wantErr    bool
	}{
		{
			name:    "SUCCESS",
			input:   input,
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: nil}, nil)

				creditcardSvc.On(
					"PaymentNotification",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("card.CardPaymentNotificationRequest"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Invalid payload",
			input:   []byte{},
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, _ *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Payment notification service",
			input:   input,
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: nil}, nil)

				creditcardSvc.On(
					"PaymentNotification",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("card.CardPaymentNotificationRequest"),
				).Return(pkgError.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest))
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Skip notification for unified refund with VOID status",
			input: func() []byte {
				req := validRequest
				req.Data.PaymentStatus = constant.CreditCardStatusVoid
				data, _ := json.Marshal(req)
				return data
			}(),
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				metadata := map[string]interface{}{
					"refundId":           "refund-123",
					"isUnifiedPaymentV2": true,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: &metadata}, nil)

				refundSvc.On(
					"GetTotalRefundedAmount", mock.Anything, mock.Anything,
				).Return(10000.00, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Skip notification for unified refund with REFUNDED status",
			input: func() []byte {
				req := validRequest
				req.Data.PaymentStatus = constant.CreditCardStatusRefunded
				data, _ := json.Marshal(req)
				return data
			}(),
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				metadata := map[string]interface{}{
					"refundId":           "refund-456",
					"isUnifiedPaymentV2": true,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: &metadata}, nil)

				refundSvc.On(
					"GetTotalRefundedAmount", mock.Anything, mock.Anything,
				).Return(10000.00, nil)

			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Skip notification for unified payment with VOIDED status for capture method MANUAL",
			input: func() []byte {
				req := validRequest
				req.Data.PaymentStatus = constant.CreditCardStatusVoid
				data, _ := json.Marshal(req)
				return data
			}(),
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				metadata := map[string]interface{}{
					"isUnifiedPaymentV2": true,
					"paymentMethodOptions": map[string]interface{}{
						"card": map[string]interface{}{
							"captureMethod": "MANUAL",
						},
					},
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: &metadata}, nil)

				refundSvc.On(
					"GetTotalRefundedAmount", mock.Anything, mock.Anything,
				).Return(0.00, nil)

			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Process notification when refundId is empty with VOID status",
			input: func() []byte {
				req := validRequest
				req.Data.PaymentStatus = constant.CreditCardStatusVoid
				data, _ := json.Marshal(req)
				return data
			}(),
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				metadata := map[string]interface{}{
					"refundId": "",
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: &metadata}, nil)

				refundSvc.On(
					"GetTotalRefundedAmount", mock.Anything, mock.Anything,
				).Return(0.00, nil)

				creditcardSvc.On(
					"PaymentNotification",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("card.CardPaymentNotificationRequest"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Skip notification for SUCCESS status and CAPTURE transaction type",
			input: func() []byte {
				req := validRequest
				req.Data.PaymentStatus = constant.CreditCardStatusSuccess
				req.Data.Type = constant.CardTransactionTypeCapture
				data, _ := json.Marshal(req)
				return data
			}(),
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				metadata := map[string]interface{}{
					"isUnifiedPaymentV2": true,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: &metadata}, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Process for SUCCESS status and CAPTURE transaction type",
			input: func() []byte {
				req := validRequest
				req.Data.PaymentStatus = constant.CreditCardStatusSuccess
				req.Data.Type = constant.CardTransactionTypeCapture
				data, _ := json.Marshal(req)
				return data
			}(),
			channel: constant.ChannelCreditCard,
			mocksSetup: func(creditcardSvc *serviceMocks.ICreditCardService, paymentSvc *serviceMocks.IPaymentService, refundSvc *serviceMocks.IRefundService) {
				metadata := map[string]interface{}{
					"isUnifiedPaymentV2": true,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{UUID: "test-uuid", Metadata: &metadata}, nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, err := loggerMocks.NewZapLogger(loggerMocks.Config{})
			assert.NoError(t, err)

			creditcardSvcMock := serviceMocks.NewICreditCardService(t)
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			paymentSvc := serviceMocks.NewIPaymentService(t)
			unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			refundSvc := serviceMocks.NewIRefundService(t)
			tc.mocksSetup(creditcardSvcMock, paymentSvc, refundSvc)

			consumer := New(mockLogger, creditcardSvcMock, orchestratorSvc, paymentSvc, unifiedPaymentSvc, refundSvc)
			ctx := context.Background()
			err = consumer.PaymentNotification(ctx, tc.input, tc.channel)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			creditcardSvcMock.AssertExpectations(t)
		})
	}
}
