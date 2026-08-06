package unifiedPaymentService_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	unifiedPaymentService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChargeList(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	tests := []struct {
		name           string
		request        *unifiedPaymentModel.FilterChargeRequest
		mockSetup      func(*repositoryMock.IPaymentRepository)
		expectedResult *commonModel.PaginationResponse
		expectedError  error
	}{
		{
			name: "SUCCESS: Valid request returns charge list",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: "merchant-" + uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			mockSetup: func(paymentRepo *repositoryMock.IPaymentRepository) {
				expectedResult := &commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 0,
						TotalPages: 0,
					},
				}
				paymentRepo.On("GetChargeList", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.FilterChargeRequest")).Return(expectedResult, nil)
			},
			expectedResult: &commonModel.PaginationResponse{
				Data: []*unifiedPaymentModel.ChargeResponse{},
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			expectedError: nil,
		},
		{
			name: "ERROR: Database error",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: "merchant-" + uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			mockSetup: func(paymentRepo *repositoryMock.IPaymentRepository) {
				databaseError := errors.New("database connection failed")
				paymentRepo.On("GetChargeList", mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.FilterChargeRequest")).Return(nil, databaseError)
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrDatabase, errors.New("database connection failed")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)

			tt.mockSetup(paymentRepo)

			svc := unifiedPaymentService.New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo)
			result, err := svc.GetChargeList(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}

func TestGetChargeDetail(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := "merchant-" + uuid.NewString()
	chargeID := "charge-" + uuid.NewString()
	differentMerchantID := "different-merchant-" + uuid.NewString()

	tests := []struct {
		name           string
		request        *unifiedPaymentModel.GetUnifiedPaymentChargeRequest
		mockSetup      func(*repositoryMock.IPaymentRepository, *repositoryMock.IStatusHistoriesRepository)
		expectedResult *unifiedPaymentModel.ChargeResponse
		expectedError  error
	}{
		{
			name: "SUCCESS: Valid request returns charge detail",
			request: &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
				ChargeID:   chargeID,
				MerchantID: merchantID,
			},
			mockSetup: func(paymentRepo *repositoryMock.IPaymentRepository, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				expectedCharge := &unifiedPaymentModel.ChargeResponse{
					ID:               chargeID,
					MerchantID:       merchantID,
					PaymentSessionID: "payment-session-id",
					Status:           "PAID",
				}
				paymentRepo.On("GetChargeByID", mock.Anything, chargeID).Return(expectedCharge, nil)
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-session-id").Return(nil, nil)
				statusHistoriesRepo.On("GetByReference", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedResult: &unifiedPaymentModel.ChargeResponse{
				ID:               chargeID,
				MerchantID:       merchantID,
				PaymentSessionID: "payment-session-id",
				Status:           "PAID",
			},
			expectedError: nil,
		},
		{
			name: "ERROR: Database error when getting charge",
			request: &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
				ChargeID:   chargeID,
				MerchantID: merchantID,
			},
			mockSetup: func(paymentRepo *repositoryMock.IPaymentRepository, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				databaseError := errors.New("database connection failed")
				paymentRepo.On("GetChargeByID", mock.Anything, chargeID).Return(nil, databaseError)
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrDatabase, errors.New("database connection failed")),
		},
		{
			name: "ERROR: Charge not found",
			request: &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
				ChargeID:   chargeID,
				MerchantID: merchantID,
			},
			mockSetup: func(paymentRepo *repositoryMock.IPaymentRepository, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				paymentRepo.On("GetChargeByID", mock.Anything, chargeID).Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrUnprocessableContent, c.ErrPaymentChargeNotFound),
		},
		{
			name: "ERROR: Merchant ID mismatch",
			request: &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
				ChargeID:   chargeID,
				MerchantID: merchantID,
			},
			mockSetup: func(paymentRepo *repositoryMock.IPaymentRepository, statusHistoriesRepo *repositoryMock.IStatusHistoriesRepository) {
				charge := &unifiedPaymentModel.ChargeResponse{
					ID:         chargeID,
					MerchantID: differentMerchantID,
					Status:     "PAID",
				}
				paymentRepo.On("GetChargeByID", mock.Anything, chargeID).Return(charge, nil)
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrUnprocessableContent, c.ErrMerchantIsNotMatch),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			statusHistoriesRepo := repositoryMock.NewIStatusHistoriesRepository(t)

			tt.mockSetup(paymentRepo, statusHistoriesRepo)

			svc := unifiedPaymentService.New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				unifiedPaymentService.WithStatusHistoriesRepository(statusHistoriesRepo))
			result, err := svc.GetChargeDetail(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}

func TestExportCharge(t *testing.T) {
	rdb, mocks := redismock.NewClientMock()

	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	storage := gcsMocks.NewGCSService(t)

	svc := unifiedPaymentService.New(
		cfg,
		log,
		paymentRepo,
		paymentMethodRepo,
		accountTrxRepo,
		unifiedPaymentService.WithCache(redisExt.WrapRedisClient(rdb, nil)),
		unifiedPaymentService.WithStorage(storage),
	)

	request := &unifiedPaymentModel.FilterChargeRequest{
		MerchantID:     "merchant-" + uuid.NewString(),
		StartCreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndCreatedAt:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Page:           1,
		PerPage:        10,
	}
	hashFilter := request.HashFilter(c.TimeLoc)

	downloadCache := &commonModel.ExportResponse{
		DownloadURL: "https://example.com/download",
		ExpiresAt:   time.Now().UTC().Add(15 * time.Minute),
	}
	rawDownloadCache, _ := json.Marshal(downloadCache)

	cacheKey := fmt.Sprintf(c.RedisKeyDownloadChargeHistoryFmt, hashFilter)

	tests := []struct {
		name           string
		setupMock      func()
		expectedResult *commonModel.ExportResponse
		expectedError  error
	}{
		{
			name: "SUCCESS: Cache found",
			setupMock: func() {
				mocks.ExpectGet(cacheKey).SetVal(string(rawDownloadCache))
			},
			expectedResult: downloadCache,
			expectedError:  nil,
		},
		{
			name: "ERROR: Get download cache non-nil error",
			setupMock: func() {
				mocks.ExpectGet(cacheKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR: Get charge list from repository",
			setupMock: func() {
				mocks.ExpectGet(cacheKey).RedisNil()
				paymentRepo.On("GetCharges", mock.Anything, request).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := svc.ExportCharge(context.Background(), request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.DownloadURL, result.DownloadURL)
					assert.NotZero(t, result.ExpiresAt)
				}
			}

			paymentRepo.AssertExpectations(t)
			storage.AssertExpectations(t)
			assert.NoError(t, mocks.ExpectationsWereMet())
		})
	}
}
