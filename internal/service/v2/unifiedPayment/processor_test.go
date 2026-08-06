package unifiedPaymentService

import (
	"context"
	"errors"
	"testing"
	"time"

	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/config"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	snapCoreQrModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreVaModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	snapVa "github.com/paper-indonesia/pdk/go/snap/structs/va"
	pdkUtil "github.com/paper-indonesia/pdk/go/util"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func TestUnifiedPaymentServiceInitEwalletPaymentLink(t *testing.T) {
	danaSuccessResponse := &ewallet.EwalletPaymentLinkResponse{
		ResponseCode:       "2001600",
		ResponseMessage:    "Successful",
		PartnerReferenceNo: "20240101120000001",
		ReferenceNo:        "TXN0001",
		WebRedirectionURL:  "https://m.sandbox.dana.id/n/cashier/new/checkout", // NOSONAR
	}
	shopeePaySuccessResponse := &ewallet.EwalletPaymentLinkResponse{
		ResponseCode:       "2001600",
		ResponseMessage:    "Successful",
		PartnerReferenceNo: "20240101120000001",
		ReferenceNo:        "TXN0001",
		WebRedirectionURL:  "https://app.uat.shopeepay.co.id/u/pay_checkout", // NOSONAR
	}

	testCases := []struct {
		name      string
		request   *unifiedPaymentModel.BaseProcessorRequest
		wantError bool
		setupMock func(mockSnapCore *mockRepo.ISnapCoreRepository, shortLinkSvc *mockService.IShortLinkService)
		validate  func(t *testing.T, result *unifiedPaymentModel.ChargePaymentMethodDetails)
	}{
		{
			name: "SUCCESS:Create ewallet DANA payment link",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-123", // NOSONAR
				ClientReferenceID: "ref-123",     // NOSONAR
				ChargeID:          "charge-123",  // NOSONAR
				Amount: unifiedPaymentModel.Amount{
					Value:    10000.0, // NOSONAR
					Currency: "IDR",   // NOSONAR
				},
				Mode:               constant.UnifiedPaymentModeAPI,
				ExpiryAt:           time.Now().Add(15 * time.Minute),
				SuccessRedirectUrl: "https://payment.example.com/callback", // NOSONAR
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: constant.UnifiedPaymentEWalletDanaAcquirer,
					},
				},
				PaymentPartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
						SubMerchantID: "123456789", // NOSONAR
					},
				},
			},
			wantError: false,
			setupMock: func(mockSnapCore *mockRepo.ISnapCoreRepository, shortLinkSvc *mockService.IShortLinkService) {
				shortLinkSvc.On("Create", mock.Anything, mock.Anything).Return(&shortLinkModel.ShortLink{ShortLinkURL: ""}, nil).Once()
				mockSnapCore.On(
					"CreateEWalletPaymentLink", mock.Anything, mock.MatchedBy(func(req *ewallet.EwalletPaymentRequest) bool {
						return req.SubMerchantId == "123456789" && req.MerchantId == "" && req.ExternalStoreId == "" // NOSONAR
					}),
				).Return(danaSuccessResponse, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.ChargePaymentMethodDetails) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Ewallet)
				assert.Equal(t, constant.UnifiedPaymentEWalletDanaAcquirer, result.Ewallet.Channel)
				assert.Equal(t, "https://m.sandbox.dana.id/n/cashier/new/checkout", result.Ewallet.WebRedirectURL)
				assert.Equal(t, "TXN0001", result.Ewallet.ReferenceNo)
				assert.Equal(t, "20240101120000001", result.Ewallet.PartnerReferenceNo)
			},
		},
		{
			name: "SUCCESS:Create ewallet SHOPEEPAY payment link",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-123", // NOSONAR
				ClientReferenceID: "ref-123",     // NOSONAR
				ChargeID:          "charge-123",  // NOSONAR
				Amount: unifiedPaymentModel.Amount{
					Value:    10000.0, // NOSONAR
					Currency: "IDR",   // NOSONAR
				},
				ExpiryAt:           time.Now().Add(15 * time.Minute),
				SuccessRedirectUrl: "https://payment.example.com/callback", // NOSONAR
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: constant.UnifiedPaymentEWalletShopeePayAcquirer,
					},
				},
				PaymentPartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
						ExternalMerchantID: "HRSSPP01-00001",   // NOSONAR
						ExternalStoreID:    "HRSSPP01-S-00009", // NOSONAR
					},
				},
			},
			wantError: false,
			setupMock: func(mockSnapCore *mockRepo.ISnapCoreRepository, shortLinkSvc *mockService.IShortLinkService) {
				mockSnapCore.On(
					"CreateEWalletPaymentLink", mock.Anything, mock.MatchedBy(func(req *ewallet.EwalletPaymentRequest) bool {
						return req.MerchantId == "HRSSPP01-00001" && req.ExternalStoreId == "HRSSPP01-S-00009" && req.SubMerchantId == "" // NOSONAR
					}),
				).Return(shopeePaySuccessResponse, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.ChargePaymentMethodDetails) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Ewallet)
				assert.Equal(t, constant.UnifiedPaymentEWalletShopeePayAcquirer, result.Ewallet.Channel)
				assert.Equal(t, "https://app.uat.shopeepay.co.id/u/pay_checkout", result.Ewallet.WebRedirectURL)
				assert.Equal(t, "TXN0001", result.Ewallet.ReferenceNo)
				assert.Equal(t, "20240101120000001", result.Ewallet.PartnerReferenceNo)
			},
		},
		{
			name: "ERROR: snap core returns error",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-123", // NOSONAR
				ClientReferenceID: "ref-123",     // NOSONAR
				ChargeID:          "charge-123",  // NOSONAR
				Amount: unifiedPaymentModel.Amount{
					Value:    10000.0, // NOSONAR
					Currency: "IDR",   // NOSONAR
				},
				ExpiryAt:           time.Now().Add(15 * time.Minute),
				SuccessRedirectUrl: "https://payment.example.com/callback", // NOSONAR
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: "SHOPEEPAY", // NOSONAR
					},
				},
			},
			wantError: true,
			setupMock: func(mockSnapCore *mockRepo.ISnapCoreRepository, shortLinkSvc *mockService.IShortLinkService) {
				mockSnapCore.On(
					"CreateEWalletPaymentLink",
					mock.Anything,
					mock.AnythingOfType("*ewallet.EwalletPaymentRequest"),
				).Return(nil, errors.New("snap core error"))
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.ChargePaymentMethodDetails) {
				assert.Nil(t, result)
			},
		},
		{
			name: "SUCCESS: with lowercase channel",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-456", // NOSONAR
				ClientReferenceID: "ref-456",     // NOSONAR
				ChargeID:          "charge-456",  // NOSONAR
				Amount: unifiedPaymentModel.Amount{
					Value:    25000.0,
					Currency: "IDR",
				},
				ExpiryAt:           time.Now().Add(30 * time.Minute),
				SuccessRedirectUrl: "https://payment.example.com/callback",
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: "shopeepay",
					},
				},
			},
			wantError: false,
			setupMock: func(mockSnapCore *mockRepo.ISnapCoreRepository, shortLinkSvc *mockService.IShortLinkService) {
				mockSnapCore.On(
					"CreateEWalletPaymentLink",
					mock.Anything,
					mock.MatchedBy(func(req *ewallet.EwalletPaymentRequest) bool {
						return req.Acquirer == "SHOPEEPAY"
					}),
				).Return(shopeePaySuccessResponse, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.ChargePaymentMethodDetails) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Ewallet)
				assert.Equal(t, "SHOPEEPAY", result.Ewallet.Channel)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := logger.NewZapLogger(logger.Config{})
			mockSnapCore := mockRepo.NewISnapCoreRepository(t)
			shortLinkSvc := mockService.NewIShortLinkService(t)

			tc.setupMock(mockSnapCore, shortLinkSvc)

			cfg := &config.Config{}
			service := &UnifiedPaymentService{
				config:       cfg,
				logger:       log,
				snapCoreRepo: mockSnapCore,
				shortLinkSvc: shortLinkSvc,
			}

			result, err := service.initEwalletPaymentLink(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, result)
			mockSnapCore.AssertExpectations(t)
		})
	}
}

