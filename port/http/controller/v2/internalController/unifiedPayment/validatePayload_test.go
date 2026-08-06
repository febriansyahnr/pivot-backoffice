package v2InternalUnifiedPaymentController

import (
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
)

func TestValidatePayload(t *testing.T) {
	minAmount := 1000.0
	maxAmount := 50000000.0
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
				MinAmount: &minAmount,
				MaxAmount: &maxAmount,
			},
			EwalletConfig: &config.UnifiedPaymentEwalletConfig{
				MinAmount: &minAmount,
				MaxAmount: &maxAmount,
			},
			CardConfig: &config.UnifiedPaymentCardConfig{
				MinAmount: &minAmount,
				MaxAmount: &maxAmount,
			},
			QrConfig: &config.UnifiedPaymentQrConfig{
				MinAmount: &minAmount,
				MaxAmount: &maxAmount,
			},
		},
		Environment: "development",
	}
	controller := &paymentController{
		config: cfg,
	}

	tests := []struct {
		name    string
		payload *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest
		wantErr bool
		err     error
		errMsg  string
	}{
		{
			name: "Valid payload - should pass",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Expiry time in the past",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(-time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("expiry time is not permitted to be less than current time")),
		},
		{
			name: "ERROR: IDR currency with decimal",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.50,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("amount value is not permitted to use decimal format")),
		},
		{
			name: "ERROR: AutoConfirm without payment method",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				AutoConfirm: true,
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrConfirmShouldChoosePaymentMethod),
		},
		{
			name: "ERROR: Multiple payment type with non-API mode",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeMultiple,
				Mode:        constant.UnifiedPaymentModeRedirect,
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentTypeMultipleNotAllowedForNonAPI),
		},
		{
			name: "ERROR: Multiple payment type with AutoConfirm false",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeMultiple,
				Mode:        constant.UnifiedPaymentModeAPI,
				AutoConfirm: false,
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentTypeMultipleNotAllowedAutoConfirmFalse),
		},
		{
			name: "ERROR: SaveForFutureUse with non-card payment method",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:         time.Now().Add(time.Hour),
				PaymentType:      constant.UnifiedPaymentTypeSingle,
				SaveForFutureUse: func() *bool { val := true; return &val }(),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodIsNotAllowedToSavedFutureUse),
		},
		{
			name: "ERROR: ShowSavedPayment with non-redirect mode",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:         time.Now().Add(time.Hour),
				PaymentType:      constant.UnifiedPaymentTypeSingle,
				Mode:             constant.UnifiedPaymentModeAPI,
				ShowSavedPayment: func() *bool { val := true; return &val }(),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodIsNotAllowedToShowSavedPayment),
		},
		{
			name: "ERROR: Multiple payment type with unsupported payment method",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeMultiple,
				Mode:        constant.UnifiedPaymentModeAPI,
				AutoConfirm: true,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentTypeMultipleNotAllowedForThisMethod),
		},
		{
			name: "ERROR: Virtual Account without payment method options",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method options for virtual account can not be empty")),
		},
		{
			name: "ERROR: Virtual Account with zero amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    0.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "PERMATA",
					},
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired),
		},
		{
			name: "ERROR: Virtual Account below minimum amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    500.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "PERMATA",
					},
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentBelowMinAmount),
		},
		{
			name: "ERROR: Virtual Account above maximum amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    60000000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "PERMATA",
					},
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAboveMaxAmount),
		},
		{
			name: "ERROR: E-wallet without payment method options",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:        "merchant-123",
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				ClientReferenceID: "client-ref-123",
				Amount: unifiedPaymentModel.Amount{
					Value:    100000,
					Currency: constant.CurrencyIDR,
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				ExpiryAt: time.Now().UTC().Add(1 * time.Hour),
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method options for ewallet can not be empty")),
		},
		{
			name: "ERROR: EWallet with zero amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    0.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: "DANA",
					},
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired),
		},
		{
			name: "ERROR: E-wallet ShopeePay exceeds max expiry time",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:        "merchant-123",
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				ClientReferenceID: "client-ref-123",
				Amount: unifiedPaymentModel.Amount{
					Value:    100000,
					Currency: constant.CurrencyIDR,
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: constant.UnifiedPaymentEWalletShopeePayAcquirer,
					},
				},
				ExpiryAt: time.Now().UTC().Add(6 * 24 * time.Hour),
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrEWalletShopeePayExceedMaxExpiryTime),
		},
		{
			name: "ERROR: E-wallet Dana exceeds max expiry time",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID:        "merchant-123",
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				ClientReferenceID: "client-ref-123",
				Amount: unifiedPaymentModel.Amount{
					Value:    100000,
					Currency: constant.CurrencyIDR,
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: constant.UnifiedPaymentEWalletDanaAcquirer,
					},
				},
				ExpiryAt: time.Now().UTC().Add(1 * time.Hour),
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrEWalletDanaExceedMaxExpiryTime),
		},
		{
			name: "ERROR: Card without card payment method detail in API mode with AutoConfirm",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				Mode:        constant.UnifiedPaymentModeAPI,
				AutoConfirm: true,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("payment method card can not be empty")),
		},
		{
			name: "ERROR: Card with zero amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    0.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired),
		},
		{
			name: "ERROR: QRIS with zero amount (non-static)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    0.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodQris,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountRequired),
		},
		{
			name: "ERROR: Static QRIS with non-zero amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0, // Non-zero amount for static payment should fail
				},
				ExpiryAt:    time.Time{},                         // Zero time for static payment
				PaymentType: constant.UnifiedPaymentTypeMultiple, // This makes it a static payment
				Mode:        constant.UnifiedPaymentModeAPI,
				AutoConfirm: true,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodQris,
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentStaticQrAmountMustBeZero),
		},
		{
			name: "ERROR: Split routing currency mismatch",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						MerchantId:  "merchant1",
						Type:        constant.SplitRoutingPaymentTypeFixed,
						Currency:    "USD",
						FixedAmount: 50000.0,
						Remarks:     "test",
					},
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("currency is not match")),
		},
		{
			name: "ERROR: Split routing amount exceeds payment amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						MerchantId:  "merchant1",
						Type:        constant.SplitRoutingPaymentTypeFixed,
						Currency:    "IDR",
						FixedAmount: 150000.0,
						Remarks:     "test",
					},
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("total split and routing amount must be not greater than payment amount")),
		},
		{
			name: "ERROR: Split routing percentage amount exceeds payment amount",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						MerchantId:       "merchant1",
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         "IDR",
						PercentageAmount: 150, // 150%
						Remarks:          "test",
					},
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("total split and routing amount must be not greater than payment amount")),
		},
		{
			name: "ERROR: Client Reference ID with special characters (feature flag disabled)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test@123#",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				MerchantID:  "test-merchant",
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrClientReferenceIDMustBeInAlphanumericFormat),
		},
		{
			name: "Valid - Client Reference ID with alphanumeric and dash",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test-123-abc",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				MerchantID:  "test-merchant",
			},
			wantErr: false,
		},
		{
			name: "Valid - Split routing with valid configuration",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						MerchantId:  "merchant1",
						Type:        constant.SplitRoutingPaymentTypeFixed,
						Currency:    "IDR",
						FixedAmount: 50000.0,
						Remarks:     "test",
					},
					{
						MerchantId:       "merchant2",
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         "IDR",
						PercentageAmount: 25, // 25% of remaining amount
						Remarks:          "test2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid - Save for future use with Card payment method",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:         time.Now().Add(time.Hour),
				PaymentType:      constant.UnifiedPaymentTypeSingle,
				SaveForFutureUse: func() *bool { val := true; return &val }(),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid - Show saved payment with Card payment method in redirect mode",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:         time.Now().Add(time.Hour),
				PaymentType:      constant.UnifiedPaymentTypeSingle,
				Mode:             constant.UnifiedPaymentModeRedirect,
				ShowSavedPayment: func() *bool { val := true; return &val }(),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid - Multiple payment type with QRIS (static payment)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    0.0, // Static payment must have zero amount
				},
				ExpiryAt:    time.Time{}, // Static payment typically has no expiry
				PaymentType: constant.UnifiedPaymentTypeMultiple,
				Mode:        constant.UnifiedPaymentModeAPI,
				AutoConfirm: true,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodQris,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid - Multiple payment type with Virtual Account (static payment)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    0.0, // Static payment must have zero amount
				},
				ExpiryAt:    time.Time{}, // Static payment typically has no expiry
				PaymentType: constant.UnifiedPaymentTypeMultiple,
				Mode:        constant.UnifiedPaymentModeAPI,
				AutoConfirm: true,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "PERMATA",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid - EWallet within expiry limits",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(25 * time.Minute), // Within Dana limit
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: "DANA",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: ExpirationMode with Virtual Account payment method",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "PERMATA",
					},
				},
				ExpirationMode: constant.UnifiedPaymentExpirationModeLoose,
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodIsNotAllowedToSetExpirationMode),
		},
		{
			name: "ERROR: ExpirationMode with QRIS payment method",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:       time.Now().Add(time.Hour),
				PaymentType:    constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodQris,
				},
				ExpirationMode: constant.UnifiedPaymentExpirationModeStrict,
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodIsNotAllowedToSetExpirationMode),
		},
		{
			name: "Valid - ExpirationMode with Card payment method (LOOSE)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ExpirationMode: constant.UnifiedPaymentExpirationModeLoose,
			},
			wantErr: false,
		},
		{
			name: "Valid - ExpirationMode with Card payment method (STRICT)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ExpirationMode: constant.UnifiedPaymentExpirationModeStrict,
			},
			wantErr: false,
		},
		{
			name: "Valid - ExpirationMode with EWallet payment method (LOOSE)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(25 * time.Minute),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: "DANA",
					},
				},
				ExpirationMode: constant.UnifiedPaymentExpirationModeLoose,
			},
			wantErr: false,
		},
		{
			name: "Valid - ExpirationMode with EWallet payment method (STRICT)",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:    time.Now().Add(25 * time.Minute),
				PaymentType: constant.UnifiedPaymentTypeSingle,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: "DANA",
					},
				},
				ExpirationMode: constant.UnifiedPaymentExpirationModeStrict,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Redirect mode with bypass status page but empty redirect URLs",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				Mode:              constant.UnifiedPaymentModeRedirect,
				BypassStatusPage:  true,
				RedirectUrl: unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "",
					FailureReturnUrl:    "",
					ExpirationReturnUrl: "",
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentRedirectUrlRequiredWhenBypassStatusPage),
		},
		{
			name: "ERROR: Redirect mode with bypass status page but missing success URL",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				Mode:              constant.UnifiedPaymentModeRedirect,
				BypassStatusPage:  true,
				RedirectUrl: unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "",
					FailureReturnUrl:    "https://example.com/failure",
					ExpirationReturnUrl: "https://example.com/expiration",
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentRedirectUrlRequiredWhenBypassStatusPage),
		},
		{
			name: "ERROR: Redirect mode with bypass status page but missing failure URL",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				Mode:              constant.UnifiedPaymentModeRedirect,
				BypassStatusPage:  true,
				RedirectUrl: unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "https://example.com/success",
					FailureReturnUrl:    "",
					ExpirationReturnUrl: "https://example.com/expiration",
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentRedirectUrlRequiredWhenBypassStatusPage),
		},
		{
			name: "ERROR: Redirect mode with bypass status page but missing expiration URL",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				Mode:              constant.UnifiedPaymentModeRedirect,
				BypassStatusPage:  true,
				RedirectUrl: unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "https://example.com/success",
					FailureReturnUrl:    "https://example.com/failure",
					ExpirationReturnUrl: "",
				},
			},
			wantErr: true,
			err:     pkgErrors.New(response.HttpErrRequest, constant.ErrUnifiedPaymentRedirectUrlRequiredWhenBypassStatusPage),
		},
		{
			name: "Valid - Redirect mode with bypass status page and all redirect URLs provided",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:         time.Now().Add(time.Hour),
				PaymentType:      constant.UnifiedPaymentTypeSingle,
				Mode:             constant.UnifiedPaymentModeRedirect,
				BypassStatusPage: true,
				RedirectUrl: unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "https://example.com/success",
					FailureReturnUrl:    "https://example.com/failure",
					ExpirationReturnUrl: "https://example.com/expiration",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid - Redirect mode with bypass status page false and empty redirect URLs",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				Mode:              constant.UnifiedPaymentModeRedirect,
				BypassStatusPage:  false,
				RedirectUrl: unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "",
					FailureReturnUrl:    "",
					ExpirationReturnUrl: "",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid - API mode with bypass status page true and empty redirect URLs",
			payload: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				ClientReferenceID: "test123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.0,
				},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentType:       constant.UnifiedPaymentTypeSingle,
				Mode:              constant.UnifiedPaymentModeAPI,
				BypassStatusPage:  true,
				RedirectUrl: unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "",
					FailureReturnUrl:    "",
					ExpirationReturnUrl: "",
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := controller.validatePayload(test.payload)

			if test.wantErr {
				assert.Error(t, err)
				if test.err != nil {
					assert.Equal(t, test.err, err)
				}
				if test.errMsg != "" {
					assert.Contains(t, err.Error(), test.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
