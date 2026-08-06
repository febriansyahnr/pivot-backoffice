package unifiedPaymentService

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidatePaymentActivation(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	paymentMethodSvc := serviceMock.NewIPaymentMethodService(t)
	qrisSvc := serviceMock.NewIQrisService(t)

	merchantID := uuid.NewString()
	merchantExternalID := "EXT_" + uuid.NewString()
	acquirer := "BCA"

	testCases := []struct {
		name                string
		paymentMethod       string
		acquirer            string
		isSplitRoute        bool
		wantErr             bool
		expectedErrContains string
		setupMock           func()
	}{
		{
			name:                "ERROR: GetActivePaymentMethodDetailForPaymentRequest returns error",
			paymentMethod:       paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			acquirer:            acquirer,
			isSplitRoute:        false,
			wantErr:             true,
			expectedErrContains: constant.ErrSomeErrorForUnitTest.Error(),
			setupMock: func() {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:                "ERROR: Payment method not found",
			paymentMethod:       paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			acquirer:            acquirer,
			isSplitRoute:        false,
			wantErr:             true,
			expectedErrContains: pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentMethodNotFound).Error(),
			setupMock: func() {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(nil, nil)
			},
		},
		{
			name:                "ERROR: Split route not allowed for facilitator model",
			paymentMethod:       paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			acquirer:            acquirer,
			isSplitRoute:        true,
			wantErr:             true,
			expectedErrContains: pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrDoNotApplySplitRouteInFacilitatorModel).Error(),
			setupMock: func() {
				activePaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Acquirer: acquirer,
					},
					ChannelType: constant.PaymentMethodChannelTypeFacilitator,
					IsActive:    true,
				}
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(activePaymentMethod, nil)
			},
		},
		{
			name:                "ERROR: QRIS merchant not registered",
			paymentMethod:       constant.UnifiedPaymentMethodQris,
			acquirer:            acquirer,
			isSplitRoute:        false,
			wantErr:             true,
			expectedErrContains: "merchant not registered qr",
			setupMock: func() {
				activePaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: acquirer,
					},
					ChannelType: constant.PaymentMethodChannelTypeAggregator,
					IsActive:    true,
				}
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(activePaymentMethod, nil)

				qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer",
					constant.ValueCtxMockType(),
					merchantExternalID,
					acquirer,
				).Once().Return(nil, pkgErr.New(response.HttpErrNotFound, constant.ErrDataNotFound))
			},
		},
		{
			name:                "ERROR: QRIS service error (non data not found)",
			paymentMethod:       constant.UnifiedPaymentMethodQris,
			acquirer:            acquirer,
			isSplitRoute:        false,
			wantErr:             true,
			expectedErrContains: constant.ErrSomeErrorForUnitTest.Error(),
			setupMock: func() {
				activePaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: acquirer,
					},
					ChannelType: constant.PaymentMethodChannelTypeAggregator,
					IsActive:    true,
				}
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(activePaymentMethod, nil)

				qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer",
					constant.ValueCtxMockType(),
					merchantExternalID,
					acquirer,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:          "SUCCESS: Virtual Account payment method",
			paymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			acquirer:      acquirer,
			isSplitRoute:  false,
			wantErr:       false,
			setupMock: func() {
				activePaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Acquirer: acquirer,
					},
					ChannelType: constant.PaymentMethodChannelTypeAggregator,
					IsActive:    true,
				}
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(activePaymentMethod, nil)
			},
		},
		{
			name:          "SUCCESS: E-wallet payment method",
			paymentMethod: paymentConstant.PAYMENT_METHOD_EWALLET,
			acquirer:      acquirer,
			isSplitRoute:  false,
			wantErr:       false,
			setupMock: func() {
				activePaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_EWALLET,
						Acquirer: acquirer,
					},
					ChannelType: constant.PaymentMethodChannelTypeAggregator,
					IsActive:    true,
				}
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(activePaymentMethod, nil)
			},
		},
		{
			name:          "SUCCESS: QRIS payment method",
			paymentMethod: constant.UnifiedPaymentMethodQris,
			acquirer:      acquirer,
			isSplitRoute:  false,
			wantErr:       false,
			setupMock: func() {
				activePaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: acquirer,
					},
					ChannelType: constant.PaymentMethodChannelTypeAggregator,
					IsActive:    true,
				}
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(activePaymentMethod, nil)

				qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer",
					constant.ValueCtxMockType(),
					merchantExternalID,
					acquirer,
				).Once().Return(&qris.Registration{}, nil)
			},
		},
		{
			name:          "SUCCESS: Credit Card payment method (no acquirer check)",
			paymentMethod: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			acquirer:      "",
			isSplitRoute:  false,
			wantErr:       false,
			setupMock: func() {
				activePaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
					ChannelType: constant.PaymentMethodChannelTypeAggregator,
					IsActive:    true,
				}
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("paymentModel.GetPaymentMethodFilterRequest"),
				).Once().Return(activePaymentMethod, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := &UnifiedPaymentService{
				config:                 cfg,
				logger:                 log,
				paymentRepo:            paymentRepo,
				paymentMethodRepo:      paymentMethodRepo,
				accountTransactionRepo: accountTrxRepo,
				paymentMethodSvc:       paymentMethodSvc,
				qrisSvc:                qrisSvc,
			}

			result, err := svc.validatePaymentActivation(context.Background(), merchantID, merchantExternalID, tc.paymentMethod, tc.acquirer, tc.isSplitRoute)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErrContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrContains)
				}
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			if tc.paymentMethod == constant.UnifiedPaymentMethodQris {
				assert.Equal(t, paymentConstant.PAYMENT_METHOD_QRIS, result.Type)
			} else {
				assert.Equal(t, tc.paymentMethod, result.Type)
			}

			if tc.acquirer != "" {
				assert.Equal(t, tc.acquirer, result.Acquirer)
			}
		})
	}
}

