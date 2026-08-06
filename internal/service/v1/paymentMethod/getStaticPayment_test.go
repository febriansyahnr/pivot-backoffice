package paymentMethodService_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/paymentMethod"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetStaticQRPaymentMethodByMerchant(t *testing.T) {
	ctx := context.Background()
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	merchantID := uuid.NewString()
	paymentMethodID := uuid.NewString()
	storeID := uuid.NewString()

	testCases := []struct {
		name           string
		filter         *paymentModel.GetPaymentMethodFilterRequest
		setupMocks     func(*repositoryMocks.IPaymentMethodRepository, *repositoryMocks.IPaymentRepository)
		expectedResult *paymentModel.PaymentMethodWithPivot
		expectedError  error
	}{
		{
			name: "SUCCESS: Get static QR payment method with QR payments",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: merchantID,
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, paymentRepo *repositoryMocks.IPaymentRepository) {
				// Mock GetActivePaymentMethodByRequest
				expectedPaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     paymentMethodID,
						Type:     constant.ChannelQris,
						Category: constant.ProductPayment,
						Name:     "Static QR Payment",
						Acquirer: "BNI",
					},
					IsActive:         true,
					ActivationStatus: constant.StatusActive,
				}

				expectedFilter := &paymentModel.GetPaymentMethodFilterRequest{
					MerchantID: merchantID,
					Category:   constant.ProductPayment,
					Status:     constant.StatusActive,
					Type:       constant.ChannelQris,
				}

				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything,
					mock.MatchedBy(func(filter *paymentModel.GetPaymentMethodFilterRequest) bool {
						return filter.MerchantID == expectedFilter.MerchantID &&
							filter.Category == expectedFilter.Category &&
							filter.Status == expectedFilter.Status &&
							filter.Type == expectedFilter.Type
					})).Return(expectedPaymentMethod, nil)

				// Mock FilterStaticQrisList
				now := time.Now()
				expiredAt := now.Add(24 * time.Hour)
				staticQrisData := []paymentModel.StaticQrisListResponse{
					{
						UUID:        "payment-uuid-1",
						ReferenceID: "ref-123",
						MerchantID:  merchantID,
						QrContent:   "qr-content-1",
						QrUrl:       "https://qr.url/1",
						StoreID:     storeID,
						Status:      constant.StatusActive,
						CreatedAt:   now,
						ExpiredAt:   &expiredAt,
					},
					{
						UUID:        "payment-uuid-2",
						ReferenceID: "ref-456",
						MerchantID:  "different-merchant-id", // Different merchant ID to test IsOnBehalf
						QrContent:   "qr-content-2",
						QrUrl:       "https://qr.url/2",
						StoreID:     storeID,
						Status:      constant.StatusActive,
						CreatedAt:   now,
						ExpiredAt:   &expiredAt,
					},
				}

				filterResult := &commonModel.PaginationResponse{
					Data: staticQrisData,
				}

				expectedStaticFilter := paymentModel.StaticQrisFilterRequest{
					MerchantID:      merchantID,
					Status:          constant.StatusActive,
					PaymentMethodID: paymentMethodID,
				}

				paymentRepo.On("FilterStaticQrisList", mock.Anything,
					mock.MatchedBy(func(filter paymentModel.StaticQrisFilterRequest) bool {
						return filter.MerchantID == expectedStaticFilter.MerchantID &&
							filter.Status == expectedStaticFilter.Status &&
							filter.PaymentMethodID == expectedStaticFilter.PaymentMethodID
					})).Return(filterResult, nil)
			},
			expectedResult: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					UUID:     paymentMethodID,
					Type:     constant.ChannelQris,
					Category: constant.ProductPayment,
					Name:     "Static QR Payment",
					Acquirer: "BNI",
				},
				IsActive:         true,
				ActivationStatus: constant.StatusActive,
				QRPayments: []paymentModel.StaticQRPaymentItem{
					{
						MerchantID:               merchantID,
						PaymentSessionID:         "payment-uuid-1",
						PaymentClientReferenceID: "ref-123",
						StoreID:                  storeID,
						IsDerived:                false,
						CreatedAt:                time.Time{}, // Will be set in test
						ExpiredAt:                time.Time{}, // Will be set in test
					},
					{
						MerchantID:               "different-merchant-id",
						PaymentSessionID:         "payment-uuid-2",
						PaymentClientReferenceID: "ref-456",
						StoreID:                  storeID,
						IsDerived:                true,        // Different merchant ID
						CreatedAt:                time.Time{}, // Will be set in test
						ExpiredAt:                time.Time{}, // Will be set in test
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "SUCCESS: Get static QR payment method without QR payments",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: merchantID,
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, paymentRepo *repositoryMocks.IPaymentRepository) {
				// Mock GetActivePaymentMethodByRequest
				expectedPaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     paymentMethodID,
						Type:     constant.ChannelQris,
						Category: constant.ProductPayment,
						Name:     "Static QR Payment",
						Acquirer: "BNI",
					},
					IsActive:         true,
					ActivationStatus: constant.StatusActive,
				}

				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Return(expectedPaymentMethod, nil)

				// Mock FilterStaticQrisList - empty result
				filterResult := &commonModel.PaginationResponse{
					Data: []paymentModel.StaticQrisListResponse{},
				}

				paymentRepo.On("FilterStaticQrisList", mock.Anything, mock.Anything).Return(filterResult, nil)
			},
			expectedResult: &paymentModel.PaymentMethodWithPivot{
				PaymentMethod: paymentModel.PaymentMethod{
					UUID:     paymentMethodID,
					Type:     constant.ChannelQris,
					Category: constant.ProductPayment,
					Name:     "Static QR Payment",
					Acquirer: "BNI",
				},
				IsActive:         true,
				ActivationStatus: constant.StatusActive,
				QRPayments:       nil, // No QR payments
			},
			expectedError: nil,
		},
		{
			name: "ERROR: GetActivePaymentMethodByRequest returns database error",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: merchantID,
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, paymentRepo *repositoryMocks.IPaymentRepository) {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrDatabase, errors.New("database error")),
		},
		{
			name: "ERROR: GetActivePaymentMethodByRequest returns nil payment method",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: merchantID,
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, paymentRepo *repositoryMocks.IPaymentRepository) {
				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound),
		},
		{
			name: "ERROR: FilterStaticQrisList returns database error",
			filter: &paymentModel.GetPaymentMethodFilterRequest{
				MerchantID: merchantID,
			},
			setupMocks: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository, paymentRepo *repositoryMocks.IPaymentRepository) {
				// Mock GetActivePaymentMethodByRequest success
				expectedPaymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     paymentMethodID,
						Type:     constant.ChannelQris,
						Category: constant.ProductPayment,
						Name:     "Static QR Payment",
						Acquirer: "BNI",
					},
					IsActive:         true,
					ActivationStatus: constant.StatusActive,
				}

				paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything, mock.Anything).Return(expectedPaymentMethod, nil)

				// Mock FilterStaticQrisList error
				paymentRepo.On("FilterStaticQrisList", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedResult: nil,
			expectedError:  pkgErrors.New(response.HttpErrDatabase, errors.New("database error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tc.setupMocks(paymentMethodRepo, paymentRepo)

			// Create service
			service := paymentMethodService.New(
				logger,
				paymentMethodRepo,
				snapCoreRepo,
				creditCardRepo,
				paymentMethodService.WithPaymentRepository(paymentRepo),
			)

			// Execute the function
			result, err := service.GetStaticQRPaymentMethodByMerchant(ctx, tc.filter)

			// Assertions
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Compare basic payment method properties
				assert.Equal(t, tc.expectedResult.UUID, result.UUID)
				assert.Equal(t, tc.expectedResult.Name, result.Name)
				assert.Equal(t, tc.expectedResult.Type, result.Type)
				assert.Equal(t, tc.expectedResult.Category, result.Category)
				assert.Equal(t, tc.expectedResult.Acquirer, result.Acquirer)
				assert.Equal(t, tc.expectedResult.IsActive, result.IsActive)
				assert.Equal(t, tc.expectedResult.ActivationStatus, result.ActivationStatus)

				// Compare QR payments
				if tc.expectedResult.QRPayments == nil {
					assert.Nil(t, result.QRPayments)
				} else {
					assert.NotNil(t, result.QRPayments)
					assert.Equal(t, len(tc.expectedResult.QRPayments), len(result.QRPayments))

					for i, expectedQR := range tc.expectedResult.QRPayments {
						actualQR := result.QRPayments[i]
						assert.Equal(t, expectedQR.MerchantID, actualQR.MerchantID)
						assert.Equal(t, expectedQR.PaymentSessionID, actualQR.PaymentSessionID)
						assert.Equal(t, expectedQR.PaymentClientReferenceID, actualQR.PaymentClientReferenceID)
						assert.Equal(t, expectedQR.StoreID, actualQR.StoreID)
						assert.Equal(t, expectedQR.IsDerived, actualQR.IsDerived)
						// Note: CreatedAt and ExpiredAt are set from the actual data, so we don't need to compare exact values
						assert.NotZero(t, actualQR.CreatedAt)
						assert.NotZero(t, actualQR.ExpiredAt)
					}
				}
			}

			// Verify mock expectations
			paymentMethodRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
		})
	}
}

