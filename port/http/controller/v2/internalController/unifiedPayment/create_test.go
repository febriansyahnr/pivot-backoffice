package v2InternalUnifiedPaymentController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	loggerMock := loggerMocks.NewILogger(t)
	loggerMock.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	minCardAmount := 10000.00
	maxCardAmount := 50000000.00
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{},
			CardConfig: &config.UnifiedPaymentCardConfig{
				MinAmount: &minCardAmount,
				MaxAmount: &maxCardAmount,
			},
		},
	}
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc), WithLogger(loggerMock))

	validRequest := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		ClientReferenceID: "123456",
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    100000.00,
		},
		AutoConfirm: true,
		Mode:        constant.UnifiedPaymentModeRedirect,
		RedirectUrl: unifiedPaymentModel.RedirectUrl{
			SuccessReturnUrl:    "https://success.url",
			FailureReturnUrl:    "https://failure.url",
			ExpirationReturnUrl: "https://expiration.url",
		},
		ExpiryAt: time.Now().Add(time.Hour),
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: constant.UnifiedPaymentMethodVA,
		},
		PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
			VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
				Channel: "PERMATA",
			},
		},
		SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
			MerchantId:       uuid.NewString(),
			Type:             constant.SplitRoutingPaymentTypePercentage,
			Currency:         constant.CurrencyIDR,
			PercentageAmount: 100,
			Remarks:          "test",
		}},
	}

	invalidClientMetadata := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		ClientReferenceID: "123456",
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    100000.00,
		},
		AutoConfirm: true,
		Mode:        constant.UnifiedPaymentModeRedirect,
		RedirectUrl: unifiedPaymentModel.RedirectUrl{
			SuccessReturnUrl:    "https://success.url",
			FailureReturnUrl:    "https://failure.url",
			ExpirationReturnUrl: "https://expiration.url",
		},
		ExpiryAt: time.Now().Add(time.Hour),
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: constant.UnifiedPaymentMethodVA,
		},
		Metadata: map[string]interface{}{
			"okelur": "okelur",
			"other":  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
			VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
				Channel: "PERMATA",
			},
		},
		SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
			MerchantId:       uuid.NewString(),
			Type:             constant.SplitRoutingPaymentTypePercentage,
			Currency:         constant.CurrencyIDR,
			PercentageAmount: 100,
			Remarks:          "test",
		}},
	}

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		requestBody   func() []byte
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			wantStatus:   http.StatusUnauthorized,
			wantResponse: `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Invalid Payload",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				return []byte(`{"missing-payload"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request payload"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Field Format Invalid",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				return []byte(`{"test": "test"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Key: 'CreateUnifiedPaymentSessionRequest.ClientReferenceID' Error:Field validation for 'ClientReferenceID' failed on the 'required' tag\nKey: 'CreateUnifiedPaymentSessionRequest.RedirectUrl' Error:Field validation for 'RedirectUrl' failed on the 'required' tag"}],"traceId":""}}`,
		},
		{
			name: "ERROR: First authorization for recurring payment",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456", // NOSONAR
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456789", // NOSONAR
					Mode:              constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",    // NOSONAR
						FailureReturnUrl:    "https://failure.url",    // NOSONAR
						ExpirationReturnUrl: "https://expiration.url", // NOSONAR
					},
					ExpiryAt:                   time.Now().Add(time.Hour),
					InitiateFirstAuthorization: true,
					RecurringID:                "3dfeefca-851f-4b46-96c4-4bc9cd69692f", // NOSONAR
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"First authorization for recurring payments is only supported for CARD payment methods"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Subsequent recurring payments",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456", // NOSONAR
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456789", // NOSONAR
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					Mode: constant.UnifiedPaymentModeAPI,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",    // NOSONAR
						FailureReturnUrl:    "https://failure.url",    // NOSONAR
						ExpirationReturnUrl: "https://expiration.url", // NOSONAR
					},
					ExpiryAt:                   time.Now().Add(time.Hour),
					InitiateFirstAuthorization: false,
					RecurringID:                "3dfeefca-851f-4b46-96c4-4bc9cd69692f", // NOSONAR
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Subsequent recurring payments are not allowed to provide a payment method"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Recurring payment mode",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456789", // NOSONAR
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					Mode: constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",    // NOSONAR
						FailureReturnUrl:    "https://failure.url",    // NOSONAR
						ExpirationReturnUrl: "https://expiration.url", // NOSONAR
					},
					ExpiryAt:                   time.Now().Add(time.Hour),
					InitiateFirstAuthorization: true,
					RecurringID:                "3dfeefca-851f-4b46-96c4-4bc9cd69692f", // NOSONAR
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Recurring payments can only be created using API mode"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Recurring payments do not allow Customer ID and Customer Object",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456789", // NOSONAR
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					Mode: constant.UnifiedPaymentModeAPI,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",    // NOSONAR
						FailureReturnUrl:    "https://failure.url",    // NOSONAR
						ExpirationReturnUrl: "https://expiration.url", // NOSONAR
					},
					ExpiryAt:                   time.Now().Add(time.Hour),
					InitiateFirstAuthorization: true,
					RecurringID:                "3dfeefca-851f-4b46-96c4-4bc9cd69692f", // NOSONAR
					CustomerID:                 "9d38865c-3c62-4942-913d-7f6767024a81",
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Customer ID or Customer Object are not required for recurring payments"}],"traceId":""}}`,
		},
		{
			name: "ERROR: For subsequent recurring payments, the autoConfirm value must be TRUE",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456789", // NOSONAR
					Mode:              constant.UnifiedPaymentModeAPI,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",    // NOSONAR
						FailureReturnUrl:    "https://failure.url",    // NOSONAR
						ExpirationReturnUrl: "https://expiration.url", // NOSONAR
					},
					ExpiryAt:    time.Now().Add(time.Hour),
					RecurringID: "3dfeefca-851f-4b46-96c4-4bc9cd69692f", // NOSONAR
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"For subsequent recurring payments, the autoConfirm value must be TRUE"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Invalid expiry time",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(-time.Hour),
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusUnprocessableEntity,
			wantResponse: `{"code":"unprocessable_entity", "error":{"details":[{"field":"", "message":"expiry time is not permitted to be less than current time"}], "traceId":"", "type":"API_ERROR"}, "message":"Unprocessable entity"}`,
		},
		{
			name: "ERROR: Currency IDR not permitted decimal format",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.03,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusUnprocessableEntity,
			wantResponse: `{"code":"unprocessable_entity", "error":{"details":[{"field":"", "message":"amount value is not permitted to use decimal format"}], "traceId":"", "type":"API_ERROR"}, "message":"Unprocessable entity"}`,
		},
		{
			name: "ERROR: Auto confirm should have payment method",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusUnprocessableEntity,
			wantResponse: `{"code":"unprocessable_entity", "error":{"details":[{"field":"", "message":"the confirm request should have a chosen payment method"}], "traceId":"", "type":"API_ERROR"}, "message":"Unprocessable entity"}`,
		},
		{
			name: "ERROR: Payment method options for VA should not be empty",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusUnprocessableEntity,
			wantResponse: `{"code":"unprocessable_entity", "error":{"details":[{"field":"", "message":"payment method options for virtual account can not be empty"}], "traceId":"", "type":"API_ERROR"}, "message":"Unprocessable entity"}`,
		},
		{
			name: "ERROR: Payment method options for Ewallet should not be empty",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodEWallet,
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusUnprocessableEntity,
			wantResponse: `{"code":"unprocessable_entity", "error":{"details":[{"field":"", "message":"payment method options for ewallet can not be empty"}], "traceId":"", "type":"API_ERROR"}, "message":"Unprocessable entity"}`,
		},
		{
			name: "ERROR: Service returns error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("CreateSession", constant.ValueCtxMockType(), constant.PtrCreateUnifiedPaymentSessionRequest()).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "ERROR: invalid client metadata length exceeded",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(invalidClientMetadata)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"metadata size limit exceeded","error":{"type":"API_ERROR","details":[{"field":"metadata","message":"metadata exceeds maximum allowed length of 512 characters"}],"traceId":""}}`,
		},
		{
			name: "ERROR: threeDsMethod EXTERNAL must use API mode",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					AutoConfirm: false,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodExternal,
						},
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"mode must be API when threeDsMethod is EXTERNAL"}],"traceId":""}}`,
		},
		{
			name: "ERROR: threeDsMethod EXTERNAL must have autoConfirm false",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeAPI,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodExternal,
						},
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"autoConfirm must be false when threeDsMethod is EXTERNAL"}],"traceId":""}}`,
		},
		{
			name: "ERROR: threeDsMethod EXTERNAL only supports CARD payment method",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					AutoConfirm: false,
					Mode:        constant.UnifiedPaymentModeAPI,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							Channel: "PERMATA",
						},
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodExternal,
						},
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"threeDsMethod EXTERNAL is only supported for CARD payment method"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Card amount below minimum",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    500.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodAutomatic,
						},
					},
					SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
						MerchantId:       uuid.NewString(),
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         constant.CurrencyIDR,
						PercentageAmount: 100,
						Remarks:          "test",
					}},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"amount value below the minimum"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Card amount above maximum",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    99999999999.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodAutomatic,
						},
					},
					SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
						MerchantId:       uuid.NewString(),
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         constant.CurrencyIDR,
						PercentageAmount: 100,
						Remarks:          "test",
					}},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"amount value above the maximum"}],"traceId":""}}`,
		},
		{
			name: "SUCCESS: Card amount validation skipped for recurring payment",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    500.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeAPI,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt:    time.Now().Add(time.Hour),
					RecurringID: "3dfeefca-851f-4b46-96c4-4bc9cd69692f",
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("CreateSession", constant.ValueCtxMockType(), constant.PtrCreateUnifiedPaymentSessionRequest()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
		{
			name: "SUCCESS: Card amount validation skipped for auto split card payment",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				autoSplit := true
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    500.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodAutomatic,
							AutoSplit:     &autoSplit,
						},
					},
					SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
						MerchantId:       uuid.NewString(),
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         constant.CurrencyIDR,
						PercentageAmount: 100,
						Remarks:          "test",
					}},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("CreateSession", constant.ValueCtxMockType(), constant.PtrCreateUnifiedPaymentSessionRequest()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
		{
			name: "ERROR: Card amount below minimum",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    500.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodAutomatic,
						},
					},
					SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
						MerchantId:       uuid.NewString(),
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         constant.CurrencyIDR,
						PercentageAmount: 100,
						Remarks:          "test",
					}},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"amount value below the minimum"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Card amount above maximum",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    99999999999.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodAutomatic,
						},
					},
					SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
						MerchantId:       uuid.NewString(),
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         constant.CurrencyIDR,
						PercentageAmount: 100,
						Remarks:          "test",
					}},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"amount value above the maximum"}],"traceId":""}}`,
		},
		{
			name: "SUCCESS: Card amount validation skipped for recurring payment",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    500.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeAPI,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt:    time.Now().Add(time.Hour),
					RecurringID: "3dfeefca-851f-4b46-96c4-4bc9cd69692f",
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("CreateSession", constant.ValueCtxMockType(), constant.PtrCreateUnifiedPaymentSessionRequest()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
		{
			name: "SUCCESS: Card amount validation skipped for auto split card payment",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				autoSplit := true
				request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ClientReferenceID: "123456",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    500.00,
					},
					AutoConfirm: true,
					Mode:        constant.UnifiedPaymentModeRedirect,
					RedirectUrl: unifiedPaymentModel.RedirectUrl{
						SuccessReturnUrl:    "https://success.url",
						FailureReturnUrl:    "https://failure.url",
						ExpirationReturnUrl: "https://expiration.url",
					},
					ExpiryAt: time.Now().Add(time.Hour),
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodCard,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ThreeDsMethod: constant.CardThreeDsMethodAutomatic,
							AutoSplit:     &autoSplit,
						},
					},
					SplitRoutingConfigurations: &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{{
						MerchantId:       uuid.NewString(),
						Type:             constant.SplitRoutingPaymentTypePercentage,
						Currency:         constant.CurrencyIDR,
						PercentageAmount: 100,
						Remarks:          "test",
					}},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("CreateSession", constant.ValueCtxMockType(), constant.PtrCreateUnifiedPaymentSessionRequest()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
		{
			name: "SUCCESS",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("CreateSession", constant.ValueCtxMockType(), constant.PtrCreateUnifiedPaymentSessionRequest()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			reqBody := []byte{}
			if test.requestBody != nil {
				reqBody = test.requestBody()
			}
			req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(reqBody))
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			for key, value := range test.requestHeader {
				req.Header.Set(key, value)
			}

			controller.Create(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponse, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
