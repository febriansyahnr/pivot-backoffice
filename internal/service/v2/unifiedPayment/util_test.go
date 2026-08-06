package unifiedPaymentService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	encryptionMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetFailureCodeOfMethodDetail(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)

	svc := UnifiedPaymentService{
		config:                 cfg,
		logger:                 log,
		paymentRepo:            paymentRepo,
		paymentMethodRepo:      paymentMethodRepo,
		accountTransactionRepo: accountTrxRepo,
	}

	testCases := []struct {
		name             string
		trxStatus        string
		methodDetail     *unifiedPaymentModel.ChargePaymentMethodDetails
		expectedFailCode string
	}{
		{
			name:             "Return empty when transaction status is not FAILED",
			trxStatus:        c.StatusSuccess,
			methodDetail:     &unifiedPaymentModel.ChargePaymentMethodDetails{},
			expectedFailCode: "",
		},
		{
			name:             "Return empty when methodDetail is nil",
			trxStatus:        c.StatusFailed,
			methodDetail:     nil,
			expectedFailCode: "",
		},
		{
			name:      "Return empty when Card is nil",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: nil,
			},
			expectedFailCode: "",
		},
		{
			name:      "Return CANCELLED_BY_USER when gateway code is ABORTED",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					ResponseCode: &unifiedPaymentModel.ChargePaymentMethodDetailCardResponseCode{
						GatewayCode: c.CreditCardGatewayCodeAborted,
					},
				},
			},
			expectedFailCode: c.FailureCodeCancelledByUser,
		},
		{
			name:      "Return AUTHENTICATION_FAILED when 3DS result is not successful",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthenticationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult{
						ThreeDsResult: "AUTHENTICATION_FAILED",
					},
				},
			},
			expectedFailCode: c.FailureCodeAuthenticationFailed,
		},
		{
			name:      "Return AUTHENTICATION_FAILED when 3DS result is AUTHENTICATION_REJECTED",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthenticationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult{
						ThreeDsResult: "AUTHENTICATION_REJECTED",
					},
				},
			},
			expectedFailCode: c.FailureCodeAuthenticationFailed,
		},
		{
			name:      "Return DECLINED_BY_CHANNEL for issuer code 01",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "01",
					},
				},
			},
			expectedFailCode: c.FailureCodeDeclinedByChannel,
		},
		{
			name:      "Return DECLINED_BY_CHANNEL for issuer code 03",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "03",
					},
				},
			},
			expectedFailCode: c.FailureCodeDeclinedByChannel,
		},
		{
			name:      "Return DECLINED_BY_CHANNEL for issuer code 54 (expired card)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "54",
					},
				},
			},
			expectedFailCode: c.FailureCodeDeclinedByChannel,
		},
		{
			name:      "Return DECLINED_BY_CHANNEL for issuer code 101 (expired card)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "101",
					},
				},
			},
			expectedFailCode: c.FailureCodeDeclinedByChannel,
		},
		{
			name:      "Return DECLINED_BY_CHANNEL for issuer code N7 (invalid CVV)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "N7",
					},
				},
			},
			expectedFailCode: c.FailureCodeDeclinedByChannel,
		},
		{
			name:      "Return DECLINED_BY_CHANNEL for issuer code 115",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "115",
					},
				},
			},
			expectedFailCode: c.FailureCodeDeclinedByChannel,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 14 (invalid account - first in group)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "14",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 111 (invalid account - last in group)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "111",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 04 (fraud - stolen card)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "04",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 07 (fraud)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "07",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 200 (fraud)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "200",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 34 (fraud - blocked account)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "34",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 59 (fraud)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "59",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return SUSPECTED_FRAUD for issuer code 83 (fraud)",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "83",
					},
				},
			},
			expectedFailCode: c.FailureCodeSuspectedFraud,
		},
		{
			name:      "Return INSUFFICIENT_FUND for issuer code 51",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "51",
					},
				},
			},
			expectedFailCode: c.FailureCodeInsufficientFund,
		},
		{
			name:      "Return INSUFFICIENT_FUND for issuer code 116",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "116",
					},
				},
			},
			expectedFailCode: c.FailureCodeInsufficientFund,
		},
		{
			name:      "Return INSUFFICIENT_FUND for issuer code 121",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "121",
					},
				},
			},
			expectedFailCode: c.FailureCodeInsufficientFund,
		},
		{
			name:      "Return CHANNEL_UNAVAILABLE for issuer code 19",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "19",
					},
				},
			},
			expectedFailCode: c.FailureCodeChannelUnavailable,
		},
		{
			name:      "Return CHANNEL_UNAVAILABLE for issuer code 80",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "80",
					},
				},
			},
			expectedFailCode: c.FailureCodeChannelUnavailable,
		},
		{
			name:      "Return CHANNEL_UNAVAILABLE for issuer code 91",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "91",
					},
				},
			},
			expectedFailCode: c.FailureCodeChannelUnavailable,
		},
		{
			name:      "Return CHANNEL_UNAVAILABLE for issuer code 911",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "911",
					},
				},
			},
			expectedFailCode: c.FailureCodeChannelUnavailable,
		},
		{
			name:      "Return UNKNOWN for unrecognized issuer code",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: "99",
					},
				},
			},
			expectedFailCode: c.FailureCodeUnknown,
		},
		{
			name:      "Return UNKNOWN when authorization result is nil",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					AuthorizationResult: nil,
				},
			},
			expectedFailCode: c.FailureCodeUnknown,
		},
		{
			name:      "Return UNKNOWN when Card has no error indicators",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{},
			},
			expectedFailCode: c.FailureCodeUnknown,
		},
		{
			name:      "Priority: CANCELLED_BY_USER over authentication failure",
			trxStatus: c.StatusFailed,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					ResponseCode: &unifiedPaymentModel.ChargePaymentMethodDetailCardResponseCode{
						GatewayCode: c.CreditCardGatewayCodeAborted,
					},
					AuthenticationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult{
						ThreeDsResult: "AUTHENTICATION_FAILED",
					},
				},
			},
			expectedFailCode: c.FailureCodeCancelledByUser,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.GetFailureCodeOfMethodDetail(tc.trxStatus, tc.methodDetail)
			assert.Equal(t, tc.expectedFailCode, result)
		})
	}
}

