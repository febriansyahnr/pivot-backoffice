package paymentMethodService

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/go/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetPaymentMethodByMerchant(t *testing.T) {
	repo := repositoryMocks.NewIPaymentMethodRepository(t)
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
	creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	redis := redisMocks.NewIRedisExt(t)

	validMerchantID := uuid.NewString()
	payloadWithPayment := &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: validMerchantID,
		Payment: &paymentModel.PaymentDetailForPaymentUIResponse{
			Amount: commonModel.Amount{
				Currency: constant.CurrencyIDR,
				Value:    "12000",
			},
		},
	}

	payload := &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: validMerchantID,
	}

	tests := []struct {
		name         string
		payload      *paymentModel.GetPaymentMethodFilterRequest
		modifierMock func()
		wantErr      bool
	}{
		{
			name:    "ERROR: GetListPaymentMethodMerchant error",
			payload: payloadWithPayment,
			modifierMock: func() {
				repo.On(
					"GetListPaymentMethodMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name:    "Success: GetListPaymentMethodMerchant with payment, it should validate the payment",
			payload: payloadWithPayment,
			modifierMock: func() {
				repo.On(
					"GetListPaymentMethodMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Return([]*paymentModel.PaymentMethodWithPivot{{
					MerchantID: validMerchantID,
				}, nil}, nil)
			},
			wantErr: false,
		},
		{
			name:    "Success: GetListPaymentMethodMerchant ",
			payload: payload,
			modifierMock: func() {
				repo.On(
					"GetListPaymentMethodMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Return([]*paymentModel.PaymentMethodWithPivot{{
					MerchantID: validMerchantID,
				}, nil}, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.modifierMock()

			svc := New(logger, repo, snapCoreRepo, creditCardRepo, WithRedisClient(redis), WithConfig(&config.Config{
				UnifiedPaymentConfig: config.UnifiedPaymentConfig{
					VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
						MinAmount: util.ValueToPtr(1000.00),
						MaxAmount: util.ValueToPtr(1000.00),
					},
					QrConfig: &config.UnifiedPaymentQrConfig{
						MinAmount: util.ValueToPtr(1000.00),
						MaxAmount: util.ValueToPtr(1000.00),
					},
					CardConfig: &config.UnifiedPaymentCardConfig{
						MinAmount: util.ValueToPtr(1000.00),
						MaxAmount: util.ValueToPtr(1000.00),
					},
					EwalletConfig: &config.UnifiedPaymentEwalletConfig{
						MinAmount: util.ValueToPtr(1000.00),
						MaxAmount: util.ValueToPtr(1000.00),
					},
				},
			}))
			got, err := svc.GetPaymentMethodByMerchant(context.Background(), tt.payload)

			if tt.wantErr {
				assert.Empty(t, got)
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetStaticVAPaymentMethodByMerchant(t *testing.T) {
	ctx := context.Background()
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		filter         *paymentModel.GetPaymentMethodFilterRequest
		setupMocks     func(*repositoryMocks.IPaymentMethodRepository, *repositoryMocks.ISnapCoreRepository, *redisMocks.IRedisExt)
		expectedResult []*paymentModel.PaymentMethodWithPivot
		shouldErr      bool
	}{
		{
			name: "SUCCESS: Get payment methods without cache, with default VA config",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository, redisClient *redisMocks.IRedisExt) {
				// Mock payment method repository
				paymentMethods := []*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID:     "pm-123",
							Name:     "Virtual Account",
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							Acquirer: "BNI",
						},
						MerchantConfigObj: nil, // No existing config
					},
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID:     "pm-456",
							Name:     "Virtual Account BCA",
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							Acquirer: "BCA",
						},
						MerchantConfigObj: nil, // No existing config
					},
				}

				expectedFilter := &paymentModel.GetPaymentMethodFilterRequest{
					MerchantID: "merchant-123",
					Status:     "ACTIVE",
					Category:   constant.TypePayment,
					Type:       paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				}

				paymentMethodRepo.On("GetListPaymentMethodMerchant", mock.Anything,
					mock.MatchedBy(func(filter *paymentModel.GetPaymentMethodFilterRequest) bool {
						return filter.MerchantID == expectedFilter.MerchantID &&
							filter.Status == expectedFilter.Status &&
							filter.Category == expectedFilter.Category &&
							filter.Type == expectedFilter.Type
					})).Return(paymentMethods, nil)

				// Mock Redis cache miss
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetErr(redis.Nil)
				redisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(getCmd).Once()

				// Mock snap core repository
				defaultVAConfigs := []*snapCoreVAModel.VirtualAccountConfigResponseData{
					{
						Acquirer:  "BNI",
						BinPrefix: "88812",
						Type:      "CLOSED_STATIC",
						BinMin:    3,
						MetadataObj: snapCoreVAModel.VirtualAccountConfigMetadata{
							MerchantPrefix: struct {
								StartRange string `json:"start_range"`
								EndRange   string `json:"end_range"`
							}{
								StartRange: "001",
								EndRange:   "999",
							},
						},
					},
					{
						Acquirer:  "BCA",
						BinPrefix: "70012",
						Type:      "OPEN_STATIC",
						BinMin:    5,
						MetadataObj: snapCoreVAModel.VirtualAccountConfigMetadata{
							MerchantPrefix: struct {
								StartRange string `json:"start_range"`
								EndRange   string `json:"end_range"`
							}{
								StartRange: "",
								EndRange:   "",
							},
						},
					},
				}

				snapCoreRepo.On("GetVirtualAccountConfig", mock.Anything,
					mock.MatchedBy(func(req *snapCoreVAModel.GetVirtualAccountConfigRequest) bool {
						return req.MerchantID == "default-merchant-id" &&
							req.IntegrationType == constant.PaymentMethodChannelTypeAggregator
					})).Return(defaultVAConfigs, nil)

				// Mock Redis cache set
				setCmd := redis.NewStatusCmd(context.Background())
				redisClient.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 24*time.Hour).Return(setCmd).Once()
			},
			expectedResult: []*paymentModel.PaymentMethodWithPivot{
				{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     "pm-123",
						Name:     "Virtual Account",
						Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Acquirer: "BNI",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
									{
										BINPrefix:  "88812",
										Type:       "CLOSED_STATIC",
										StartRange: "001",
										EndRange:   "999",
									},
								},
							},
						},
					},
				},
				{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     "pm-456",
						Name:     "Virtual Account BCA",
						Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Acquirer: "BCA",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
									{
										BINPrefix:  "70012",
										Type:       "OPEN_STATIC",
										StartRange: "00001",
										EndRange:   "99999",
									},
								},
							},
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "SUCCESS: Get payment methods with existing cache",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository, redisClient *redisMocks.IRedisExt) {
				// Mock payment method repository
				paymentMethods := []*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID:     "pm-123",
							Name:     "Virtual Account",
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							Acquirer: "BNI",
						},
						MerchantConfigObj: nil,
					},
				}

				paymentMethodRepo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return(paymentMethods, nil)

				// Mock Redis cache hit
				cachedVAConfigs := []*snapCoreVAModel.VirtualAccountConfigResponseData{
					{
						Acquirer:  "BNI",
						BinPrefix: "88812",
						Type:      "CLOSED_STATIC",
						BinMin:    3,
						MetadataObj: snapCoreVAModel.VirtualAccountConfigMetadata{
							MerchantPrefix: struct {
								StartRange string `json:"start_range"`
								EndRange   string `json:"end_range"`
							}{
								StartRange: "001",
								EndRange:   "99999",
							},
						},
					},
				}

				cachedData, _ := json.Marshal(cachedVAConfigs)
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetVal(string(cachedData))
				redisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(getCmd).Once()

				// Snap core should not be called when cache exists
			},
			expectedResult: []*paymentModel.PaymentMethodWithPivot{
				{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     "pm-123",
						Name:     "Virtual Account",
						Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Acquirer: "BNI",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
									{
										BINPrefix:  "88812",
										Type:       "CLOSED_STATIC",
										StartRange: "001",
										EndRange:   "99999",
									},
								},
							},
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "SUCCESS: Payment method with existing config should not be overridden",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository, redisClient *redisMocks.IRedisExt) {
				// Mock payment method repository with existing config
				existingConfig := &paymentModel.PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
								{
									BINPrefix:  "12345",
									Type:       "CUSTOM_TYPE",
									StartRange: "100",
									EndRange:   "200",
								},
							},
						},
					},
				}

				paymentMethods := []*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID:     "pm-123",
							Name:     "Virtual Account",
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							Acquirer: "BNI",
						},
						MerchantConfigObj: existingConfig,
					},
				}

				paymentMethodRepo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return(paymentMethods, nil)

				// Cache and snap core won't be called since config exists, but service still tries
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetErr(redis.Nil)
				redisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(getCmd).Maybe()
				snapCoreRepo.On("GetVirtualAccountConfig", mock.Anything, mock.Anything).Return([]*snapCoreVAModel.VirtualAccountConfigResponseData{}, nil).Maybe()
				setCmd := redis.NewStatusCmd(context.Background())
				redisClient.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 24*time.Hour).Return(setCmd).Maybe()
			},
			expectedResult: []*paymentModel.PaymentMethodWithPivot{
				{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     "pm-123",
						Name:     "Virtual Account",
						Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Acquirer: "BNI",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
									{
										BINPrefix:  "12345",
										Type:       "CUSTOM_TYPE",
										StartRange: "100",
										EndRange:   "200",
									},
								},
							},
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "ERROR: Payment method repository error",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository, redisClient *redisMocks.IRedisExt) {
				paymentMethodRepo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedResult: nil,
			shouldErr:      true,
		},
		{
			name: "ERROR: Snap core repository error",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository, redisClient *redisMocks.IRedisExt) {
				// Mock payment method repository
				paymentMethods := []*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID:     "pm-123",
							Name:     "Virtual Account",
							Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
							Acquirer: "BNI",
						},
						MerchantConfigObj: nil,
					},
				}

				paymentMethodRepo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return(paymentMethods, nil)

				// Mock Redis cache miss
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetErr(redis.Nil)
				redisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(getCmd).Once()

				// Mock snap core repository error
				snapCoreRepo.On("GetVirtualAccountConfig", mock.Anything, mock.Anything).Return(nil, errors.New("snap core error"))
			},
			expectedResult: nil,
			shouldErr:      true,
		},
		{
			name: "SUCCESS: Empty payment methods list",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository, redisClient *redisMocks.IRedisExt) {
				paymentMethodRepo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{}, nil)
				// Service still tries to get VA config even for empty list
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetErr(redis.Nil)
				redisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(getCmd).Once()

				// Mock snap core repository
				snapCoreRepo.On("GetVirtualAccountConfig", mock.Anything, mock.Anything).Return([]*snapCoreVAModel.VirtualAccountConfigResponseData{}, nil)

				// Mock Redis cache set
				setCmd := redis.NewStatusCmd(context.Background())
				redisClient.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 24*time.Hour).Return(setCmd).Once()
			},
			expectedResult: []*paymentModel.PaymentMethodWithPivot{},
			shouldErr:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			paymentMethodRepo := &repositoryMocks.IPaymentMethodRepository{}
			snapCoreRepo := &repositoryMocks.ISnapCoreRepository{}
			redisClient := &redisMocks.IRedisExt{}

			tc.setupMocks(paymentMethodRepo, snapCoreRepo, redisClient)

			// Create service with test config
			cfg := &config.Config{
				UnifiedPaymentConfig: config.UnifiedPaymentConfig{
					VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
						DefaultVaConfigMerchantId: "default-merchant-id",
						DefaultVaRangeStart:       "1",
						DefaultVaRangeEnd:         "99999",
					},
				},
			}

			service := &PaymentMethodService{
				logger:            logger,
				paymentMethodRepo: paymentMethodRepo,
				snapCoreRepo:      snapCoreRepo,
				config:            cfg,
				redis:             redisClient,
			}

			// Execute the function
			result, err := service.GetStaticVAPaymentMethodByMerchant(ctx, tc.filter)

			// Assertions
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedResult), len(result))

				// Compare the results
				for i, expected := range tc.expectedResult {
					assert.Equal(t, expected.UUID, result[i].UUID)
					assert.Equal(t, expected.Name, result[i].Name)
					assert.Equal(t, expected.Type, result[i].Type)
					assert.Equal(t, expected.Acquirer, result[i].Acquirer)

					if expected.MerchantConfigObj != nil {
						assert.NotNil(t, result[i].MerchantConfigObj)
						assert.NotNil(t, result[i].MerchantConfigObj.PartnerConfig)
						assert.NotNil(t, result[i].MerchantConfigObj.PartnerConfig.VirtualAccount)

						expectedItems := expected.MerchantConfigObj.PartnerConfig.VirtualAccount.Items
						actualItems := result[i].MerchantConfigObj.PartnerConfig.VirtualAccount.Items

						assert.Equal(t, len(expectedItems), len(actualItems))
						for j, expectedItem := range expectedItems {
							assert.Equal(t, expectedItem.BINPrefix, actualItems[j].BINPrefix)
							assert.Equal(t, expectedItem.Type, actualItems[j].Type)
							assert.Equal(t, expectedItem.StartRange, actualItems[j].StartRange)
							assert.Equal(t, expectedItem.EndRange, actualItems[j].EndRange)
						}
					} else {
						assert.Nil(t, result[i].MerchantConfigObj)
					}
				}
			}

			// Verify mock expectations
			paymentMethodRepo.AssertExpectations(t)
			snapCoreRepo.AssertExpectations(t)
			redisClient.AssertExpectations(t)
		})
	}
}