func TestInitEncryptedCardAuthentication(t *testing.T) {
	tests := []struct {
		name              string
		request           *unifiedPaymentModel.BaseProcessorRequest
		mockPayment       *paymentModel.Payment
		mockPaymentError  error
		mockRedisSetError error
		mockAuthResponse  *card.EncryptedCardAuthenticationResponse
		mockAuthError     error
		expectedResult    *unifiedPaymentModel.ChargePaymentMethodDetails
		expectedError     string
	}{
		{
			name: "success",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-123",
				MerchantID:        "merchant-456",
				ClientReferenceID: "client-789",
				Amount:            unifiedPaymentModel.Amount{Value: 1000.50, Currency: "IDR"},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					CardPaymentMethodDetail: &unifiedPaymentModel.CardPaymentMethodDetail{
						EncryptedCard: "encrypted-card-123",
						CVC:           "123",
					},
				},
			},
			mockPayment: &paymentModel.Payment{
				UUID:       "payment-123",
				MerchantID: "merchant-456",
				Amount:     decimal.New(100, 2),
				Currency:   "IDR",
				Status:     constant.StatusPending,
				Metadata: &map[string]any{
					"orderInformation": map[string]any{
						"billingInfo": map[string]any{
							"phoneNumber": map[string]any{
								"countryCode": "+62",
								"number":      "0812345678",
							},
						},
					},
				},
			},
			mockPaymentError:  nil,
			mockRedisSetError: nil,
			mockAuthResponse: &card.EncryptedCardAuthenticationResponse{
				CardID: "card-123",
				CardInfo: card.EncryptedCardInformationResponse{
					First6Digits: "123456",
					Last4Digits:  "9876",
					First8Digits: "12345678",
					ExpiryMonth:  "12",
					ExpiryYear:   "25",
					Fingerprint:  "fingerprint-123",
				},
				Bin: card.Bin{
					CardType:      "CREDIT",
					IssuerName:    "Bank ABC",
					CardBrand:     "VISA",
					IssuerCountry: "ID",
				},
				AuthenticationResponse: card.AuthenticationResponse{
					Status:  constant.CreditCardProcessorStatusPending,
					Message: "Authentication pending",
				},
			},
			mockAuthError: nil,
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				ProcessorReference: constant.CreditCardCoreProcessor,
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					First6:      "123456",
					Last4:       "9876",
					First8:      "12345678",
					ExpMonth:    "12",
					ExpYear:     "25",
					Fingerprint: "fingerprint-123",
					BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
						Type:        "CREDIT",
						IssuingBank: "Bank ABC",
						Brand:       "VISA",
						Country:     "ID",
					},
				},
			},
			expectedError: "",
		},
		{
			name: "payment not found",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:  "invalid-payment",
				MerchantID: "merchant-456",
			},
			mockPayment:      nil,
			mockPaymentError: errors.New("payment not found"),
			expectedResult:   nil,
			expectedError:    "error when get data",
		},
		{
			name: "redis cache error",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-456",
			},
			mockPayment: &paymentModel.Payment{
				UUID:       "payment-123",
				MerchantID: "merchant-456",
			},
			mockPaymentError:  nil,
			mockRedisSetError: errors.New("redis connection failed"),
			expectedResult:    nil,
			expectedError:     "error store in cache",
		},
		{
			name: "authentication service error",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-123",
				MerchantID:        "merchant-456",
				ClientReferenceID: "client-789",
				Amount:            unifiedPaymentModel.Amount{Value: 1000.50, Currency: "IDR"},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					CardPaymentMethodDetail: &unifiedPaymentModel.CardPaymentMethodDetail{
						EncryptedCard: "encrypted-card-123",
						CVC:           "123",
					},
				},
			},
			mockPayment: &paymentModel.Payment{
				UUID:       "payment-123",
				MerchantID: "merchant-456",
			},
			mockPaymentError:  nil,
			mockRedisSetError: nil,
			mockAuthResponse:  nil,
			mockAuthError:     errors.New("authentication service unavailable"),
			expectedResult:    nil,
			expectedError:     "authentication service unavailable",
		},
		{
			name: "authentication failed status",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-123",
				MerchantID:        "merchant-456",
				ClientReferenceID: "client-789",
				Amount:            unifiedPaymentModel.Amount{Value: 1000.50, Currency: "IDR"},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					CardPaymentMethodDetail: &unifiedPaymentModel.CardPaymentMethodDetail{
						EncryptedCard: "encrypted-card-123",
						CVC:           "123",
					},
				},
			},
			mockPayment: &paymentModel.Payment{
				UUID:       "payment-123",
				MerchantID: "merchant-456",
			},
			mockPaymentError:  nil,
			mockRedisSetError: nil,
			mockAuthResponse: &card.EncryptedCardAuthenticationResponse{
				AuthenticationResponse: card.AuthenticationResponse{
					Status:  "FAILED",
					Message: "Invalid card details",
				},
			},
			mockAuthError:  nil,
			expectedResult: nil,
			expectedError:  "failed to create encrypted card authentication link: Invalid card details",
		},
		{
			name: "success with 3DS authentication data",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-3ds",
				MerchantID:        "merchant-456",
				ClientReferenceID: "client-3ds",
				DerivedMerchantID: "derived-merchant-id",
				Amount:            unifiedPaymentModel.Amount{Value: 5000.00, Currency: "IDR"},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					CardPaymentMethodDetail: &unifiedPaymentModel.CardPaymentMethodDetail{
						EncryptedCard: "encrypted-card-3ds",
						CVC:           "456",
					},
				},
				RecurringID: "11700e40-83af-4f45-96fa-eebb98efa34f",
			},
			mockPayment: &paymentModel.Payment{
				UUID:       "payment-3ds",
				MerchantID: "merchant-456",
				Amount:     decimal.New(500, 1),
				Currency:   "IDR",
				Status:     constant.StatusPending,
			},
			mockPaymentError:  nil,
			mockRedisSetError: nil,
			mockAuthResponse: &card.EncryptedCardAuthenticationResponse{
				CardID: "card-3ds",
				CardInfo: card.EncryptedCardInformationResponse{
					First6Digits: "555666",
					Last4Digits:  "1234",
					First8Digits: "55566677",
					ExpiryMonth:  "03",
					ExpiryYear:   "26",
					Fingerprint:  "fingerprint-3ds",
				},
				Bin: card.Bin{
					CardType:      "DEBIT",
					IssuerName:    "Bank DEF",
					CardBrand:     "MASTERCARD",
					IssuerCountry: "ID",
				},
				AuthenticationResponse: card.AuthenticationResponse{
					Status:  constant.CreditCardProcessorStatusPending,
					Message: "3DS Authentication required",
					AuthenticationURL: card.AuthenticationURLDetail{
						URL: "https://3ds.example.com/authenticate",
					},
					AuthenticationData: &card.EncryptedCardAuthenticationData{
						ThreeDsVer:           "2.0",
						AuthenticationResult: "Y",
						EciCode:              "05",
					},
					AcquirerTransactionID: "acquirer-txn-3ds",
				},
			},
			mockAuthError: nil,
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				ProcessorReference:   constant.CreditCardCoreProcessor,
				ProcessorReferenceID: "acquirer-txn-3ds",
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					First6:      "555666",
					Last4:       "1234",
					First8:      "55566677",
					ExpMonth:    "03",
					ExpYear:     "26",
					Fingerprint: "fingerprint-3ds",
					BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
						Type:        "DEBIT",
						IssuingBank: "Bank DEF",
						Brand:       "MASTERCARD",
						Country:     "ID",
					},
					ACSURL: "https://3ds.example.com/authenticate",
					AuthenticationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult{
						ThreeDsVersion: "2.0",
						ThreeDsResult:  "Y",
						EciCode:        "05",
					},
				},
			},
			expectedError: "",
		},
		{
			name: "success with threeDsMethod set in PaymentMethodOptions",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-with-3ds-method",
				MerchantID:        "merchant-456",
				ClientReferenceID: "client-ref-123",
				Amount:            unifiedPaymentModel.Amount{Value: 2000.00, Currency: "IDR"},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					CardPaymentMethodDetail: &unifiedPaymentModel.CardPaymentMethodDetail{
						EncryptedCard: "encrypted-card-456",
						CVC:           "789",
					},
				},
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					Card: &unifiedPaymentModel.PaymentMethodOptionCard{
						ThreeDsMethod: constant.CardThreeDsMethodNever,
					},
				},
			},
			mockPayment: &paymentModel.Payment{
				UUID:       "payment-with-3ds-method",
				MerchantID: "merchant-456",
				Amount:     decimal.New(200, 1),
				Currency:   "IDR",
				Status:     constant.StatusPending,
			},
			mockPaymentError:  nil,
			mockRedisSetError: nil,
			mockAuthResponse: &card.EncryptedCardAuthenticationResponse{
				CardID: "card-456",
				CardInfo: card.EncryptedCardInformationResponse{
					First6Digits: "424242",
					Last4Digits:  "4242",
					First8Digits: "42424242",
					ExpiryMonth:  "06",
					ExpiryYear:   "27",
					Fingerprint:  "fingerprint-456",
				},
				Bin: card.Bin{
					CardType:      "CREDIT",
					IssuerName:    "Test Bank",
					CardBrand:     "VISA",
					IssuerCountry: "ID",
				},
				AuthenticationResponse: card.AuthenticationResponse{
					Status:  constant.CreditCardProcessorStatusPending,
					Message: "Authentication pending",
				},
			},
			mockAuthError: nil,
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				ProcessorReference: constant.CreditCardCoreProcessor,
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					First6:      "424242",
					Last4:       "4242",
					First8:      "42424242",
					ExpMonth:    "06",
					ExpYear:     "27",
					Fingerprint: "fingerprint-456",
					BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
						Type:        "CREDIT",
						IssuingBank: "Test Bank",
						Brand:       "VISA",
						Country:     "ID",
					},
				},
			},
			expectedError: "",
		},
		{
			name: "success with external 3DS info in PaymentMethodOptions",
			request: &unifiedPaymentModel.BaseProcessorRequest{
				PaymentID:         "payment-external-3ds",
				MerchantID:        "merchant-789",
				ClientReferenceID: "client-external-3ds",
				Amount:            unifiedPaymentModel.Amount{Value: 3000.00, Currency: "IDR"},
				ExpiryAt:          time.Now().Add(time.Hour),
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					CardPaymentMethodDetail: &unifiedPaymentModel.CardPaymentMethodDetail{
						EncryptedCard: "encrypted-card-external",
						CVC:           "321",
					},
				},
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					Card: &unifiedPaymentModel.PaymentMethodOptionCard{
						ThreeDsMethod: constant.CardThreeDsMethodExternal,
						ThreeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
							TransactionID:        "550e8400-e29b-41d4-a716-446655440000",
							ThreeDSVersion:       "2.1.0",
							ECI:                  "05",
							TransactionStatus:    "Y",
							AuthenticationScheme: "VISA",
							ACSTransactionID:     "8a880dc0-d2d2-4067-bcb1-b08d1690b26e",
							ACSReference:         "ACS-REF-123456",
							MCC:                  "7699",
							BankMerchantId:       "1234321",
						},
					},
				},
			},
			mockPayment: &paymentModel.Payment{
				UUID:       "payment-external-3ds",
				MerchantID: "merchant-789",
				Amount:     decimal.New(300, 1),
				Currency:   "IDR",
				Status:     constant.StatusPending,
			},
			mockPaymentError:  nil,
			mockRedisSetError: nil,
			mockAuthResponse: &card.EncryptedCardAuthenticationResponse{
				CardID: "card-external-3ds",
				CardInfo: card.EncryptedCardInformationResponse{
					First6Digits: "411111",
					Last4Digits:  "1111",
					First8Digits: "41111111",
					ExpiryMonth:  "09",
					ExpiryYear:   "28",
					Fingerprint:  "fingerprint-external-3ds",
				},
				Bin: card.Bin{
					CardType:      "CREDIT",
					IssuerName:    "External Bank",
					CardBrand:     "VISA",
					IssuerCountry: "ID",
				},
				AuthenticationResponse: card.AuthenticationResponse{
					Status:                constant.CreditCardProcessorStatusPending,
					Message:               "External 3DS completed",
					AcquirerTransactionID: "acquirer-external-3ds",
				},
			},
			mockAuthError: nil,
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				ProcessorReference:   constant.CreditCardCoreProcessor,
				ProcessorReferenceID: "acquirer-external-3ds",
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					First6:      "411111",
					Last4:       "1111",
					First8:      "41111111",
					ExpMonth:    "09",
					ExpYear:     "28",
					Fingerprint: "fingerprint-external-3ds",
					BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
						Type:        "CREDIT",
						IssuingBank: "External Bank",
						Brand:       "VISA",
						Country:     "ID",
					},
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := &mockRepo.IPaymentRepository{}
			mockRedis := &redisMocks.IRedisExt{}
			mockCreditcardSvc := &mockService.ICreditCardService{}

			service := &UnifiedPaymentService{
				config:        &config.Config{},
				logger:        logger.NewSlogger(logger.Config{}),
				paymentRepo:   mockPaymentRepo,
				redis:         mockRedis,
				creditcardSvc: mockCreditcardSvc,
			}

			ctx := context.Background()

			// Mock payment repository
			mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, tt.request.PaymentID, tt.request.MerchantID).
				Return(tt.mockPayment, tt.mockPaymentError)

			// Mock redis only if payment is found
			if tt.mockPaymentError == nil {

				merchantID := tt.request.MerchantID
				if tt.request.DerivedMerchantID != "" {
					merchantID = tt.request.DerivedMerchantID
				}

				cacheKey := fmt.Sprintf(constant.TemporaryPaymentRecordCacheKey, merchantID, tt.request.PaymentID)
				statusCmd := &redis.StatusCmd{}
				if tt.mockRedisSetError != nil {
					statusCmd.SetErr(tt.mockRedisSetError)
				} else {
					statusCmd.SetVal("OK")
				}

				mockRedis.On("Set", mock.Anything, cacheKey, mock.Anything, constant.TemporaryPaymentRecordTTL).
					Return(statusCmd)

				// Mock creditcard service only if redis is successful
				if tt.mockRedisSetError == nil {
					// Capture the payload to verify ThreeDsMethod and ExternalThreeDsInfo if provided
					var capturedPayload *card.EncryptedCardAuthenticationRequest
					mockCreditcardSvc.On("CreateEncryptedCardAuthenticationLink", mock.Anything, mock.AnythingOfType("*card.EncryptedCardAuthenticationRequest")).
						Run(func(args mock.Arguments) {
							capturedPayload = args.Get(1).(*card.EncryptedCardAuthenticationRequest)
						}).
						Return(tt.mockAuthResponse, tt.mockAuthError)

					// After the function call, verify the payload if PaymentMethodOptions.Card is provided
					defer func() {
						if tt.request.PaymentMethodOptions != nil && tt.request.PaymentMethodOptions.Card != nil {
							assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsMethod, capturedPayload.ThreeDsMethod)
							if tt.request.PaymentMethodOptions.Card.ThreeDsMethod == constant.CardThreeDsMethodExternal &&
								tt.request.PaymentMethodOptions.Card.ThreeDsInfo != nil {
								assert.NotNil(t, capturedPayload.ExternalThreeDsInfo)
								assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsInfo.TransactionID, capturedPayload.ExternalThreeDsInfo.TransactionID)
								assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsInfo.ThreeDSVersion, capturedPayload.ExternalThreeDsInfo.ThreeDSVersion)
								assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsInfo.ECI, capturedPayload.ExternalThreeDsInfo.ECI)
								assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsInfo.TransactionStatus, capturedPayload.ExternalThreeDsInfo.TransactionStatus)
								assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsInfo.AuthenticationScheme, capturedPayload.ExternalThreeDsInfo.AuthenticationScheme)
								assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsInfo.ACSTransactionID, capturedPayload.ExternalThreeDsInfo.ACSTransactionID)
								assert.Equal(t, tt.request.PaymentMethodOptions.Card.ThreeDsInfo.ACSReference, capturedPayload.ExternalThreeDsInfo.ACSReference)
							}
						}
					}()
				}
			}

			result, err := service.initEncryptedCardAuthentication(ctx, tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.ProcessorReference, result.ProcessorReference)
				if tt.expectedResult.ProcessorReferenceID != "" {
					assert.Equal(t, tt.expectedResult.ProcessorReferenceID, result.ProcessorReferenceID)
				}
				if tt.expectedResult.Card != nil {
					assert.NotNil(t, result.Card)
					assert.Equal(t, tt.expectedResult.Card.First6, result.Card.First6)
					assert.Equal(t, tt.expectedResult.Card.Last4, result.Card.Last4)
					assert.Equal(t, tt.expectedResult.Card.First8, result.Card.First8)
					assert.Equal(t, tt.expectedResult.Card.ExpMonth, result.Card.ExpMonth)
					assert.Equal(t, tt.expectedResult.Card.ExpYear, result.Card.ExpYear)
					assert.Equal(t, tt.expectedResult.Card.Fingerprint, result.Card.Fingerprint)
					assert.Equal(t, tt.expectedResult.Card.BinInformations.Type, result.Card.BinInformations.Type)
					assert.Equal(t, tt.expectedResult.Card.BinInformations.IssuingBank, result.Card.BinInformations.IssuingBank)
					assert.Equal(t, tt.expectedResult.Card.BinInformations.Brand, result.Card.BinInformations.Brand)
					assert.Equal(t, tt.expectedResult.Card.BinInformations.Country, result.Card.BinInformations.Country)
					if tt.expectedResult.Card.ACSURL != "" {
						assert.Equal(t, tt.expectedResult.Card.ACSURL, result.Card.ACSURL)
					}
					if tt.expectedResult.Card.AuthenticationResult != nil {
						assert.NotNil(t, result.Card.AuthenticationResult)
						assert.Equal(t, tt.expectedResult.Card.AuthenticationResult.ThreeDsVersion, result.Card.AuthenticationResult.ThreeDsVersion)
						assert.Equal(t, tt.expectedResult.Card.AuthenticationResult.ThreeDsResult, result.Card.AuthenticationResult.ThreeDsResult)
						assert.Equal(t, tt.expectedResult.Card.AuthenticationResult.EciCode, result.Card.AuthenticationResult.EciCode)
					}
				}
			}

			mockPaymentRepo.AssertExpectations(t)
			if tt.mockPaymentError == nil {
				mockRedis.AssertExpectations(t)
				if tt.mockRedisSetError == nil {
					mockCreditcardSvc.AssertExpectations(t)
				}
			}
		})
	}
}

