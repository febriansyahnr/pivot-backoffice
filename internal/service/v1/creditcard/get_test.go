package creditcard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"

	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetPaymentById(t *testing.T) {
	paymentID := uuid.New()
	conf := &config.Config{}
	now := time.Now().UTC()
	expiredAt := now.Add(constant.CreditCardPaymentExpired)
	nowBefore, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	assert.NoError(t, err)
	alreadyExpiredAt := nowBefore.Add(constant.CreditCardPaymentExpired)

	referenceId := "some-reference-id"
	mid := "some-mid"

	creditCard := &paymentModel.Payment{
		UUID:        paymentID.String(),
		ReferenceID: &referenceId,
		MerchantID:  mid,
		Amount:      decimal.NewFromFloat(10000),
		Currency:    "IDR",
		PaymentURL:  "https://creditcard-webview-stg.harsya.com/payment/creditcard/pay/hOIuKiu-6NxhiFnJPMDWIke9qq0YsbpERh4Atnn-AEY=",
		Status:      "PENDING",
		ExpiredAt:   &expiredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	creditCardAlreadyExpired := &paymentModel.Payment{
		UUID:        paymentID.String(),
		ReferenceID: &referenceId,
		MerchantID:  mid,
		Amount:      decimal.NewFromFloat(10000),
		Currency:    "IDR",
		PaymentURL:  "https://creditcard-webview-stg.harsya.com/payment/creditcard/pay/hOIuKiu-6NxhiFnJPMDWIke9qq0YsbpERh4Atnn-AEY=",
		Status:      "PENDING",
		ExpiredAt:   &alreadyExpiredAt,
		CreatedAt:   nowBefore,
		UpdatedAt:   nowBefore,
	}

	testCases := []struct {
		name      string
		uuid      string
		mid       string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.IPaymentRepository, mockRedis *redisMocks.IRedisExt)
	}{
		{
			name:    "SUCCESS",
			uuid:    paymentID.String(),
			mid:     mid,
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IPaymentRepository, mockRedis *redisMocks.IRedisExt) {
				mockRepo.On("GetPaymentById", mock.Anything, mock.AnythingOfType("string")).Return(creditCard, nil)
			},
		},
		{
			name:    "SUCCESS: But expired",
			uuid:    paymentID.String(),
			mid:     mid,
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IPaymentRepository, mockRedis *redisMocks.IRedisExt) {
				mockRepo.On("GetPaymentById", mock.Anything, mock.AnythingOfType("string")).Return(creditCardAlreadyExpired, nil)
				mockRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
			},
		},
		{
			name:    "ERROR: Credit Card Not Found",
			uuid:    "non-existent-uuid",
			mid:     mid,
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IPaymentRepository, mockRedis *redisMocks.IRedisExt) {
				mockRepo.On("GetPaymentById", mock.Anything, mock.AnythingOfType("string")).Return(nil, nil)
				redisResult := &redis.StringCmd{}
				mockRedis.On("Get", mock.Anything, mock.Anything).Return(redisResult)
			},
		},
		{
			name:    "ERROR: Credit Card Not Found, Redis Error",
			uuid:    "non-existent-uuid",
			mid:     mid,
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IPaymentRepository, mockRedis *redisMocks.IRedisExt) {
				mockRepo.On("GetPaymentById", mock.Anything, mock.AnythingOfType("string")).Return(nil, nil)
				redisResult := &redis.StringCmd{}
				redisResult.SetErr(errors.New("error"))
				mockRedis.On("Get", mock.Anything, mock.Anything).Return(redisResult)
			},
		},
		{
			name:    "ERROR: Database Failure",
			uuid:    paymentID.String(),
			mid:     mid,
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IPaymentRepository, mockRedis *redisMocks.IRedisExt) {
				mockRepo.On("GetPaymentById", mock.Anything, mock.AnythingOfType("string")).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewIPaymentRepository(t)
			mockRedis := redisMocks.NewIRedisExt(t)

			ctx := context.Background()

			tc.mockSetup(mockRepo, mockRedis)

			svc := New(conf, mockLogger, nil, mockRepo, nil, nil, WithRedis(mockRedis))
			response, err := svc.GetPaymentById(ctx, tc.mid, tc.uuid)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetStoredCardByCustomerID(t *testing.T) {
	conf := &config.Config{}
	now := time.Now().UTC()
	customerID := uuid.New().String()
	merchantID := "some-merchant-id"

	cardPaymentMethod := &unifiedPaymentModel.CustomerPaymentMethod{
		Token:          "card-token-123",
		PaymentMethod:  constant.UnifiedPaymentMethodCard,
		PaymentChannel: "VISA",
		Status:         "ACTIVE",
		CreatedAt:      now,
		Card: &unifiedPaymentModel.CustomerPaymentMethodCard{
			Fingerprint: "card-fingerprint-123",
			Network:     "VISA",
			Last4:       "1234",
			ExpMonth:    "12",
			ExpYear:     "2025",
		},
	}

	ewalletPaymentMethod := &unifiedPaymentModel.CustomerPaymentMethod{
		Token:          "ewallet-token-123",
		PaymentMethod:  "EWALLET",
		PaymentChannel: "OVO",
		Status:         "ACTIVE",
		CreatedAt:      now,
	}

	customer := &customerModel.Customer{
		UUID:      customerID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata: map[string]interface{}{
			"paymentMethods": []*unifiedPaymentModel.CustomerPaymentMethod{
				cardPaymentMethod,
				ewalletPaymentMethod,
			},
		},
	}

	testCases := []struct {
		name              string
		customerID        string
		merchantID        string
		wantErr           bool
		expectedCardCount int
		mockSetup         func(mockCustomerRepo *repositoryMocks.ICustomerRepository)
	}{
		{
			name:              "SUCCESS - Customer with stored cards",
			customerID:        customerID,
			merchantID:        merchantID,
			wantErr:           false,
			expectedCardCount: 1,
			mockSetup: func(mockCustomerRepo *repositoryMocks.ICustomerRepository) {
				mockCustomerRepo.On("GetCustomerById", mock.Anything, customerID, merchantID).Return(customer, nil)
			},
		},
		{
			name:              "SUCCESS - Customer with no stored cards",
			customerID:        customerID,
			merchantID:        merchantID,
			wantErr:           false,
			expectedCardCount: 0,
			mockSetup: func(mockCustomerRepo *repositoryMocks.ICustomerRepository) {
				customerWithoutCards := &customerModel.Customer{
					UUID:      customerID,
					FirstName: "Jane",
					LastName:  "Doe",
					Email:     "jane.doe@example.com",
					CreatedAt: now,
					UpdatedAt: now,
					Metadata: map[string]interface{}{
						"paymentMethods": []*unifiedPaymentModel.CustomerPaymentMethod{
							ewalletPaymentMethod,
						},
					},
				}
				mockCustomerRepo.On("GetCustomerById", mock.Anything, customerID, merchantID).Return(customerWithoutCards, nil)
			},
		},
		{
			name:       "ERROR - Customer not found",
			customerID: "non-existent-customer",
			merchantID: merchantID,
			wantErr:    true,
			mockSetup: func(mockCustomerRepo *repositoryMocks.ICustomerRepository) {
				mockCustomerRepo.On("GetCustomerById", mock.Anything, "non-existent-customer", merchantID).Return(nil, nil)
			},
		},
		{
			name:       "ERROR - Database error",
			customerID: customerID,
			merchantID: merchantID,
			wantErr:    true,
			mockSetup: func(mockCustomerRepo *repositoryMocks.ICustomerRepository) {
				mockCustomerRepo.On("GetCustomerById", mock.Anything, customerID, merchantID).Return(nil, errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCustomerRepo := repositoryMocks.NewICustomerRepository(t)

			ctx := context.Background()

			tc.mockSetup(mockCustomerRepo)

			svc := New(conf, mockLogger, nil, nil, nil, nil, WithCustomerRepo(mockCustomerRepo))
			response, err := svc.GetStoredCardByCustomerID(ctx, tc.merchantID, tc.customerID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response, tc.expectedCardCount)

				for _, cardMethod := range response {
					assert.Equal(t, constant.UnifiedPaymentMethodCard, cardMethod.PaymentMethod)
					assert.NotEmpty(t, cardMethod.Token)
					assert.NotEmpty(t, cardMethod.Status)
				}
			}

			mockCustomerRepo.AssertExpectations(t)
		})
	}
}

func TestGetCardEncryptionPublicKey(t *testing.T) {
	cardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

	service := New(nil, nil, nil, nil, nil, cardRepo)

	merchantID := "a1c9d94f-d8af-43c0-870e-760977da044e"

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []byte
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				cardRepo.On("GetCardEncryptionPublicKey", mock.Anything, merchantID).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				cardRepo.On("GetCardEncryptionPublicKey", mock.Anything, merchantID).Once().Return([]byte(`public-key`), nil)
			},
			wantResult: []byte(`public-key`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetCardEncryptionPublicKey(t.Context(), merchantID)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