func TestOverridePartnerErrorToRateLimitIfNeeded(t *testing.T) {
	tests := []struct {
		name                   string
		inputError             error
		featureName            string
		expectedError          error
		shouldBeRateLimitError bool
		shouldPreserveOriginal bool
		skipContext            bool
	}{
		{
			name:                   "nil error returns nil",
			inputError:             nil,
			featureName:            constant.FeatureConfirmPaymentCheckoutUI,
			expectedError:          nil,
			shouldBeRateLimitError: false,
		},
		{
			name:                   "checkout UI with internal error returns rate limit error",
			inputError:             pkgErr.New(httpResponse.HttpErrInternal, errors.New("partner service unavailable")),
			featureName:            constant.FeatureConfirmPaymentCheckoutUI,
			shouldBeRateLimitError: true,
		},
		{
			name:                   "checkout UI with request timeout error returns rate limit error",
			inputError:             pkgErr.New(httpResponse.HttpErrRequestTimeout, errors.New("request timeout")),
			featureName:            constant.FeatureConfirmPaymentCheckoutUI,
			shouldBeRateLimitError: true,
		},
		{
			name:                   "checkout UI with non-internal error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrRequest, errors.New("invalid request")),
			featureName:            constant.FeatureConfirmPaymentCheckoutUI,
			shouldPreserveOriginal: true,
		},
		{
			name:                   "non-checkout UI with internal error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrInternal, errors.New("database error")),
			featureName:            "other-feature",
			shouldPreserveOriginal: true,
		},
		{
			name:                   "non-checkout UI with request timeout error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrRequestTimeout, errors.New("request timeout")),
			featureName:            "other-feature",
			shouldPreserveOriginal: true,
		},
		{
			name:                   "non-checkout UI with request error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrRequest, errors.New("validation error")),
			featureName:            "other-feature",
			shouldPreserveOriginal: true,
		},
		{
			name:                   "checkout UI with unauthorized error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrUnauthorized, errors.New("unauthorized")),
			featureName:            constant.FeatureConfirmPaymentCheckoutUI,
			shouldPreserveOriginal: true,
		},
		{
			name:                   "checkout UI with not found error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrNotFound, errors.New("resource not found")),
			featureName:            constant.FeatureConfirmPaymentCheckoutUI,
			shouldPreserveOriginal: true,
		},
		{
			name:                   "empty feature name with internal error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrInternal, errors.New("internal error")),
			featureName:            "",
			shouldPreserveOriginal: true,
		},
		{
			name:                   "nil ctx with internal error preserves original error",
			inputError:             pkgErr.New(httpResponse.HttpErrInternal, errors.New("internal error")),
			featureName:            "",
			shouldPreserveOriginal: true,
			skipContext:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if !tt.skipContext {
				ctx = context.WithValue(ctx, constant.CtxFeatureName, tt.featureName)
			}

			svc := &UnifiedPaymentService{}
			result := svc.overridePartnerErrorToRateLimitIfNeeded(ctx, tt.inputError)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, result)
			} else if tt.shouldBeRateLimitError {
				assert.NotNil(t, result)
				httpErrorType, _ := pkgErr.ExtractError(result)
				assert.Equal(t, httpResponse.HttpErrRequestLimitExceeded, httpErrorType)
			} else if tt.shouldPreserveOriginal {
				assert.Equal(t, tt.inputError, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestGetCardMIDAcquirer(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		name           string
		payment        *paymentModel.Payment
		notifReq       *unifiedPaymentModel.PaymentNotificationRequest
		mockSetup      func(pmSvc *serviceMock.IPaymentMethodService)
		expectedResult string
		expectedError  error
	}{
		{
			name: "returns error when FindPaymentMethodByIdAndMerchant fails",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							MID: "MID123456",
						},
					},
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(nil, errors.New("database error"))
			},
			expectedResult: "",
			expectedError:  errors.New("database error"),
		},
		{
			name: "returns empty string when payment method is nil",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							MID: "MID123456",
						},
					},
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(nil, nil)
			},
			expectedResult: "",
			expectedError:  nil,
		},
		{
			name: "returns empty string when MerchantConfigObj is nil",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							MID: "MID123456",
						},
					},
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(&paymentModel.PaymentMethodWithPivot{
						MerchantConfigObj: nil,
					}, nil)
			},
			expectedResult: "",
			expectedError:  nil,
		},
		{
			name: "returns acquirer when MID is found in config",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							MID: "MID123456",
						},
					},
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(&paymentModel.PaymentMethodWithPivot{
						MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
							PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
								Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
									Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
										{
											AcquirerMerchantID: "MID123456",
											Acquirer:           "BCA",
										},
										{
											AcquirerMerchantID: "MID789012",
											Acquirer:           "MANDIRI",
										},
									},
								},
							},
						},
					}, nil)
			},
			expectedResult: "BCA",
			expectedError:  nil,
		},
		{
			name: "returns empty string when MID is not found in config",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							MID: "MID_UNKNOWN",
						},
					},
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(&paymentModel.PaymentMethodWithPivot{
						MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
							PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
								Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
									Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
										{
											AcquirerMerchantID: "MID123456",
											Acquirer:           "BCA",
										},
									},
								},
							},
						},
					}, nil)
			},
			expectedResult: "",
			expectedError:  nil,
		},
		{
			name: "returns empty string when PartnerConfig is nil",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							MID: "MID123456",
						},
					},
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(&paymentModel.PaymentMethodWithPivot{
						MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
							PartnerConfig: nil,
						},
					}, nil)
			},
			expectedResult: "",
			expectedError:  nil,
		},
		{
			name: "returns empty string when Card config is nil",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							MID: "MID123456",
						},
					},
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(&paymentModel.PaymentMethodWithPivot{
						MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
							PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
								Card: nil,
							},
						},
					}, nil)
			},
			expectedResult: "",
			expectedError:  nil,
		},
		{
			name: "returns empty string when Card MIDInfo is nil (empty MID from GetCardMID)",
			payment: &paymentModel.Payment{
				UUID:            "payment-uuid-123",
				PaymentMethodID: "pm-id-123",
				MerchantID:      "merchant-id-123",
			},
			notifReq: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: nil,
				},
			},
			mockSetup: func(pmSvc *serviceMock.IPaymentMethodService) {
				pmSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-id-123", "merchant-id-123").
					Return(&paymentModel.PaymentMethodWithPivot{
						MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
							PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
								Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
									Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
										{
											AcquirerMerchantID: "MID123456",
											Acquirer:           "BCA",
										},
									},
								},
							},
						},
					}, nil)
			},
			expectedResult: "",
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pmSvc := serviceMock.NewIPaymentMethodService(t)
			tc.mockSetup(pmSvc)

			svc := &UnifiedPaymentService{
				config:           cfg,
				logger:           log,
				paymentMethodSvc: pmSvc,
			}

			result, err := svc.GetCardMIDAcquirer(context.Background(), tc.payment, tc.notifReq)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestPrepareCardAuthentication(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		name                      string
		cardAuthenticationRequest *unifiedPaymentModel.CardAuthenticationRequest
		mockCreditCardSvc         func(svc *serviceMock.ICreditCardService)
		mockCryptoProvider        func(cp *encryptionMock.CryptoProvider)
		expectedAuthRequest       creditcardCoreProcessorModel.AuthenticationRequest
		expectedError             error
	}{
		{
			name: "SUCCESS: prepare card authentication",
			cardAuthenticationRequest: &unifiedPaymentModel.CardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-123",
				ClientTransactionID: "ctx-txn-123",
				Amount:              100000,
				Currency:            "IDR",
				Card: &unifiedPaymentModel.CardAuthenticationRequestCard{
					Fingerprint: "fp-abc123",
				},
			},
			mockCreditCardSvc: func(svc *serviceMock.ICreditCardService) {
				svc.On("GetCardEncryptionPublicKey", mock.Anything, "merchant-123").
					Return([]byte("cert-pem-bytes"), nil)
			},
			mockCryptoProvider: func(cp *encryptionMock.CryptoProvider) {
				cp.On("EncryptPKCS7", []byte("cert-pem-bytes"), mock.AnythingOfType("[]uint8")).
					Return("encrypted-payload-string", nil)
			},
			expectedAuthRequest: creditcardCoreProcessorModel.AuthenticationRequest{
				MerchantID:       "merchant-123",
				PaymentID:        "payment-123",
				EncryptedPayload: "encrypted-payload-string",
			},
			expectedError: nil,
		},
		{
			name: "ERROR: GetCardEncryptionPublicKey fails",
			cardAuthenticationRequest: &unifiedPaymentModel.CardAuthenticationRequest{
				PaymentID:  "payment-456",
				MerchantID: "merchant-456",
			},
			mockCreditCardSvc: func(svc *serviceMock.ICreditCardService) {
				svc.On("GetCardEncryptionPublicKey", mock.Anything, "merchant-456").
					Return(nil, errors.New("public key not found"))
			},
			mockCryptoProvider:  func(cp *encryptionMock.CryptoProvider) {},
			expectedAuthRequest: creditcardCoreProcessorModel.AuthenticationRequest{},
			expectedError:       errors.New("public key not found"),
		},
		{
			name: "ERROR: EncryptPKCS7 fails",
			cardAuthenticationRequest: &unifiedPaymentModel.CardAuthenticationRequest{
				PaymentID:  "payment-789",
				MerchantID: "merchant-789",
			},
			mockCreditCardSvc: func(svc *serviceMock.ICreditCardService) {
				svc.On("GetCardEncryptionPublicKey", mock.Anything, "merchant-789").
					Return([]byte("cert-pem-bytes"), nil)
			},
			mockCryptoProvider: func(cp *encryptionMock.CryptoProvider) {
				cp.On("EncryptPKCS7", []byte("cert-pem-bytes"), mock.AnythingOfType("[]uint8")).
					Return("", errors.New("encryption failed"))
			},
			expectedAuthRequest: creditcardCoreProcessorModel.AuthenticationRequest{},
			expectedError:       errors.New("encryption failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			creditcardSvc := serviceMock.NewICreditCardService(t)
			cryptoProvider := encryptionMock.NewCryptoProvider(t)

			tc.mockCreditCardSvc(creditcardSvc)
			tc.mockCryptoProvider(cryptoProvider)

			svc := &UnifiedPaymentService{
				config:         cfg,
				logger:         log,
				creditcardSvc:  creditcardSvc,
				cryptoProvider: cryptoProvider,
			}

			ctx, authRequest, err := svc.PrepareCardAuthentication(context.Background(), tc.cardAuthenticationRequest)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, authRequest)
				assert.Equal(t, tc.expectedAuthRequest.MerchantID, authRequest.MerchantID)
				assert.Equal(t, tc.expectedAuthRequest.PaymentID, authRequest.PaymentID)
				assert.Equal(t, tc.expectedAuthRequest.EncryptedPayload, authRequest.EncryptedPayload)
			}

			// verify context is enriched with CtxClientReqKey
			assert.NotNil(t, ctx)
		})
	}
}