func TestValidatePostalCode(t *testing.T) {
	testCases := []struct {
		name       string
		postalCode string
		wantErr    error
	}{
		{
			name:       "empty postal code",
			postalCode: "",
			wantErr:    nil,
		},
		{
			name:       "valid numeric postal code",
			postalCode: "12345",
			wantErr:    nil,
		},
		{
			name:       "valid alphanumeric postal code",
			postalCode: "AB123",
			wantErr:    nil,
		},
		{
			name:       "valid postal code with hyphen",
			postalCode: "12345-6789",
			wantErr:    nil,
		},
		{
			name:       "valid postal code with hyphen and letters",
			postalCode: "A1B-2C3",
			wantErr:    nil,
		},
		{
			name:       "invalid postal code too long",
			postalCode: "12345678901",
			wantErr:    constant.ErrPostalCodeTooLong,
		},
		{
			name:       "invalid postal code with special character",
			postalCode: "ABC@123",
			wantErr:    constant.ErrPostalCodeInvalidFormat,
		},
		{
			name:       "invalid postal code with space",
			postalCode: "123 45",
			wantErr:    constant.ErrPostalCodeInvalidFormat,
		},
		{
			name:       "invalid postal code only hyphens",
			postalCode: "-----",
			wantErr:    constant.ErrPostalCodeInvalidFormat,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePostalCode(tc.postalCode)
			if tc.wantErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestValidateProcessingConfig(t *testing.T) {
	type Mockers struct {
		creditCardProcessorRepo *repositoryMock.ICreditcardCoreProcessorRepository
	}

	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		desc      string
		wantError bool
		setupMock func(mockers Mockers)
	}{
		{
			desc:      "success when processing config is nil",
			wantError: false,
			setupMock: func(mockers Mockers) {},
		},
		{
			desc:      "success when processing config has no values",
			wantError: false,
			setupMock: func(mockers Mockers) {},
		},
		{
			desc:      "success with valid bank merchant id",
			wantError: false,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeDirect,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"BANK123",
				).Return(mid, nil)
			},
		},
		{
			desc:      "success with valid merchant id tag",
			wantError: false,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeDirect,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"BANK456",
				).Return(mid, nil)
			},
		},
		{
			desc:      "error when bank merchant id not found",
			wantError: true,
			setupMock: func(mockers Mockers) {
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"INVALID_MID",
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:      "error when bank merchant id is null",
			wantError: true,
			setupMock: func(mockers Mockers) {
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"NULL_MID",
				).Return(nil, nil)
			},
		},
		{
			desc:      "error when mid type is not direct",
			wantError: true,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeAggregator,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"AGGREGATOR_MID",
				).Return(mid, nil)
			},
		},
		{
			desc:      "error when merchant id tag not found in payment method config",
			wantError: true,
			setupMock: func(mockers Mockers) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				creditCardProcessorRepo: repositoryMock.NewICreditcardCoreProcessorRepository(t),
			}
			tc.setupMock(mockers)

			svc := &UnifiedPaymentService{
				config:                  cfg,
				logger:                  log,
				creditCardProcessorRepo: mockers.creditCardProcessorRepo,
			}

			var processingConfig *unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig
			var paymentMethod *paymentModel.PaymentMethodWithPivot

			switch tc.desc {
			case "success when processing config is nil":
				processingConfig = nil
			case "success when processing config has no values":
				processingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{}
			case "success with valid bank merchant id":
				processingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "BANK123",
				}
			case "success with valid merchant id tag":
				processingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					MerchantIdTag: "tag1",
				}
				paymentMethod = &paymentModel.PaymentMethodWithPivot{
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{
										MerchantIDTag:      "tag1",
										AcquirerMerchantID: "BANK456",
									},
								},
							},
						},
					},
				}
			case "error when bank merchant id not found":
				processingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "INVALID_MID",
				}
			case "error when bank merchant id is null":
				processingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "NULL_MID",
				}
			case "error when mid type is not direct":
				processingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "AGGREGATOR_MID",
				}
			case "error when merchant id tag not found in payment method config":
				processingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					MerchantIdTag: "invalid_tag",
				}
				paymentMethod = &paymentModel.PaymentMethodWithPivot{
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{
										MerchantIDTag:      "different_tag",
										AcquirerMerchantID: "BANK456",
									},
								},
							},
						},
					},
				}
			}

			err := svc.validateProcessingConfig(context.Background(), processingConfig, paymentMethod, constant.CardThreeDsMethodAutomatic)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockers.creditCardProcessorRepo.AssertExpectations(t)
		})
	}
}

