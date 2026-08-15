package card_test

import (
	"testing"
	"time"

	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcard"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCreateCardPaymentRequestValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name   string
		req    card.CreateCardPaymentRequest
		hasErr bool
	}{
		{
			name: "valid",
			req: card.CreateCardPaymentRequest{
				PaymentUUID:          uuid.New(),
				ReferenceID:          "ref123",
				Amount:               decimal.NewFromInt(100),
				Currency:             "IDR",
				AuthenticationMethod: "3ds",
				RedirectUrl: card.CreditcardRedirectUrlRequest{
					SuccessUrl: "http://example.com/success",
					FailedUrl:  "http://example.com/fail",
				},
			},
			hasErr: false,
		},
		{
			name:   "missing required fields",
			req:    card.CreateCardPaymentRequest{},
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.req)
			if tt.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardPaymentNotificationRequestValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name   string
		req    card.CardPaymentNotificationRequest
		hasErr bool
	}{
		{
			name: "valid",
			req: card.CardPaymentNotificationRequest{
				Event: "payment.success",
				Data: card.PaymentNotificationDataRequest{
					PaymentUUID:           uuid.New(),
					AuthenticationMethod:  "3ds",
					AcquirerTransactionID: "tx123",
					MerchantID:            uuid.New(),
					ReferenceID:           "ref123",
					Amount:                decimal.NewFromInt(200),
					Currency:              "IDR",
					PaymentStatus:         "success",
					Updated:               time.Now(),
					RedirectUrl: card.CreditcardRedirectUrlRequest{
						SuccessUrl: "https://example.com/success",
						FailedUrl:  "https://example.com/fail",
					},
				},
			},
			hasErr: false,
		},
		{
			name:   "missing required fields",
			req:    card.CardPaymentNotificationRequest{},
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.req)
			if tt.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToCreditcardCoreGetTransactionListRequest(t *testing.T) {
	req := card.GetTransactionListRequest{
		Page:                1,
		PerPage:             10,
		DateFrom:            "2024-02-01",
		DateTo:              "2024-02-10",
		TrxType:             "SALE",
		ChargeStatus:        "SUCCESS",
		VoidStatus:          "NONE",
		ClientTransactionID: "client-tx-123",
		IssuingBank:         "Bank XYZ",
		CardFingerprint:     "fingerprint-abc",
		PaymentUUID:         uuid.New().String(),
		ChargeFrom:          "2024-02-01",
		ChargeTo:            "2024-02-10",
		RefundFrom:          "2024-02-05",
		RefundTo:            "2024-02-10",
	}

	expected := &creditcardCoreProcessorModel.GetTransactionListRequest{
		Limit:               req.PerPage,
		Page:                req.Page,
		DateFrom:            req.DateFrom,
		DateTo:              req.DateTo,
		TrxType:             req.TrxType,
		ChargeStatus:        req.ChargeStatus,
		VoidStatus:          req.VoidStatus,
		ClientTransactionID: req.ClientTransactionID,
		IssuingBank:         req.IssuingBank,
		CardFingerprint:     req.CardFingerprint,
		PaymentUUID:         req.PaymentUUID,
		ChargeFrom:          req.ChargeFrom,
		ChargeTo:            req.ChargeTo,
		RefundFrom:          req.RefundFrom,
		RefundTo:            req.RefundTo,
	}

	actual := req.ToCreditcardCoreGetTransactionListRequest()

	assert.Equal(t, expected, actual, "Converted request does not match expected result")
}
