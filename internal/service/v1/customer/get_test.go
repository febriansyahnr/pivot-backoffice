package customerService

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCustomerList(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	testCases := []struct {
		name        string
		setup       func(customerRepo *repositoryMock.ICustomerRepository)
		merchantId  string
		phoneNumber string
		wantErr     bool
	}{
		{
			name: "SUCCESS: Get Customer list with phone number filter",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerList", mock.Anything, "parent-merchant-id", "09123456789", int64(1), int64(1)).Return(
					[]customerModel.Customer{}, &commonModel.Meta{}, nil)
			},
			merchantId:  "parent-merchant-id",
			phoneNumber: "09123456789",
		},
		{
			name: "ERROR: Get Customer list with phone number filter not found",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerList", mock.Anything, "parent-merchant-id", "09123456789", int64(1), int64(1)).Return(
					nil, nil, nil)
			},
			merchantId:  "parent-merchant-id",
			phoneNumber: "09123456789",
			wantErr:     true,
		},
		{
			name: "SUCCESS: Get Customer list",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerList", mock.Anything, "parent-merchant-id", "", int64(1), int64(1)).Return(
					[]customerModel.Customer{}, &commonModel.Meta{}, nil)
			},
			merchantId:  "parent-merchant-id",
			phoneNumber: "",
		},
		{
			name: "ERROR: Error when get custoemr list",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerList", mock.Anything, "parent-merchant-id", "", int64(1), int64(1)).Return(
					nil, nil, fmt.Errorf("Error get customer list"))
			},
			merchantId:  "parent-merchant-id",
			phoneNumber: "",
			wantErr:     true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)

			service := New(customerRepo, accountService, mockLogger)
			_, err := service.GetCustomerList(context.Background(), tc.merchantId, tc.phoneNumber, int64(1), int64(1))
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestGetCustomerById(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	testCases := []struct {
		name       string
		setup      func(customerRepo *repositoryMock.ICustomerRepository)
		merchantId string
		id         string
		wantErr    bool
	}{
		{
			name: "SUCCESS: Get Customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "parent-merchant-id").Return(
					&customerModel.Customer{
						UUID:       "customer-id",
						MerchantID: "parent-merchant-id",
					}, nil)
			},
			merchantId: "parent-merchant-id",
			id:         "customer-id",
		},
		{
			name: "FAILED: Failed when get customer, not found",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "non-parent-merchant-id").Return(
					nil, nil)
			},
			merchantId: "non-parent-merchant-id",
			id:         "customer-id",
			wantErr:    true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)

			service := New(customerRepo, accountService, mockLogger)
			_, err := service.GetCustomerById(context.Background(), tc.id, tc.merchantId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindCustomerByID(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	testCases := []struct {
		name       string
		setup      func(customerRepo *repositoryMock.ICustomerRepository)
		merchantId string
		id         string
		wantErr    bool
	}{
		{
			name: "SUCCESS: Get Customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("FindCustomerById", mock.Anything, "customer-id").Return(
					&customerModel.Customer{
						UUID:       "customer-id",
						MerchantID: "parent-merchant-id",
					}, nil)
			},
			id: "customer-id",
		},
		{
			name: "FAILED: Failed when get customer, not found",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("FindCustomerById", mock.Anything, "customer-id").Return(
					nil, nil)
			},
			id:      "customer-id",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)

			service := New(customerRepo, accountService, mockLogger)
			_, err := service.FindCustomerByID(context.Background(), tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetCustomerByPhoneNumber(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	now := time.Now().UTC()

	testCases := []struct {
		name        string
		setup       func(customerRepo *repositoryMock.ICustomerRepository)
		phoneNumber string
		merchantId  string
		wantErr     bool
	}{
		{
			name: "SUCCESS: Get Customer by phone number",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, "09123456789", "merchant-id").Return(
					&customerModel.Customer{
						UUID:        "customer-id",
						MerchantID:  "merchant-id",
						FirstName:   "John",
						LastName:    "Doe",
						Email:       "john.doe@example.com",
						PhoneNumber: "09123456789",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
			},
			phoneNumber: "09123456789",
			merchantId:  "merchant-id",
			wantErr:     false,
		},
		{
			name: "ERROR: Customer not found",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, "09999999999", "merchant-id").Return(
					nil, nil)
			},
			phoneNumber: "09999999999",
			merchantId:  "merchant-id",
			wantErr:     true,
		},
		{
			name: "ERROR: Database error",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, "09123456789", "merchant-id").Return(
					nil, fmt.Errorf("database connection error"))
			},
			phoneNumber: "09123456789",
			merchantId:  "merchant-id",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)

			service := New(customerRepo, accountService, mockLogger)
			result, err := service.GetCustomerByPhoneNumber(context.Background(), tc.phoneNumber, tc.merchantId)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.phoneNumber, result.PhoneNumber)
				assert.Equal(t, tc.merchantId, result.MerchantID)
			}

			customerRepo.AssertExpectations(t)
		})
	}
}