func TestValidateBankMerchantId(t *testing.T) {
	type Mockers struct {
		creditCardProcessorRepo *repositoryMock.ICreditcardCoreProcessorRepository
	}

	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		desc      string
		wantError bool
		setupMock func(mockers Mockers)
	}{
		{
			desc:      "success with valid direct mid",
			wantError: false,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeDirect,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"VALID_DIRECT_MID",
				).Return(mid, nil)
			},
		},
		{
			desc:      "error when repository returns error",
			wantError: true,
			setupMock: func(mockers Mockers) {
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"ERROR_MID",
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:      "error when mid not found",
			wantError: true,
			setupMock: func(mockers Mockers) {
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"NOT_FOUND_MID",
				).Return(nil, nil)
			},
		},
		{
			desc:      "error when mid type is aggregator",
			wantError: true,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeAggregator,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"AGGREGATOR_MID",
				).Return(mid, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				creditCardProcessorRepo: repositoryMock.NewICreditcardCoreProcessorRepository(t),
			}
			tc.setupMock(mockers)

			svc := &UnifiedPaymentService{
				config:                  cfg,
				logger:                  log,
				creditCardProcessorRepo: mockers.creditCardProcessorRepo,
			}

			var bankMerchantId string
			switch tc.desc {
			case "success with valid direct mid":
				bankMerchantId = "VALID_DIRECT_MID"
			case "error when repository returns error":
				bankMerchantId = "ERROR_MID"
			case "error when mid not found":
				bankMerchantId = "NOT_FOUND_MID"
			case "error when mid type is aggregator":
				bankMerchantId = "AGGREGATOR_MID"
			}

			err := svc.validateBankMerchantId(context.Background(), bankMerchantId, constant.CardThreeDsMethodAutomatic)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockers.creditCardProcessorRepo.AssertExpectations(t)
		})
	}
}

func TestResolveBankMerchantIdFromTag(t *testing.T) {
	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		desc          string
		merchantIdTag string
		paymentMethod *paymentModel.PaymentMethodWithPivot
		expectedMID   string
		wantError     bool
	}{
		{
			desc:          "success with valid tag",
			merchantIdTag: "tag1",
			paymentMethod: &paymentModel.PaymentMethodWithPivot{
				MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									MerchantIDTag:      "tag1",
									AcquirerMerchantID: "BANK123",
								},
								{
									MerchantIDTag:      "tag2",
									AcquirerMerchantID: "BANK456",
								},
							},
						},
					},
				},
			},
			expectedMID: "BANK123",
			wantError:   false,
		},
		{
			desc:          "error when merchant config is nil",
			merchantIdTag: "tag1",
			paymentMethod: &paymentModel.PaymentMethodWithPivot{
				MerchantConfigObj: nil,
			},
			wantError: true,
		},
		{
			desc:          "error when partner config is nil",
			merchantIdTag: "tag1",
			paymentMethod: &paymentModel.PaymentMethodWithPivot{
				MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
					PartnerConfig: nil,
				},
			},
			wantError: true,
		},
		{
			desc:          "error when card config is nil",
			merchantIdTag: "tag1",
			paymentMethod: &paymentModel.PaymentMethodWithPivot{
				MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: nil,
					},
				},
			},
			wantError: true,
		},
		{
			desc:          "error when tag not found",
			merchantIdTag: "nonexistent_tag",
			paymentMethod: &paymentModel.PaymentMethodWithPivot{
				MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									MerchantIDTag:      "tag1",
									AcquirerMerchantID: "BANK123",
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			svc := &UnifiedPaymentService{
				config: cfg,
				logger: log,
			}

			result, err := svc.resolveBankMerchantIdFromTag(tc.merchantIdTag, tc.paymentMethod)

			if tc.wantError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMID, result)
			}
		})
	}
}