func TestPaymentMethodService_isValidPaymentMethodByPaymentDetail(t *testing.T) {
	// Setup logger mock
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	// Setup service instance
	service := &PaymentMethodService{
		logger: logger,
	}

	tests := []struct {
		name             string
		payment          *paymentModel.PaymentDetailForPaymentUIResponse
		paymentMethodCfg *paymentModel.PaymentMethodWithPivot
		setupConfig      func(*PaymentMethodService)
		expectedResult   bool
	}{
		{
			name:             "should return false when payment is nil",
			payment:          nil,
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{},
			setupConfig:      func(s *PaymentMethodService) {},
			expectedResult:   false,
		},
		{
			name: "should return false when paymentMethodCfg is nil",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-123",
				Amount: commonModel.Amount{
					Value:    "100000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: nil,
			setupConfig:      func(s *PaymentMethodService) {},
			expectedResult:   false,
		},
		{
			name:             "should return false when both payment and paymentMethodCfg are nil",
			payment:          nil,
			paymentMethodCfg: nil,
			setupConfig:      func(s *PaymentMethodService) {},
			expectedResult:   false,
		},
		{
			name: "should return false when payment amount is invalid (not a number)",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-123",
				Amount: commonModel.Amount{
					Value:    "invalid-amount",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig:    func(s *PaymentMethodService) {},
			expectedResult: false,
		},
		{
			name: "should return true when amount is valid and no config exists for payment method type",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-123",
				Amount: commonModel.Amount{
					Value:    "100000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "UNKNOWN_TYPE",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				// Clear the config map
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedResult: true,
		},
		{
			name: "should return false when amount is less than minimum for VIRTUAL_ACCOUNT",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-123",
				Amount: commonModel.Amount{
					Value:    "5000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 1000000.0,
					},
				}
			},
			expectedResult: false,
		},
		{
			name: "should return false when amount is greater than maximum for VIRTUAL_ACCOUNT",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-123",
				Amount: commonModel.Amount{
					Value:    "2000000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 1000000.0,
					},
				}
			},
			expectedResult: false,
		},
		{
			name: "should return true when amount is within valid range for VIRTUAL_ACCOUNT",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-123",
				Amount: commonModel.Amount{
					Value:    "50000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {

				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 1000000.0,
					},
				}
			},
			expectedResult: true,
		},
		{
			name: "should return true when amount equals minimum for EWALLET",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-456",
				Amount: commonModel.Amount{
					Value:    "10000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "EWALLET",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"EWALLET": {
						MinAmount: 10000.0,
						MaxAmount: 5000000.0,
					},
				}
			},
			expectedResult: true,
		},
		{
			name: "should return true when amount equals maximum for QRIS",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-789",
				Amount: commonModel.Amount{
					Value:    "10000000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "QRIS",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"QRIS": {
						MinAmount: 1.0,
						MaxAmount: 10000000.0,
					},
				}
			},
			expectedResult: true,
		},
		{
			name: "should return false when amount is below minimum for CREDIT_CARD",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-999",
				Amount: commonModel.Amount{
					Value:    "500",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "CREDIT_CARD",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"CREDIT_CARD": {
						MinAmount: 1000.0,
						MaxAmount: 50000000.0,
					},
				}
			},
			expectedResult: false,
		},
		{
			name: "should handle decimal amounts correctly",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-decimal",
				Amount: commonModel.Amount{
					Value:    "15000.50",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 20000.0,
					},
				}
			},
			expectedResult: true,
		},
		{
			name: "should return true when payment method config is nil in map",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-no-config",
				Amount: commonModel.Amount{
					Value:    "50000",
					Currency: "IDR",
				},
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": nil,
				}
			},
			expectedResult: true,
		},
		{
			name: "should return true when payment expiry is within 3 months limit",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-expiry-valid",
				Amount: commonModel.Amount{
					Value:    "50000",
					Currency: "IDR",
				},
				ExpiredAt: func() *time.Time {
					t := time.Now().Add(60 * 24 * time.Hour) // 60 days from now
					return &t
				}(),
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Name: "BNI Virtual Account",
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 1000000.0,
					},
				}
			},
			expectedResult: true,
		},
		{
			name: "should log warning but return true when payment expiry exceeds 3 months limit",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-expiry-exceeded",
				Amount: commonModel.Amount{
					Value:    "50000",
					Currency: "IDR",
				},
				ExpiredAt: func() *time.Time {
					t := time.Now().Add(100 * 24 * time.Hour) // 100 days from now (exceeds 90 days)
					return &t
				}(),
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Name: "BCA Virtual Account",
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 90,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 1000000.0,
					},
				}
			},
			expectedResult: false,
		},
		{
			name: "should return true when payment expiry is exactly at 3 months limit",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-expiry-exact",
				Amount: commonModel.Amount{
					Value:    "50000",
					Currency: "IDR",
				},
				ExpiredAt: func() *time.Time {
					t := time.Now().Add(3 * 30 * 24 * time.Hour) // Exactly 90 days
					return &t
				}(),
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Name: "Mandiri Virtual Account",
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 1000000.0,
					},
				}
			},
			expectedResult: true,
		},
		{
			name: "should handle payment with zero value expiry time",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-zero-expiry",
				Amount: commonModel.Amount{
					Value:    "50000",
					Currency: "IDR",
				},
				ExpiredAt: func() *time.Time {
					t := time.Time{} // Zero time
					return &t
				}(),
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Name: "Permata Virtual Account",
					Type: "VIRTUAL_ACCOUNT",
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MinAmount: 10000.0,
						MaxAmount: 1000000.0,
					},
				}
			},
			expectedResult: true,
		},
		{
			name: "should validate expiry for EWALLET payment method",
			payment: &paymentModel.PaymentDetailForPaymentUIResponse{
				UUID: "payment-ewallet-expiry",
				Amount: commonModel.Amount{
					Value:    "75000",
					Currency: "IDR",
				},
				ExpiredAt: func() *time.Time {
					t := time.Now().Add(120 * 24 * time.Hour) // 120 days (exceeds limit)
					return &t
				}(),
			},
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Name: "OVO",
					Type: "EWALLET",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 90,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"EWALLET": {
						MinAmount: 10000.0,
						MaxAmount: 5000000.0,
					},
				}
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config for this test case
			tt.setupConfig(service)

			// Execute the function
			result := service.isValidPaymentMethodByPaymentDetail(context.Background(), tt.payment, tt.paymentMethodCfg)

			// Assert the result
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestPaymentMethodService_getPaymentMethodExpiryLimit(t *testing.T) {
	// Setup logger mock
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	tests := []struct {
		name             string
		paymentMethodCfg *paymentModel.PaymentMethodWithPivot
		setupConfig      func(*PaymentMethodService)
		expectedIsZero   bool
		description      string
	}{
		{
			name: "should return zero time when no config exists (service-level and payment-method)",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     0,
						MaxExpiryDurationUnit: "",
					},
				}
			},
			expectedIsZero: true,
			description:    "Both service-level and payment-method config are empty",
		},
		{
			name: "should return zero time when service-level config has zero duration",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     0,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
					},
				}
			},
			expectedIsZero: true,
			description:    "Service-level config has unit but zero duration",
		},
		{
			name: "should return zero time when service-level config has empty unit",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     90,
						MaxExpiryDurationUnit: "",
					},
				}
			},
			expectedIsZero: true,
			description:    "Service-level config has duration but empty unit",
		},
		{
			name: "should use service-level config when payment-method config is empty",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     90,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
					},
				}
			},
			expectedIsZero: false,
			description:    "Service-level config is valid, payment-method config is empty",
		},
		{
			name: "should use service-level config with MINUTES unit",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "QRIS",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"QRIS": {
						MaxExpiryDuration:     30,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
					},
				}
			},
			expectedIsZero: false,
			description:    "Service-level config with MINUTES unit",
		},
		{
			name: "should use service-level config with HOURS unit",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "EWALLET",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"EWALLET": {
						MaxExpiryDuration:     24,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
					},
				}
			},
			expectedIsZero: false,
			description:    "Service-level config with HOURS unit",
		},
		{
			name: "should override service-level config with payment-method config (precedence test)",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 60,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     90,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
					},
				}
			},
			expectedIsZero: false,
			description:    "Payment-method config should override service-level config",
		},
		{
			name: "should use payment-method config when service-level config doesn't exist",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "UNKNOWN_TYPE",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 30,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedIsZero: false,
			description:    "Payment-method config exists but no service-level config",
		},
		{
			name: "should use payment-method config with MINUTES unit",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 15,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitMinutes,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedIsZero: false,
			description:    "Payment-method config with MINUTES unit",
		},
		{
			name: "should use payment-method config with HOURS unit",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 48,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitHours,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedIsZero: false,
			description:    "Payment-method config with HOURS unit",
		},
		{
			name: "should fallback to service-level config when payment-method config has zero duration",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     90,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
					},
				}
			},
			expectedIsZero: false,
			description:    "Payment-method config has unit but zero duration, falls back to service-level config",
		},
		{
			name: "should fallback to service-level config when payment-method config has empty unit",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 90,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     90,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
					},
				}
			},
			expectedIsZero: false,
			description:    "Payment-method config has duration but empty unit, falls back to service-level config",
		},
		{
			name: "should handle nil service-level config for payment method type",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": nil,
				}
			},
			expectedIsZero: true,
			description:    "Service-level config exists but is nil",
		},
		{
			name: "should handle nil ConfigObj in payment method",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type:      "VIRTUAL_ACCOUNT",
					ConfigObj: nil,
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     90,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
					},
				}
			},
			expectedIsZero: false,
			description:    "ConfigObj is nil, should use service-level config",
		},
		{
			name: "should handle large duration values",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 365,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedIsZero: false,
			description:    "Large duration value (365 days)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup service instance
			service := &PaymentMethodService{
				logger: logger,
			}

			// Setup config for this test case
			tt.setupConfig(service)

			// Execute the function
			result := service.getPaymentMethodExpiryLimit(context.Background(), tt.paymentMethodCfg)

			// Assert the result
			if tt.expectedIsZero {
				assert.True(t, result.IsZero(), "Expected zero time but got: %v. %s", result, tt.description)
			} else {
				assert.False(t, result.IsZero(), "Expected non-zero time but got zero time. %s", tt.description)
				// Verify that the returned time is in the future
				assert.True(t, result.After(time.Now()), "Expected time to be in the future. %s", tt.description)
			}
		})
	}
}