func TestGetStaticQRPaymentMethodByMerchant_FilterValidation(t *testing.T) {
	ctx := context.Background()
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	merchantID := uuid.NewString()
	paymentMethodID := uuid.NewString()

	t.Run("SUCCESS: Verify filter parameters are set correctly", func(t *testing.T) {
		// Setup mocks
		paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
		paymentRepo := repositoryMocks.NewIPaymentRepository(t)
		snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
		creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

		// Mock GetActivePaymentMethodByRequest with exact filter validation
		expectedPaymentMethod := &paymentModel.PaymentMethodWithPivot{
			PaymentMethod: paymentModel.PaymentMethod{
				UUID:     paymentMethodID,
				Type:     constant.ChannelQris,
				Category: constant.ProductPayment,
				Name:     "Static QR Payment",
				Acquirer: "BNI",
			},
		}

		paymentMethodRepo.On("GetActivePaymentMethodByRequest", mock.Anything,
			mock.MatchedBy(func(filter *paymentModel.GetPaymentMethodFilterRequest) bool {
				// Validate that the service sets the correct filter parameters
				return filter.MerchantID == merchantID &&
					filter.Category == constant.ProductPayment &&
					filter.Status == constant.StatusActive &&
					filter.Type == constant.ChannelQris
			})).Return(expectedPaymentMethod, nil)

		// Mock FilterStaticQrisList with exact filter validation
		filterResult := &commonModel.PaginationResponse{
			Data: []paymentModel.StaticQrisListResponse{},
		}

		paymentRepo.On("FilterStaticQrisList", mock.Anything,
			mock.MatchedBy(func(filter paymentModel.StaticQrisFilterRequest) bool {
				// Validate that the service sets the correct static QR filter parameters
				return filter.MerchantID == merchantID &&
					filter.Status == constant.StatusActive &&
					filter.PaymentMethodID == paymentMethodID
			})).Return(filterResult, nil)

		// Create service
		service := paymentMethodService.New(
			logger,
			paymentMethodRepo,
			snapCoreRepo,
			creditCardRepo,
			paymentMethodService.WithPaymentRepository(paymentRepo),
		)

		// Execute with minimal filter
		inputFilter := &paymentModel.GetPaymentMethodFilterRequest{
			MerchantID: merchantID,
		}

		result, err := service.GetStaticQRPaymentMethodByMerchant(ctx, inputFilter)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify that the input filter was modified by the service
		assert.Equal(t, constant.ProductPayment, inputFilter.Category)
		assert.Equal(t, constant.StatusActive, inputFilter.Status)
		assert.Equal(t, constant.ChannelQris, inputFilter.Type)

		// Verify mock expectations
		paymentMethodRepo.AssertExpectations(t)
		paymentRepo.AssertExpectations(t)
	})
}