func TestValidatePaymentExpiry(t *testing.T) {
	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := uuid.NewString()
	now := time.Now()
	futureTime20Minutes := now.Add(20 * time.Minute)
	futureTime25Hours := now.Add(25 * time.Hour)
	futureTime31Minutes := now.Add(31 * time.Minute)
	futureTime10Hours := now.Add(10 * time.Hour)
	futureTime2Days := now.Add(48 * time.Hour)
	futureTime8Days := now.Add(8 * 24 * time.Hour)
	futureTime40Days := now.Add(40 * 24 * time.Hour)

	testCases := []struct {
		name                string
		setupService        func() *UnifiedPaymentService
		cmd                 paymentModel.PaymentRequestExpiryValidation
		wantErr             bool
		expectedErrContains string
	}{
		{
			name: "SUCCESS: Zero expiry time - should skip validation",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: cfg,
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: time.Time{},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Card method without ExpiryAt in PaymentMethodOptions - should skip validation",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: cfg,
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: time.Time{},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: VA with expiry within default limit (24 hours)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime10Hours,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: VA with expiry exceeding default limit (24 hours)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime25Hours,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr:             true,
			expectedErrContains: "max expiry time is 24 HOURS",
		},
		{
			name: "SUCCESS: QRIS with expiry within default limit (30 minutes)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							QrConfig: &config.UnifiedPaymentQrConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_QRIS,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime20Minutes,
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						QR: &unifiedPaymentModel.PaymentMethodOptionQR{
							ExpiryAt: &futureTime20Minutes,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: QRIS with expiry exceeding default limit (30 minutes)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							QrConfig: &config.UnifiedPaymentQrConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_QRIS,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime31Minutes,
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						QR: &unifiedPaymentModel.PaymentMethodOptionQR{
							ExpiryAt: &futureTime31Minutes,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr:             true,
			expectedErrContains: "max expiry time is 30 MINUTES",
		},
		{
			name: "SUCCESS: VA with payment method options expiry at",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: time.Time{},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							ExpiryAt: &futureTime10Hours,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: VA with payment method options expiry exceeding limit",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: time.Time{},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							ExpiryAt: &futureTime2Days,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr:             true,
			expectedErrContains: "max expiry time is 24 HOURS",
		},
		{
			name: "SUCCESS: VA with custom payment method config override",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime2Days,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 7,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: VA with custom payment method config override - exceeds limit",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime8Days,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 7,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
						},
					},
				},
			},
			wantErr:             true,
			expectedErrContains: "max expiry time is 7 DAYS",
		},
		{
			name: "SUCCESS: Merchant should skip validation based on ExpiryConfig (PARTIAL mode with excluded merchant)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode:              paymentConstant.UnifiedPaymentExpiryModePartial,
								ExcludedMerchants: []string{merchantID},
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime25Hours,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Unified payment method QRIS constant variant",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							QrConfig: &config.UnifiedPaymentQrConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     constant.UnifiedPaymentMethodQris,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime20Minutes,
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						QR: &unifiedPaymentModel.PaymentMethodOptionQR{
							ExpiryAt: &futureTime20Minutes,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: No VA config set - validation fails with zero duration",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime25Hours,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr:             true,
			expectedErrContains: "max expiry time is 0",
		},
		{
			name: "SUCCESS: Payment method config has zero duration - uses default config instead",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     48,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime25Hours,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 0,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Payment method config has empty unit - should use default config",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					ExpiryAt: futureTime10Hours,
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 10,
								Unit:     "",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Card with expiry within default limit (30 days)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							CardConfig: &config.UnifiedPaymentCardConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ExpiryAt: &futureTime2Days,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Card with expiry exceeding default limit (30 days)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							CardConfig: &config.UnifiedPaymentCardConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ExpiryAt: &futureTime40Days,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr:             true,
			expectedErrContains: "max expiry time is 30 DAYS",
		},
		{
			name: "SUCCESS: Unified card method with expiry within limit",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							CardConfig: &config.UnifiedPaymentCardConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     constant.UnifiedPaymentMethodCard,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ExpiryAt: &futureTime10Hours,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Ewallet with expiry within default limit (30 minutes)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							EwalletConfig: &config.UnifiedPaymentEwalletConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_EWALLET,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
							ExpiryAt: &futureTime20Minutes,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Ewallet with expiry exceeding default limit (30 minutes)",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							EwalletConfig: &config.UnifiedPaymentEwalletConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_EWALLET,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
							ExpiryAt: &futureTime31Minutes,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			wantErr:             true,
			expectedErrContains: "max expiry time is 30 MINUTES",
		},
		{
			name: "SUCCESS: Card with custom payment method config override",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							CardConfig: &config.UnifiedPaymentCardConfig{
								MaxExpiryDuration:     7,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Card: &unifiedPaymentModel.PaymentMethodOptionCard{
							ExpiryAt: &futureTime8Days,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 10,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Ewallet with custom payment method config override",
			setupService: func() *UnifiedPaymentService {
				return &UnifiedPaymentService{
					config: &config.Config{
						Environment: "test",
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							EwalletConfig: &config.UnifiedPaymentEwalletConfig{
								MaxExpiryDuration:     15,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
					logger: log,
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_EWALLET,
				MerchantID: merchantID,
				UnifiedPaymentRequest: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
							ExpiryAt: &futureTime31Minutes,
						},
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 45,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := tc.setupService()

			err := svc.ValidatePaymentExpiry(context.Background(), tc.cmd)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErrContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrContains)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestValidateCardSupportedUseCase(t *testing.T) {
	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		name                  string
		threeDsMethod         string
		cardSupportedUseCases []*paymentMethodModel.CardSupportedUseCase
		isRecurringPayment    bool
		isAutoSplitPayment    bool
		cardOnFile            *unifiedPaymentModel.PaymentMethodOptionCardOnFileConfig
		wantErr               bool
		expectedErrMsg        string
	}{
		{
			name:                  "empty threeDsMethod should pass",
			threeDsMethod:         "",
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{},
			wantErr:               false,
		},
		{
			name:          "NEVER with AllowBypass3Ds=true should pass",
			threeDsMethod: constant.CardThreeDsMethodNever,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       true,
					AllowExternalThreeDs: false,
				},
			},
			wantErr: false,
		},
		{
			name:          "NEVER with AllowBypass3Ds=false should fail",
			threeDsMethod: constant.CardThreeDsMethodNever,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: false,
				},
			},
			wantErr:        true,
			expectedErrMsg: "payment does not allow bypassing 3DS",
		},
		{
			name:          "EXTERNAL with AllowExternalThreeDs=true should pass",
			threeDsMethod: constant.CardThreeDsMethodExternal,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: true,
				},
			},
			wantErr: false,
		},
		{
			name:          "EXTERNAL with AllowExternalThreeDs=false should fail",
			threeDsMethod: constant.CardThreeDsMethodExternal,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: false,
				},
			},
			wantErr:        true,
			expectedErrMsg: "external 3DS is not enabled for this merchant",
		},
		{
			name:          "EXTERNAL with multiple configs, one allows should pass",
			threeDsMethod: constant.CardThreeDsMethodExternal,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: false,
				},
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: true,
				},
			},
			wantErr: false,
		},
		{
			name:          "EXTERNAL with multiple configs, none allows should fail",
			threeDsMethod: constant.CardThreeDsMethodExternal,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: false,
				},
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: false,
				},
			},
			wantErr:        true,
			expectedErrMsg: "external 3DS is not enabled for this merchant",
		},
		{
			name:          "AUTOMATIC should pass regardless of flags",
			threeDsMethod: constant.CardThreeDsMethodAutomatic,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: false,
				},
			},
			wantErr: false,
		},
		{
			name:          "CHALLENGE should pass regardless of flags",
			threeDsMethod: constant.CardThreeDsMethodChallenge,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       false,
					AllowExternalThreeDs: false,
				},
			},
			wantErr: false,
		},
		{
			name:               "NEVER for subsequent recurring payments",
			threeDsMethod:      constant.CardThreeDsMethodNever,
			isRecurringPayment: true,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds: false,
				},
			},
			wantErr: false,
		},
		{
			name:               "CHALLENGE for parent split payment",
			threeDsMethod:      constant.CardThreeDsMethodChallenge,
			isAutoSplitPayment: true,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds: false,
				},
			},
			wantErr: false,
		},
		{
			name:          "ERROR: MIT with non-NEVER ThreeDsMethod should fail",
			threeDsMethod: constant.CardThreeDsMethodAutomatic,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds: true,
				},
			},
			cardOnFile: &unifiedPaymentModel.PaymentMethodOptionCardOnFileConfig{
				Initiator:                    constant.COFInitiatorMerchant,
				PreviousNetworkTransactionID: "some-txn-id",
			},
			wantErr:        true,
			expectedErrMsg: constant.ErrInvalidMITThreeDSMethod.Error(),
		},
		{
			name:          "ERROR: MIT with CHALLENGE ThreeDsMethod should fail",
			threeDsMethod: constant.CardThreeDsMethodChallenge,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds: true,
				},
			},
			cardOnFile: &unifiedPaymentModel.PaymentMethodOptionCardOnFileConfig{
				Initiator:                    constant.COFInitiatorMerchant,
				PreviousNetworkTransactionID: "some-txn-id",
			},
			wantErr:        true,
			expectedErrMsg: constant.ErrInvalidMITThreeDSMethod.Error(),
		},
		{
			name:          "ERROR: MIT with empty PreviousNetworkTransactionID should fail",
			threeDsMethod: constant.CardThreeDsMethodNever,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds: true,
				},
			},
			cardOnFile: &unifiedPaymentModel.PaymentMethodOptionCardOnFileConfig{
				Initiator:                    constant.COFInitiatorMerchant,
				PreviousNetworkTransactionID: "",
			},
			wantErr:        true,
			expectedErrMsg: constant.ErrMissingMerchantPrevioustNetworkTransactionID.Error(),
		},
		{
			name:          "ERROR: CIT with PreviousNetworkTransactionID should fail",
			threeDsMethod: constant.CardThreeDsMethodNever,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds: true,
				},
			},
			cardOnFile: &unifiedPaymentModel.PaymentMethodOptionCardOnFileConfig{
				Initiator:                    constant.COFInitiatorCustomer,
				PreviousNetworkTransactionID: "prev-network",
			},
			wantErr:        true,
			expectedErrMsg: constant.ErrMerchantPreviousNetworkTransactionIDNotAllowedForCIT.Error(),
		},
		{
			name:          "SUCCESS: MIT with NEVER ThreeDsMethod and valid PreviousNetworkTransactionID should pass",
			threeDsMethod: constant.CardThreeDsMethodNever,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds: true,
				},
			},
			cardOnFile: &unifiedPaymentModel.PaymentMethodOptionCardOnFileConfig{
				Initiator:                    constant.COFInitiatorMerchant,
				PreviousNetworkTransactionID: "valid-network-txn-id",
			},
			wantErr: false,
		},
		{
			name:          "SUCCESS: CardOnFile with CUSTOMER initiator should pass regardless of ThreeDsMethod",
			threeDsMethod: constant.CardThreeDsMethodExternal,
			cardSupportedUseCases: []*paymentMethodModel.CardSupportedUseCase{
				{
					AllowBypass3Ds:       true,
					AllowExternalThreeDs: true,
				},
			},
			cardOnFile: &unifiedPaymentModel.PaymentMethodOptionCardOnFileConfig{
				Initiator:                    constant.COFInitiatorCustomer,
				PreviousNetworkTransactionID: "",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &UnifiedPaymentService{
				config: cfg,
				logger: log,
			}

			request := &unifiedPaymentModel.ValidateCardRequest{
				Mode:                  constant.UnifiedPaymentModeAPI,
				IsConfirmStep:         false,
				CardSupportedUseCases: tc.cardSupportedUseCases,
				IsRecurringPayment:    tc.isRecurringPayment,
				IsAutoSplitPayment:    tc.isAutoSplitPayment,
			}

			// Only set CardPaymentMethodOptions if threeDsMethod is not empty or cardOnFile is set
			if tc.threeDsMethod != "" || tc.cardOnFile != nil {
				request.CardPaymentMethodOptions = &unifiedPaymentModel.PaymentMethodOptionCard{
					ThreeDsMethod: tc.threeDsMethod,
					CardOnFile:    tc.cardOnFile,
				}
			}

			err := svc.validateCard(
				context.Background(),
				request,
				&paymentModel.PaymentMethodWithPivot{},
			)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateThreeDsInfo(t *testing.T) {
	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	validUUID := uuid.NewString()

	testCases := []struct {
		name                string
		threeDsInfo         *unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo
		wantErr             bool
		expectedErrContains string
	}{
		{
			name:                "ERROR: threeDsInfo is nil",
			threeDsInfo:         nil,
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentRequireThreeDsInfoForThreeDsMethodExternal.Error(),
		},
		{
			name: "ERROR: ThreeDSVersion is 3DS v1 (starts with 1.)",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "1.0.2",
				ECI:                  "05",
				TransactionStatus:    "Y",
				AuthenticationScheme: "VISA",
			},
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalidThreeDsInfoFormat.Error(),
		},
		{
			name: "ERROR: TransactionStatus is not Y (authentication failed)",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.1.0",
				ECI:                  "05",
				TransactionStatus:    "N",
				AuthenticationScheme: "VISA",
			},
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPayment3DsAuthenticationNotSuccessful.Error(),
		},
		{
			name: "ERROR: ECI does not match scheme (VISA should be 05, got 01)",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.1.0",
				ECI:                  "01",
				TransactionStatus:    "Y",
				AuthenticationScheme: "VISA",
			},
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name: "SUCCESS: Valid VISA 3DS info with ECI 05",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.1.0",
				ECI:                  "05",
				TransactionStatus:    "Y",
				AuthenticationScheme: "VISA",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Valid MASTERCARD 3DS info with ECI 02",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.2.0",
				ECI:                  "02",
				TransactionStatus:    "Y",
				AuthenticationScheme: "MASTERCARD",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Valid JCB 3DS info with ECI 05",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.1.0",
				ECI:                  "05",
				TransactionStatus:    "Y",
				AuthenticationScheme: "JCB",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Valid AMEX 3DS info with ECI 05",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.1.0",
				ECI:                  "05",
				TransactionStatus:    "Y",
				AuthenticationScheme: "AMEX",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Valid UNIONPAY 3DS info",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.1.0",
				ECI:                  "05",
				TransactionStatus:    "Y",
				AuthenticationScheme: "UNIONPAY",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Valid with ACSTransactionID and ACSReference",
			threeDsInfo: &unifiedPaymentModel.PaymentMethodOptionCardThreeDsInfo{
				TransactionID:        validUUID,
				ThreeDSVersion:       "2.1.0",
				ECI:                  "05",
				TransactionStatus:    "Y",
				AuthenticationScheme: "VISA",
				ACSTransactionID:     uuid.NewString(),
				ACSReference:         "ACSREF123456",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &UnifiedPaymentService{
				config: cfg,
				logger: log,
			}

			err := svc.validateThreeDsInfo(context.Background(), tc.threeDsInfo)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErrContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateECIMatchesScheme(t *testing.T) {
	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		name                string
		eci                 string
		scheme              string
		wantErr             bool
		expectedErrContains string
	}{
		{
			name:                "ERROR: Invalid scheme",
			eci:                 "05",
			scheme:              "UNKNOWN_SCHEME",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: VISA with invalid ECI 01",
			eci:                 "01",
			scheme:              "VISA",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: VISA with invalid ECI 02",
			eci:                 "02",
			scheme:              "VISA",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:    "SUCCESS: VISA with valid ECI 05",
			eci:     "05",
			scheme:  "VISA",
			wantErr: false,
		},
		{
			name:                "ERROR: VISA with invalid ECI 06",
			eci:                 "06",
			scheme:              "VISA",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: VISA with invalid ECI 07",
			eci:                 "07",
			scheme:              "VISA",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: MASTERCARD with invalid ECI 01",
			eci:                 "01",
			scheme:              "MASTERCARD",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:    "SUCCESS: MASTERCARD with valid ECI 02",
			eci:     "02",
			scheme:  "MASTERCARD",
			wantErr: false,
		},
		{
			name:                "ERROR: MASTERCARD with invalid ECI 06",
			eci:                 "06",
			scheme:              "MASTERCARD",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: MASTERCARD with invalid ECI 07",
			eci:                 "07",
			scheme:              "MASTERCARD",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: MASTERCARD with invalid ECI 05",
			eci:                 "05",
			scheme:              "MASTERCARD",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:    "SUCCESS: JCB with valid ECI 05",
			eci:     "05",
			scheme:  "JCB",
			wantErr: false,
		},
		{
			name:                "ERROR: JCB with invalid ECI 06",
			eci:                 "06",
			scheme:              "JCB",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: JCB with invalid ECI 07",
			eci:                 "07",
			scheme:              "JCB",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: JCB with invalid ECI 01",
			eci:                 "01",
			scheme:              "JCB",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:    "SUCCESS: AMEX with valid ECI 05",
			eci:     "05",
			scheme:  "AMEX",
			wantErr: false,
		},
		{
			name:                "ERROR: AMEX with invalid ECI 06",
			eci:                 "06",
			scheme:              "AMEX",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: AMEX with invalid ECI 07",
			eci:                 "07",
			scheme:              "AMEX",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:    "SUCCESS: UNIONPAY with valid ECI 05",
			eci:     "05",
			scheme:  "UNIONPAY",
			wantErr: false,
		},
		{
			name:    "SUCCESS: UNIONPAY with valid ECI 02",
			eci:     "02",
			scheme:  "UNIONPAY",
			wantErr: false,
		},
		{
			name:                "ERROR: UNIONPAY with invalid ECI 06",
			eci:                 "06",
			scheme:              "UNIONPAY",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
		{
			name:                "ERROR: UNIONPAY with invalid ECI 07",
			eci:                 "07",
			scheme:              "UNIONPAY",
			wantErr:             true,
			expectedErrContains: constant.ErrUnifiedPaymentInvalid3DsAuthenticationResult.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &UnifiedPaymentService{
				config: cfg,
				logger: log,
			}

			err := svc.validateECIMatchesScheme(tc.eci, tc.scheme)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErrContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBankMerchantId_WithExternalThreeDsMethod(t *testing.T) {
	type Mockers struct {
		creditCardProcessorRepo *repositoryMock.ICreditcardCoreProcessorRepository
	}

	cfg := &config.Config{Environment: "test"}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		desc          string
		threeDsMethod string
		wantError     bool
		setupMock     func(mockers Mockers)
	}{
		{
			desc:          "success with EXTERNAL threeDsMethod and aggregator MID type",
			threeDsMethod: constant.CardThreeDsMethodExternal,
			wantError:     false,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeAggregator,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"EXTERNAL_AGGREGATOR_MID",
				).Return(mid, nil)
			},
		},
		{
			desc:          "success with EXTERNAL threeDsMethod and direct MID type",
			threeDsMethod: constant.CardThreeDsMethodExternal,
			wantError:     false,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeDirect,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"EXTERNAL_DIRECT_MID",
				).Return(mid, nil)
			},
		},
		{
			desc:          "error with non-EXTERNAL threeDsMethod and aggregator MID type",
			threeDsMethod: constant.CardThreeDsMethodAutomatic,
			wantError:     true,
			setupMock: func(mockers Mockers) {
				mid := &creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.CreditCardMidTypeAggregator,
				}
				mockers.creditCardProcessorRepo.On(
					"GetMIDByAcquirerMID",
					mock.Anything,
					"NON_EXTERNAL_AGGREGATOR_MID",
				).Return(mid, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				creditCardProcessorRepo: repositoryMock.NewICreditcardCoreProcessorRepository(t),
			}
			tc.setupMock(mockers)

			svc := &UnifiedPaymentService{
				config:                  cfg,
				logger:                  log,
				creditCardProcessorRepo: mockers.creditCardProcessorRepo,
			}

			var bankMerchantId string
			switch tc.desc {
			case "success with EXTERNAL threeDsMethod and aggregator MID type":
				bankMerchantId = "EXTERNAL_AGGREGATOR_MID"
			case "success with EXTERNAL threeDsMethod and direct MID type":
				bankMerchantId = "EXTERNAL_DIRECT_MID"
			case "error with non-EXTERNAL threeDsMethod and aggregator MID type":
				bankMerchantId = "NON_EXTERNAL_AGGREGATOR_MID"
			}

			err := svc.validateBankMerchantId(context.Background(), bankMerchantId, tc.threeDsMethod)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockers.creditCardProcessorRepo.AssertExpectations(t)
		})
	}
}