func TestGetQRSnapBasicPayload(t *testing.T) {
	log, _ := logger.NewZapLogger(logger.Config{})

	testCases := []struct {
		name             string
		request          *unifiedPaymentModel.InitProcessorQRISRequest
		mockQrisResponse *qris.Registration
		mockQrisError    error
		wantError        bool
		validateRequest  func(t *testing.T, result snapCoreQrModel.GenerateQrMpmRequest)
		validateError    func(t *testing.T, err error)
	}{
		{
			name: "SUCCESS: With partner config provided",
			request: &unifiedPaymentModel.InitProcessorQRISRequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					ClientReferenceID: "ref-123",
					ChargeID:          "charge-123",
					Amount: unifiedPaymentModel.Amount{
						Value:    50000.0,
						Currency: "IDR",
					},
					IsStaticPayment: false,
					PaymentPartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							AcquirerMerchantID: "MERCHANT-123",
							AcquirerTerminalID: "TERMINAL-123",
							Acquirer:           "NOBU",
						},
					},
				},
				ExpiryAt: time.Now().Add(15 * time.Minute),
			},
			wantError: false,
			validateRequest: func(t *testing.T, result snapCoreQrModel.GenerateQrMpmRequest) {
				assert.Equal(t, "ref-123", result.PartnerReferenceNo)
				assert.Equal(t, "50000.00", result.Amount.Value)
				assert.Equal(t, "IDR", result.Amount.Currency)
				assert.Equal(t, constant.QrTypeDynamic, result.QrType)
				assert.Equal(t, "MERCHANT-123", result.SubMerchantID)
				assert.Equal(t, "TERMINAL-123", result.TerminalID)
				assert.Equal(t, "MERCHANT-123", result.MerchantID)
				assert.Equal(t, "NOBU", result.Acquirer)
				assert.Greater(t, result.ValidityPeriod, 0)
				assert.Contains(t, result.AdditionalInfo, constant.ProcessorExternalIDKey)
				assert.Equal(t, "charge-123", result.AdditionalInfo[constant.ProcessorExternalIDKey])
			},
		},
		{
			name: "SUCCESS: Static payment with partner config",
			request: &unifiedPaymentModel.InitProcessorQRISRequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					ClientReferenceID: "ref-456",
					ChargeID:          "charge-456",
					Amount: unifiedPaymentModel.Amount{
						Value:    25000.0,
						Currency: "IDR",
					},
					IsStaticPayment: true,
					PaymentPartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							AcquirerMerchantID: "STATIC-MERCHANT-456",
							AcquirerTerminalID: "STATIC-TERMINAL-456",
							Acquirer:           "DANA",
						},
					},
				},
				ExpiryAt: time.Now().Add(30 * time.Minute),
			},
			wantError: false,
			validateRequest: func(t *testing.T, result snapCoreQrModel.GenerateQrMpmRequest) {
				assert.Equal(t, "ref-456", result.PartnerReferenceNo)
				assert.Equal(t, "25000.00", result.Amount.Value)
				assert.Equal(t, constant.QrTypeStatic, result.QrType)
				assert.Equal(t, 0, result.ValidityPeriod)
				assert.Equal(t, "STATIC-MERCHANT-456", result.SubMerchantID)
				assert.Equal(t, "DANA", result.Acquirer)
			},
		},
		{
			name: "SUCCESS: Dynamic with QR registration - merchant type",
			request: &unifiedPaymentModel.InitProcessorQRISRequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					ClientReferenceID:  "ref-789",
					ChargeID:           "charge-789",
					MerchantExternalID: "external-789",
					Amount: unifiedPaymentModel.Amount{
						Value:    75000.0,
						Currency: "IDR",
					},
					IsStaticPayment: false,
				},
				ExpiryAt: time.Now().Add(10 * time.Minute),
			},
			mockQrisResponse: &qris.Registration{
				Id:                       "qr-reg-123",
				ExternalId:               "external-789",
				MerchantType:             constant.QrMerchantTypeMerchant,
				Acquirer:                 "NOBU",
				AcquirerMerchantId:       &[]string{"ACQ-MERCHANT-789"}[0],
				AcquirerTerminalId:       &[]string{"ACQ-TERMINAL-789"}[0],
				AcquirerParentMerchantId: "PARENT-MERCHANT-789",
			},
			mockQrisError: nil,
			wantError:     false,
			validateRequest: func(t *testing.T, result snapCoreQrModel.GenerateQrMpmRequest) {
				assert.Equal(t, "ref-789", result.PartnerReferenceNo)
				assert.Equal(t, "75000.00", result.Amount.Value)
				assert.Equal(t, constant.QrTypeDynamic, result.QrType)
				assert.Equal(t, "ACQ-MERCHANT-789", result.SubMerchantID)
				assert.Equal(t, "ACQ-MERCHANT-789", result.MerchantID)
				assert.Equal(t, "ACQ-TERMINAL-789", result.TerminalID)
				assert.Equal(t, "NOBU", result.Acquirer)
				assert.Equal(t, "", result.StoreID)
			},
		},
		{
			name: "SUCCESS: Dynamic with QR registration - sub-merchant type",
			request: &unifiedPaymentModel.InitProcessorQRISRequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					ClientReferenceID:  "ref-sub",
					ChargeID:           "charge-sub",
					MerchantExternalID: "external-sub",
					Amount: unifiedPaymentModel.Amount{
						Value:    30000.0,
						Currency: "IDR",
					},
					IsStaticPayment: false,
				},
				ExpiryAt: time.Now().Add(20 * time.Minute),
			},
			mockQrisResponse: &qris.Registration{
				Id:                       "qr-reg-sub",
				ExternalId:               "external-sub",
				MerchantType:             constant.QrMerchantTypeSubMerchant,
				Acquirer:                 "DANA",
				AcquirerMerchantId:       &[]string{"SUB-MERCHANT-ID"}[0],
				AcquirerTerminalId:       &[]string{"SUB-TERMINAL-ID"}[0],
				AcquirerParentMerchantId: "PARENT-MERCHANT-ID",
			},
			mockQrisError: nil,
			wantError:     false,
			validateRequest: func(t *testing.T, result snapCoreQrModel.GenerateQrMpmRequest) {
				assert.Equal(t, "ref-sub", result.PartnerReferenceNo)
				assert.Equal(t, "30000.00", result.Amount.Value)
				assert.Equal(t, "PARENT-MERCHANT-ID", result.SubMerchantID)
				assert.Equal(t, "SUB-MERCHANT-ID", result.MerchantID)
				assert.Equal(t, "SUB-MERCHANT-ID", result.StoreID)
				assert.Equal(t, "DANA", result.Acquirer)
			},
		},
		{
			name: "SUCCESS: Validity period exceeds maximum, should be capped",
			request: &unifiedPaymentModel.InitProcessorQRISRequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					ClientReferenceID: "ref-long",
					ChargeID:          "charge-long",
					Amount: unifiedPaymentModel.Amount{
						Value:    10000.0,
						Currency: "IDR",
					},
					IsStaticPayment: false,
					PaymentPartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							AcquirerMerchantID: "LONG-MERCHANT",
							Acquirer:           "TEST",
						},
					},
				},
				ExpiryAt: time.Now().Add(25 * time.Hour), // Way longer than max
			},
			wantError: false,
			validateRequest: func(t *testing.T, result snapCoreQrModel.GenerateQrMpmRequest) {
				assert.Equal(t, constant.QrisDynamicValidityPeriodMax, result.ValidityPeriod)
			},
		},
		{
			name: "ERROR: QR registration not found",
			request: &unifiedPaymentModel.InitProcessorQRISRequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					ClientReferenceID:  "ref-not-found",
					ChargeID:           "charge-not-found",
					MerchantExternalID: "external-not-found",
					Amount: unifiedPaymentModel.Amount{
						Value:    15000.0,
						Currency: "IDR",
					},
				},
				ExpiryAt: time.Now().Add(10 * time.Minute),
			},
			mockQrisResponse: nil,
			mockQrisError:    pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrDataNotFound),
			wantError:        true,
			validateError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), constant.ErrMerchantNotRegisteredQR.Error())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQrisSvc := mockService.NewIQrisService(t)

			service := &UnifiedPaymentService{
				config:  &config.Config{},
				logger:  log,
				qrisSvc: mockQrisSvc,
			}

			// Only mock QRIS service if no partner config is provided
			if tc.request.PaymentPartnerConfig == nil || tc.request.PaymentPartnerConfig.Qris == nil {
				mockQrisSvc.On("FindQrRegistrationByExternalID", mock.Anything, tc.request.MerchantExternalID).
					Return(tc.mockQrisResponse, tc.mockQrisError)
			}

			result, err := service.getQRSnapBasicPayload(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
				if tc.validateError != nil {
					tc.validateError(t, err)
				}
			} else {
				assert.NoError(t, err)
				if tc.validateRequest != nil {
					tc.validateRequest(t, result)
				}
			}

			// Only assert expectations if QRIS service was mocked
			if tc.request.PaymentPartnerConfig == nil || tc.request.PaymentPartnerConfig.Qris == nil {
				mockQrisSvc.AssertExpectations(t)
			}
		})
	}
}