func TestGetCustomerByIDForUnifiedPayment(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	now := time.Now().UTC()

	cardPaymentMethod := &unifiedPaymentModel.CustomerPaymentMethod{
		Token:          "card-token-123",
		PaymentMethod:  "CARD",
		PaymentChannel: "VISA",
		Status:         "ACTIVE",
		CreatedAt:      now,
	}

	testCases := []struct {
		name       string
		setup      func(customerRepo *repositoryMock.ICustomerRepository)
		customerId string
		merchantId string
		wantErr    bool
	}{
		{
			name: "SUCCESS: Get Customer for unified payment",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "merchant-id").Return(
					&customerModel.Customer{
						UUID:        "customer-id",
						MerchantID:  "merchant-id",
						FirstName:   "John",
						LastName:    "Doe",
						Email:       "john.doe@example.com",
						PhoneNumber: "09123456789",
						CreatedAt:   now,
						UpdatedAt:   now,
						Metadata: map[string]interface{}{
							"paymentMethods": []*unifiedPaymentModel.CustomerPaymentMethod{
								cardPaymentMethod,
							},
						},
					}, nil)
			},
			customerId: "customer-id",
			merchantId: "merchant-id",
			wantErr:    false,
		},
		{
			name: "SUCCESS: Get Customer with no stored payment methods",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "merchant-id").Return(
					&customerModel.Customer{
						UUID:        "customer-id",
						MerchantID:  "merchant-id",
						FirstName:   "Jane",
						LastName:    "Smith",
						Email:       "jane.smith@example.com",
						PhoneNumber: "09987654321",
						CreatedAt:   now,
						UpdatedAt:   now,
						Metadata:    map[string]interface{}{},
					}, nil)
			},
			customerId: "customer-id",
			merchantId: "merchant-id",
			wantErr:    false,
		},
		{
			name: "ERROR: Customer not found",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "non-existent-id", "merchant-id").Return(
					nil, nil)
			},
			customerId: "non-existent-id",
			merchantId: "merchant-id",
			wantErr:    true,
		},
		{
			name: "ERROR: Database error",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "merchant-id").Return(
					nil, fmt.Errorf("database connection failed"))
			},
			customerId: "customer-id",
			merchantId: "merchant-id",
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)

			service := New(customerRepo, accountService, mockLogger)
			result, err := service.GetCustomerByIDForUnifiedPayment(context.Background(), tc.customerId, tc.merchantId)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.GivenName)
				assert.NotEmpty(t, result.SureName)
				assert.NotEmpty(t, result.Email)
			}

			customerRepo.AssertExpectations(t)
		})
	}
}

func TestGetCardFundedPayoutSavedCardDetail(t *testing.T) {
	pdkLog := loggerMock.NewILogger(t)
	repo := repositoryMock.NewICustomerRepository(t)

	service := New(repo, nil, pdkLog)

	var (
		cardID     = "58cdb5c7-2f9f-4446-b089-00761eb75c24"
		merchantID = "1d064989-7cae-4fa5-978a-732da1311bad"
	)
	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *cardFundedPayoutModel.GetSavedCardResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetCardFundedPayoutSavedCardDetail", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
				pdkLog.On("Error", mock.Anything, "Failed to retrieve card details", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetCardFundedPayoutSavedCardDetail", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("card details with ID %s were not found", cardID)),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetCardFundedPayoutSavedCardDetail", mock.Anything, mock.Anything,
				).Once().Return(&cardFundedPayoutModel.GetSavedCardResponse{ID: cardID}, nil)
			},
			wantResult: &cardFundedPayoutModel.GetSavedCardResponse{ID: cardID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := service.GetCardFundedPayoutSavedCardDetail(t.Context(), cardFundedPayoutModel.GetSavedCardDetailRequest{
				CardID:     cardID,
				MerchantID: merchantID,
			})
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