func TestPaymentMethodService_getPaymentMethodExpiryLimit_CalculationAccuracy(t *testing.T) {
	// Setup logger mock
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	tests := []struct {
		name                string
		paymentMethodCfg    *paymentModel.PaymentMethodWithPivot
		setupConfig         func(*PaymentMethodService)
		expectedMinDuration time.Duration
		expectedMaxDuration time.Duration
		description         string
	}{
		{
			name: "should calculate correct time for 30 minutes",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "QRIS",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 30,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitMinutes,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedMinDuration: 29 * time.Minute,
			expectedMaxDuration: 31 * time.Minute,
			description:         "30 minutes duration",
		},
		{
			name: "should calculate correct time for 24 hours",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "EWALLET",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 24,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitHours,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedMinDuration: 23*time.Hour + 59*time.Minute,
			expectedMaxDuration: 24*time.Hour + 1*time.Minute,
			description:         "24 hours duration",
		},
		{
			name: "should calculate correct time for 90 days",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 90,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedMinDuration: 89*24*time.Hour + 23*time.Hour,
			expectedMaxDuration: 90*24*time.Hour + 1*time.Hour,
			description:         "90 days duration",
		},
		{
			name: "should calculate correct time for 1 minute",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "QRIS",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 1,
							Unit:     paymentConstant.UnifiedPaymentExpiryUnitMinutes,
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{}
			},
			expectedMinDuration: 0 * time.Minute,
			expectedMaxDuration: 2 * time.Minute,
			description:         "1 minute duration",
		},
		{
			name: "should use service-level config for calculation",
			paymentMethodCfg: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
					ConfigObj: &paymentModel.PaymentMethodConfig{
						ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
							Duration: 0,
							Unit:     "",
						},
					},
				},
			},
			setupConfig: func(s *PaymentMethodService) {
				s.paymentMethodValidationConfig = map[string]*PaymentMethodValidationConfig{
					"VIRTUAL_ACCOUNT": {
						MaxExpiryDuration:     7,
						MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
					},
				}
			},
			expectedMinDuration: 6*24*time.Hour + 23*time.Hour,
			expectedMaxDuration: 7*24*time.Hour + 1*time.Hour,
			description:         "7 days from service-level config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup service instance
			service := &PaymentMethodService{
				logger: logger,
			}

			// Setup config for this test case
			tt.setupConfig(service)

			// Capture time before calling the function
			timeBefore := time.Now()

			// Execute the function
			result := service.getPaymentMethodExpiryLimit(context.Background(), tt.paymentMethodCfg)

			// Capture time after calling the function
			timeAfter := time.Now()

			// Calculate the actual duration from now
			durationFromBefore := result.Sub(timeBefore)
			durationFromAfter := result.Sub(timeAfter)

			// Assert that the duration is within expected range
			// We use the before and after times to account for execution time
			assert.True(t,
				durationFromBefore >= tt.expectedMinDuration && durationFromAfter <= tt.expectedMaxDuration,
				"Expected duration between %v and %v, but got %v. %s",
				tt.expectedMinDuration, tt.expectedMaxDuration, durationFromBefore, tt.description)
		})
	}
}