func TestInitQRIS_ErrorPropagation(t *testing.T) {
	baseRequest := &unifiedPaymentModel.InitProcessorQRISRequest{
		BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
			ClientReferenceID: "ref-123",
			ChargeID:          "charge-123",
			Amount: unifiedPaymentModel.Amount{
				Value:    10000,
				Currency: "IDR",
			},
			PaymentPartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
				Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
					AcquirerMerchantID: "merchant-123",
					AcquirerTerminalID: "terminal-123",
					Acquirer:           "NOBU",
				},
			},
		},
		ExpiryAt: time.Now().Add(15 * time.Minute),
	}

	tests := []struct {
		name      string
		repoErr   error
		wantError error
	}{
		{
			name:      "request error should be preserved",
			repoErr:   pkgErr.New(httpResponse.HttpErrRequest, errors.New("invalid request")),
			wantError: pkgErr.New(httpResponse.HttpErrRequest, errors.New("invalid request")),
		},
		{
			name:      "downstream timeout should stay timeout",
			repoErr:   pkgErr.New(httpResponse.HttpErrRequestTimeout, errors.New("timeout")),
			wantError: pkgErr.New(httpResponse.HttpErrRequestTimeout, constant.ErrPartnerInGeneral),
		},
		{
			name:      "downstream bad gateway should stay bad gateway",
			repoErr:   pkgErr.New(httpResponse.HttpErrBadGateway, errors.New("bad gateway")),
			wantError: pkgErr.New(httpResponse.HttpErrBadGateway, constant.ErrPartnerInGeneral),
		},
		{
			name:      "downstream unavailable should stay service unavailable",
			repoErr:   pkgErr.New(httpResponse.HttpErrServiceUnavailable, errors.New("unavailable")),
			wantError: pkgErr.New(httpResponse.HttpErrServiceUnavailable, constant.ErrPartnerInGeneral),
		},
		{
			name:      "downstream request limit exceeded should stay downstream request limit exceeded",
			repoErr:   pkgErr.New(httpResponse.HttpErrRequestLimitExceeded, errors.New("throttled")),
			wantError: pkgErr.New(httpResponse.HttpErrRequestLimitExceeded, constant.ErrPartnerInGeneral),
		},
		{
			name:      "downstream third party should stay third party",
			repoErr:   pkgErr.New(httpResponse.HttpErrThirdParty, errors.New("partner internal")),
			wantError: pkgErr.New(httpResponse.HttpErrThirdParty, constant.ErrPartnerInGeneral),
		},
		{
			name:      "unknown error should become internal general",
			repoErr:   errors.New("unexpected"),
			wantError: pkgErr.New(httpResponse.HttpErrInternal, constant.ErrPartnerInGeneral),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSnapCoreRepo := mockRepo.NewISnapCoreRepository(t)
			mockSnapCoreRepo.On(
				"GenerateQrMpm",
				mock.Anything,
				mock.Anything,
			).Once().Return(nil, tc.repoErr)

			service := &UnifiedPaymentService{
				snapCoreRepo: mockSnapCoreRepo,
			}

			result, err := service.initQRIS(context.Background(), baseRequest)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.Equal(t, tc.wantError.Error(), err.Error())
			mockSnapCoreRepo.AssertExpectations(t)
		})
	}
}

func TestInitVirtualAccount(t *testing.T) {
	type Mockers struct {
		snapCoreRepo *mockRepo.ISnapCoreRepository
	}

	derivedMID := "TEST-MID-123"

	vaFacilStaticWithVaNumber := &unifiedPaymentModel.InitProcessorVARequest{
		BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
			PaymentMethodChannelType: constant.PaymentMethodChannelTypeFacilitator,
			IsStaticPayment:          true,
			Amount: unifiedPaymentModel.Amount{
				Value:    10_000,
				Currency: "IDR",
			},
		},
		Acquirer:      "PERMATA",
		VAAccountName: "Test VA Name",
		VANumber:      "9999999999",
	}
	vaAggreDynamic := &unifiedPaymentModel.InitProcessorVARequest{
		BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
			PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
			Amount: unifiedPaymentModel.Amount{
				Value:    10_000,
				Currency: "IDR",
			},
		},
		Acquirer:      "PERMATA",
		VAAccountName: "Test VA Name",
	}

	testCases := []struct {
		desc                string
		request             *unifiedPaymentModel.InitProcessorVARequest
		mockError           error
		wantError           error
		expectedSubCompany  string
		checkSubCompany     bool
		expectedAccountName string
		checkAccountName    bool
	}{
		{
			desc: "AGGREGATOR channel type should set SubCompany to DerivedMID",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
					DerivedMID:               derivedMID,
					DerivedMerchantID:        "merchant-123",
					DerivedMerchantShortName: "Test Merchant",
					ChargeID:                 "charge-123",
					Amount:                   unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				Acquirer: "PERMATA",
				ExpiryAt: time.Now().Add(15 * time.Minute),
			},
			checkSubCompany:    true,
			expectedSubCompany: derivedMID,
		},
		{
			desc: "FACILITATOR channel type should set SubCompany to empty string",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: constant.PaymentMethodChannelTypeFacilitator,
					DerivedMID:               derivedMID,
					DerivedMerchantID:        "merchant-123",
					DerivedMerchantShortName: "Test Merchant",
					ChargeID:                 "charge-123",
					Amount:                   unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				Acquirer: "PERMATA",
				ExpiryAt: time.Now().Add(15 * time.Minute),
			},
			checkSubCompany:    true,
			expectedSubCompany: "",
		},
		{
			desc: "Other channel type should set SubCompany to empty string",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: "OTHER_TYPE",
					DerivedMID:               derivedMID,
					DerivedMerchantID:        "merchant-123",
					DerivedMerchantShortName: "Test Merchant",
					ChargeID:                 "charge-123",
					Amount:                   unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				Acquirer: "PERMATA",
				ExpiryAt: time.Now().Add(15 * time.Minute),
			},
			checkSubCompany:    true,
			expectedSubCompany: "",
		},
		{
			desc:      "ERROR: Internal server error",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrInternal, assert.AnError),
			wantError: pkgErr.New(httpResponse.HttpErrInternal, constant.ErrPartnerInGeneral),
		},
		{
			desc:      "ERROR: Partner timeout should stay downstream timeout",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrRequestTimeout, errors.New("timeout")),
			wantError: pkgErr.New(httpResponse.HttpErrRequestTimeout, constant.ErrPartnerInGeneral),
		},
		{
			desc:      "ERROR: Partner bad gateway should stay downstream bad gateway",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrBadGateway, errors.New("bad gateway")),
			wantError: pkgErr.New(httpResponse.HttpErrBadGateway, constant.ErrPartnerInGeneral),
		},
		{
			desc:      "ERROR: Partner unavailable should stay downstream service unavailable",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrServiceUnavailable, errors.New("unavailable")),
			wantError: pkgErr.New(httpResponse.HttpErrServiceUnavailable, constant.ErrPartnerInGeneral),
		},
		{
			desc:      "ERROR: Partner rate limit should stay downstream request limit exceeded",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrRequestLimitExceeded, errors.New("throttled")),
			wantError: pkgErr.New(httpResponse.HttpErrRequestLimitExceeded, constant.ErrPartnerInGeneral),
		},
		{
			desc:      "ERROR: Partner generic 500 should stay downstream third party",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrThirdParty, errors.New("partner internal")),
			wantError: pkgErr.New(httpResponse.HttpErrThirdParty, constant.ErrPartnerInGeneral),
		},
		{
			desc:      "ERROR: Static VA number out of range",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrRequest, errors.New("number out of range")),
			wantError: pkgErr.New(httpResponse.HttpErrRequest, fmt.Errorf(constant.ErrDetailMsgVaNumberIsOutsideValidRangeFmt, "static")),
		},
		{
			desc:      "ERROR: Static VA number still in use",
			request:   vaFacilStaticWithVaNumber,
			mockError: pkgErr.New(httpResponse.HttpErrRequest, errors.New("va number still in use")),
			wantError: pkgErr.New(httpResponse.HttpErrRequest, errors.New(constant.ErrDetailMsgVaNumberStillInUse)),
		},
		{
			desc:      "ERROR: Dynamic VA number still in use",
			request:   vaAggreDynamic,
			mockError: pkgErr.New(httpResponse.HttpErrRequest, errors.New("va number still in use")),
			wantError: pkgErr.New(httpResponse.HttpErrRequest, fmt.Errorf(constant.ErrDetailMsgNoAvailableVaNumberToAssignFmt, "dynamic")),
		},
		{
			desc:      "ERROR: Dynamic VA generic request error",
			request:   vaAggreDynamic,
			mockError: pkgErr.New(httpResponse.HttpErrRequest, assert.AnError),
			wantError: pkgErr.New(httpResponse.HttpErrRequest, assert.AnError),
		},
		{
			desc:    "SUCCESS: Dynamic VA creation",
			request: vaAggreDynamic,
		},
		{
			desc: "IsSnap true with VAAccountName should use VAAccountName directly",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
					DerivedMID:               derivedMID, DerivedMerchantID: "merchant-123",
					DerivedMerchantShortName: "Merchant Shop", ChargeID: "charge-123",
					IsSnap: true,
					Amount: unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				VAAccountName: "SNAP VA Name",
				Acquirer:      "PERMATA",
				ExpiryAt:      time.Now().Add(15 * time.Minute),
			},
			checkAccountName:    true,
			expectedAccountName: "SNAP VA Name",
		},
		{
			desc: "IsSnap false with FACILITATOR and VAAccountName should use VAAccountName directly",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: constant.PaymentMethodChannelTypeFacilitator,
					DerivedMID:               derivedMID, DerivedMerchantID: "merchant-123",
					DerivedMerchantShortName: "Merchant Shop", ChargeID: "charge-123",
					Amount: unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				VAAccountName: "Custom VA Name",
				Acquirer:      "PERMATA",
				ExpiryAt:      time.Now().Add(15 * time.Minute),
			},
			checkAccountName:    true,
			expectedAccountName: "Custom VA Name",
		},
		{
			desc: "IsSnap false with AGGREGATOR and VAAccountName should concatenate",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
					DerivedMID:               derivedMID, DerivedMerchantID: "merchant-123",
					DerivedMerchantShortName: "Merchant Shop", ChargeID: "charge-123",
					Amount: unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				VAAccountName: "Custom VA Name",
				Acquirer:      "PERMATA",
				ExpiryAt:      time.Now().Add(15 * time.Minute),
			},
			checkAccountName:    true,
			expectedAccountName: "Merchant Shop - Custom VA Name",
		},
		{
			desc: "IsSnap true without VAAccountName should use merchant short name",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
					DerivedMID:               derivedMID, DerivedMerchantID: "merchant-123",
					DerivedMerchantShortName: "Merchant Shop", ChargeID: "charge-123",
					IsSnap: true,
					Amount: unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				Acquirer: "PERMATA",
				ExpiryAt: time.Now().Add(15 * time.Minute),
			},
			checkAccountName:    true,
			expectedAccountName: "Merchant Shop",
		},
		{
			desc: "IsSnap false with AGGREGATOR without VAAccountName should use merchant short name",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodChannelType: constant.PaymentMethodChannelTypeAggregator,
					DerivedMID:               derivedMID, DerivedMerchantID: "merchant-123",
					DerivedMerchantShortName: "Merchant Shop", ChargeID: "charge-123",
					Amount: unifiedPaymentModel.Amount{Value: 10000.0, Currency: "IDR"},
				},
				Acquirer: "PERMATA",
				ExpiryAt: time.Now().Add(15 * time.Minute),
			},
			checkAccountName:    true,
			expectedAccountName: "Merchant Shop",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				snapCoreRepo: mockRepo.NewISnapCoreRepository(t),
			}

			var capturedRequest snapCoreVaModel.CreateVirtualAccountRequest
			if tc.mockError != nil {
				mockers.snapCoreRepo.On(
					"CreateVirtualAccount", mock.Anything, mock.Anything,
				).Once().Return(nil, tc.mockError)
			} else {
				mockers.snapCoreRepo.On(
					"CreateVirtualAccount", mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					capturedRequest = args.Get(1).(snapCoreVaModel.CreateVirtualAccountRequest)
				}).Return(&snapCoreVaModel.CreateVirtualAccountResponseData{}, nil)
			}

			service := &UnifiedPaymentService{
				config: &config.Config{
					UnifiedPaymentConfig: config.UnifiedPaymentConfig{
						VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
							MinAmount: pdkUtil.ValueToPtr(10_000.00),
							MaxAmount: pdkUtil.ValueToPtr(1_000_000.00),
						},
					},
				},
				snapCoreRepo: mockers.snapCoreRepo,
			}

			_, err := service.initVirtualAccount(context.Background(), tc.request)

			if tc.wantError != nil {
				assert.Equal(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.checkSubCompany {
				assert.Equal(t, tc.expectedSubCompany, capturedRequest.SubCompany)
			}
			if tc.checkAccountName {
				assert.Equal(t, tc.expectedAccountName, capturedRequest.AccountName)
			}

			mockers.snapCoreRepo.AssertExpectations(t)
		})
	}
}

func TestParseBillDetails(t *testing.T) {
	tests := []struct {
		name           string
		request        *unifiedPaymentModel.InitProcessorVARequest
		expectedResult []snapCoreVaModel.BillDetail
	}{
		{
			name:           "nil billDetails should return nil",
			request:        nil,
			expectedResult: nil,
		},
		{
			name: "empty billDetails should return nil",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							BillDetails: &[]snapVa.BillDetail{},
						},
					},
				},
			},
			expectedResult: nil,
		},
		{
			name: "single billDetail with all fields populated",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							BillDetails: &[]snapVa.BillDetail{
								{
									BillerReferenceId: "BILLER-REF-001",
									BillCode:          "BILL-CODE-001",
									BillNo:            "BILL-NO-001",
									BillName:          "Monthly Subscription",
									BillShortName:     "Sub",
									BillDescription: &snapVa.Description{
										English:   "Monthly subscription payment",
										Indonesia: "Pembayaran langganan bulanan",
									},
									BillSubCompany: "SUB-COMPANY-001",
									BillAmount: snapVa.Amount{
										Value:    "150000.00",
										Currency: "IDR",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: []snapCoreVaModel.BillDetail{
				{
					BillerReferenceId: "BILLER-REF-001",
					BillCode:          "BILL-CODE-001",
					BillNo:            "BILL-NO-001",
					BillName:          "Monthly Subscription",
					BillShortName:     "Sub",
					BillDescription: snapCoreVaModel.Description{
						English:   "Monthly subscription payment",
						Indonesia: "Pembayaran langganan bulanan",
					},
					BillSubCompany: "SUB-COMPANY-001",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "150000.00",
						Currency: "IDR",
					},
				},
			},
		},
		{
			name: "multiple billDetails with different data",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							BillDetails: &[]snapVa.BillDetail{
								{
									BillerReferenceId: "BILLER-REF-001",
									BillCode:          "BILL-CODE-001",
									BillNo:            "BILL-NO-001",
									BillName:          "Internet Bill",
									BillShortName:     "Internet",
									BillDescription: &snapVa.Description{
										English:   "Monthly internet bill",
										Indonesia: "Tagihan internet bulanan",
									},
									BillSubCompany: "SUB-001",
									BillAmount: snapVa.Amount{
										Value:    "300000.00",
										Currency: "IDR",
									},
								},
								{
									BillerReferenceId: "BILLER-REF-002",
									BillCode:          "BILL-CODE-002",
									BillNo:            "BILL-NO-002",
									BillName:          "Electricity Bill",
									BillShortName:     "Electricity",
									BillDescription: &snapVa.Description{
										English:   "Monthly electricity bill",
										Indonesia: "Tagihan listrik bulanan",
									},
									BillSubCompany: "SUB-002",
									BillAmount: snapVa.Amount{
										Value:    "500000.00",
										Currency: "IDR",
									},
								},
								{
									BillerReferenceId: "BILLER-REF-003",
									BillCode:          "BILL-CODE-003",
									BillNo:            "BILL-NO-003",
									BillName:          "Water Bill",
									BillShortName:     "Water",
									BillDescription: &snapVa.Description{
										English:   "Monthly water bill",
										Indonesia: "Tagihan air bulanan",
									},
									BillSubCompany: "SUB-003",
									BillAmount: snapVa.Amount{
										Value:    "100000.00",
										Currency: "IDR",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: []snapCoreVaModel.BillDetail{
				{
					BillerReferenceId: "BILLER-REF-001",
					BillCode:          "BILL-CODE-001",
					BillNo:            "BILL-NO-001",
					BillName:          "Internet Bill",
					BillShortName:     "Internet",
					BillDescription: snapCoreVaModel.Description{
						English:   "Monthly internet bill",
						Indonesia: "Tagihan internet bulanan",
					},
					BillSubCompany: "SUB-001",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "300000.00",
						Currency: "IDR",
					},
				},
				{
					BillerReferenceId: "BILLER-REF-002",
					BillCode:          "BILL-CODE-002",
					BillNo:            "BILL-NO-002",
					BillName:          "Electricity Bill",
					BillShortName:     "Electricity",
					BillDescription: snapCoreVaModel.Description{
						English:   "Monthly electricity bill",
						Indonesia: "Tagihan listrik bulanan",
					},
					BillSubCompany: "SUB-002",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "500000.00",
						Currency: "IDR",
					},
				},
				{
					BillerReferenceId: "BILLER-REF-003",
					BillCode:          "BILL-CODE-003",
					BillNo:            "BILL-NO-003",
					BillName:          "Water Bill",
					BillShortName:     "Water",
					BillDescription: snapCoreVaModel.Description{
						English:   "Monthly water bill",
						Indonesia: "Tagihan air bulanan",
					},
					BillSubCompany: "SUB-003",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "100000.00",
						Currency: "IDR",
					},
				},
			},
		},
		{
			name: "billDetail with empty string values",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							BillDetails: &[]snapVa.BillDetail{
								{
									BillerReferenceId: "",
									BillCode:          "",
									BillNo:            "",
									BillName:          "",
									BillShortName:     "",
									BillDescription: &snapVa.Description{
										English:   "",
										Indonesia: "",
									},
									BillSubCompany: "",
									BillAmount: snapVa.Amount{
										Value:    "0.00",
										Currency: "",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: []snapCoreVaModel.BillDetail{
				{
					BillerReferenceId: "",
					BillCode:          "",
					BillNo:            "",
					BillName:          "",
					BillShortName:     "",
					BillDescription: snapCoreVaModel.Description{
						English:   "",
						Indonesia: "",
					},
					BillSubCompany: "",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "0.00",
						Currency: "",
					},
				},
			},
		},
		{
			name: "billDetail with special characters",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							BillDetails: &[]snapVa.BillDetail{
								{
									BillerReferenceId: "BILLER/REF-001",
									BillCode:          "BILL@CODE#001",
									BillNo:            "BILL-NO-001!",
									BillName:          "Bill & Payment",
									BillShortName:     "B&P",
									BillDescription: &snapVa.Description{
										English:   "Payment for <service>",
										Indonesia: "Pembayaran untuk <layanan>",
									},
									BillSubCompany: "SUB-COMPANY-001",
									BillAmount: snapVa.Amount{
										Value:    "250000.50",
										Currency: "IDR",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: []snapCoreVaModel.BillDetail{
				{
					BillerReferenceId: "BILLER/REF-001",
					BillCode:          "BILL@CODE#001",
					BillNo:            "BILL-NO-001!",
					BillName:          "Bill & Payment",
					BillShortName:     "B&P",
					BillDescription: snapCoreVaModel.Description{
						English:   "Payment for <service>",
						Indonesia: "Pembayaran untuk <layanan>",
					},
					BillSubCompany: "SUB-COMPANY-001",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "250000.50",
						Currency: "IDR",
					},
				},
			},
		},
		{
			name: "billDetail with different currencies",
			request: &unifiedPaymentModel.InitProcessorVARequest{
				BaseProcessorRequest: &unifiedPaymentModel.BaseProcessorRequest{
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							BillDetails: &[]snapVa.BillDetail{
								{
									BillerReferenceId: "BILLER-REF-USD",
									BillCode:          "BILL-CODE-USD",
									BillNo:            "BILL-NO-USD",
									BillName:          "USD Payment",
									BillShortName:     "USD",
									BillDescription: &snapVa.Description{
										English:   "Payment in USD",
										Indonesia: "Pembayaran dalam USD",
									},
									BillSubCompany: "SUB-USD",
									BillAmount: snapVa.Amount{
										Value:    "100.00",
										Currency: "USD",
									},
								},
								{
									BillerReferenceId: "BILLER-REF-EUR",
									BillCode:          "BILL-CODE-EUR",
									BillNo:            "BILL-NO-EUR",
									BillName:          "EUR Payment",
									BillShortName:     "EUR",
									BillDescription: &snapVa.Description{
										English:   "Payment in EUR",
										Indonesia: "Pembayaran dalam EUR",
									},
									BillSubCompany: "SUB-EUR",
									BillAmount: snapVa.Amount{
										Value:    "90.00",
										Currency: "EUR",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: []snapCoreVaModel.BillDetail{
				{
					BillerReferenceId: "BILLER-REF-USD",
					BillCode:          "BILL-CODE-USD",
					BillNo:            "BILL-NO-USD",
					BillName:          "USD Payment",
					BillShortName:     "USD",
					BillDescription: snapCoreVaModel.Description{
						English:   "Payment in USD",
						Indonesia: "Pembayaran dalam USD",
					},
					BillSubCompany: "SUB-USD",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "100.00",
						Currency: "USD",
					},
				},
				{
					BillerReferenceId: "BILLER-REF-EUR",
					BillCode:          "BILL-CODE-EUR",
					BillNo:            "BILL-NO-EUR",
					BillName:          "EUR Payment",
					BillShortName:     "EUR",
					BillDescription: snapCoreVaModel.Description{
						English:   "Payment in EUR",
						Indonesia: "Pembayaran dalam EUR",
					},
					BillSubCompany: "SUB-EUR",
					BillAmount: snapCoreVaModel.Amount{
						Value:    "90.00",
						Currency: "EUR",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBillDetails(tt.request)

			if tt.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, len(tt.expectedResult), len(result))

				for i, expected := range tt.expectedResult {
					assert.Equal(t, expected.BillerReferenceId, result[i].BillerReferenceId)
					assert.Equal(t, expected.BillCode, result[i].BillCode)
					assert.Equal(t, expected.BillNo, result[i].BillNo)
					assert.Equal(t, expected.BillName, result[i].BillName)
					assert.Equal(t, expected.BillShortName, result[i].BillShortName)
					assert.Equal(t, expected.BillDescription.English, result[i].BillDescription.English)
					assert.Equal(t, expected.BillDescription.Indonesia, result[i].BillDescription.Indonesia)
					assert.Equal(t, expected.BillSubCompany, result[i].BillSubCompany)
					assert.Equal(t, expected.BillAmount.Value, result[i].BillAmount.Value)
					assert.Equal(t, expected.BillAmount.Currency, result[i].BillAmount.Currency)
				}
			}
		})
	}
}
