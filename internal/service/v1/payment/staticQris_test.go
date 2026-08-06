package paymentService_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestFilterStaticQrisList(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	expectedResponse := &commonModel.PaginationResponse{
		Data: []paymentModel.StaticQrisListResponse{
			{
				UUID:        "payment-123",
				ReferenceID: "ref-123",
				Status:      "ACTIVE",
			},
		},
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    10,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(*repositoryMocks.IPaymentRepository)
		request   paymentModel.StaticQrisFilterRequest
	}{
		{
			name:    "SUCCESS: Filter static QRIS list",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("FilterStaticQrisList",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticQrisFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
		},
		{
			name:    "ERROR: Repository error",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("FilterStaticQrisList",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticQrisFilterRequest")).
					Return((*commonModel.PaginationResponse)(nil), constant.ErrSomeErrorForUnitTest)
			},
			request: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			tc.setupMock(paymentRepo)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			response, err := paymentSvc.FilterStaticQrisList(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.IsType(t, &commonModel.PaginationResponse{}, response)
			}

		})
	}
}

func TestGetStaticQrisDetail(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	expectedResponse := &paymentModel.StaticQrisDetailResponse{
		UUID:        "payment-123",
		ReferenceID: "ref-123",
		Status:      "ACTIVE",
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(*repositoryMocks.IPaymentRepository)
		request   paymentModel.StaticQrisDetailRequest
	}{
		{
			name:    "SUCCESS: Get static QRIS detail",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticQrisDetail",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticQrisDetailRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticQrisDetailRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
			},
		},
		{
			name:    "ERROR: Repository error",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticQrisDetail",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticQrisDetailRequest")).
					Return((*paymentModel.StaticQrisDetailResponse)(nil), constant.ErrSomeErrorForUnitTest)
			},
			request: paymentModel.StaticQrisDetailRequest{
				PaymentID:  "invalid-payment",
				MerchantID: "merchant-123",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			tc.setupMock(paymentRepo)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			response, err := paymentSvc.GetStaticQrisDetail(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.IsType(t, &paymentModel.StaticQrisDetailResponse{}, response)
			}

		})
	}
}

func TestGetStaticQrisTransactions(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	expectedResponse := &commonModel.PaginationResponse{
		Data: []paymentModel.StaticQrisTransactionItem{
			{
				UUID:        "tx-123",
				ReferenceID: "tx-123",
				AmountValue: "10000",
				Status:      "SUCCESS",
			},
		},
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    10,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(*repositoryMocks.IPaymentRepository)
		request   paymentModel.StaticQrisTransactionFilterRequest
	}{
		{
			name:    "SUCCESS: Get static QRIS transactions",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticQrisTransactions",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticQrisTransactionFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "paymentDate",
			},
		},
		{
			name:    "ERROR: Repository error",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticQrisTransactions",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticQrisTransactionFilterRequest")).
					Return((*commonModel.PaginationResponse)(nil), constant.ErrSomeErrorForUnitTest)
			},
			request: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "invalid-payment",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			tc.setupMock(paymentRepo)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			response, err := paymentSvc.GetStaticQrisTransactions(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.IsType(t, &commonModel.PaginationResponse{}, response)
			}

		})
	}
}

func TestDeactivateStaticQris(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name       string
		wantErr    bool
		setupMock  func(*repositoryMocks.IPaymentRepository)
		paymentID  string
		merchantID string
		request    paymentModel.StaticQrisUpdateStatusRequest
	}{
		{
			name:    "SUCCESS: Deactivate static QRIS",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// Mock getting current payment status
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{Status: "ACTIVE"}, nil)
				// Mock updating payment status
				repo.On("UpdatePaymentStatus",
					constant.ValueCtxMockType(), "payment-123", "merchant-123", "INACTIVE", mock.AnythingOfType("time.Time")).
					Return(nil)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticQrisUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Payment not currently ACTIVE",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// Mock getting current payment status as already INACTIVE
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{Status: "INACTIVE"}, nil)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticQrisUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Repository error on getting payment",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return((*paymentModel.Payment)(nil), constant.ErrSomeErrorForUnitTest)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticQrisUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Repository error on updating status",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{Status: "ACTIVE"}, nil)
				repo.On("UpdatePaymentStatus",
					constant.ValueCtxMockType(), "payment-123", "merchant-123", "INACTIVE", mock.AnythingOfType("time.Time")).
					Return(constant.ErrSomeErrorForUnitTest)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticQrisUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Empty payment ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// No mock setup needed as validation happens before repository call
			},
			paymentID:  "",
			merchantID: "merchant-123",
			request: paymentModel.StaticQrisUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Empty merchant ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// No mock setup needed as validation happens before repository call
			},
			paymentID:  "payment-123",
			merchantID: "",
			request: paymentModel.StaticQrisUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			tc.setupMock(paymentRepo)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			err := paymentSvc.DeactivateStaticQris(ctx, tc.paymentID, tc.merchantID, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestGetMaxActiveStaticQRPerMerchant(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		config         *config.Config
		expectedResult int
	}{
		{
			name: "SUCCESS: Config has valid value",
			config: &config.Config{
				UnifiedPaymentConfig: config.UnifiedPaymentConfig{
					QrConfig: &config.UnifiedPaymentQrConfig{
						MaxActiveStaticQRPerMerchant: 5,
					},
				},
			},
			expectedResult: 5,
		},
		{
			name: "SUCCESS: Config has zero value",
			config: &config.Config{
				UnifiedPaymentConfig: config.UnifiedPaymentConfig{
					QrConfig: &config.UnifiedPaymentQrConfig{
						MaxActiveStaticQRPerMerchant: 0,
					},
				},
			},
			expectedResult: 0,
		},
		{
			name: "SUCCESS: Config has large number",
			config: &config.Config{
				UnifiedPaymentConfig: config.UnifiedPaymentConfig{
					QrConfig: &config.UnifiedPaymentQrConfig{
						MaxActiveStaticQRPerMerchant: 100,
					},
				},
			},
			expectedResult: 100,
		},
		{
			name:           "SUCCESS: Config is nil",
			config:         nil,
			expectedResult: 0,
		},
		{
			name: "SUCCESS: QrConfig is nil",
			config: &config.Config{
				UnifiedPaymentConfig: config.UnifiedPaymentConfig{
					QrConfig: nil,
				},
			},
			expectedResult: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			// Set config if provided
			if tc.config != nil {
				WithConfig(tc.config)(paymentSvc)
			}

			result := paymentSvc.GetMaxActiveStaticQRPerMerchant()
			assert.Equal(t, tc.expectedResult, result)

		})
	}
}

func TestGetFirstActiveStaticQris(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	refID := "ref-123"
	expectedPayment := &paymentModel.Payment{
		UUID:        "payment-123",
		ReferenceID: &refID,
		MerchantID:  "merchant-123",
		Status:      constant.StatusActive,
		Type:        constant.UnifiedPaymentTypeMultiple,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	testCases := []struct {
		name               string
		wantErr            bool
		setupMock          func(*repositoryMocks.IPaymentRepository)
		merchantID         string
		partnerReferenceNo string
	}{
		{
			name:    "SUCCESS: Get first active static QRIS payment",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetFirstActiveStaticQrisByMerchant",
					constant.ValueCtxMockType(), "merchant-123", "ref-123").
					Return(expectedPayment, nil)
			},
			merchantID:         "merchant-123",
			partnerReferenceNo: "ref-123",
		},
		{
			name:    "ERROR: Repository error",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetFirstActiveStaticQrisByMerchant",
					constant.ValueCtxMockType(), "merchant-123", "ref-error").
					Return((*paymentModel.Payment)(nil), constant.ErrSomeErrorForUnitTest)
			},
			merchantID:         "merchant-123",
			partnerReferenceNo: "ref-error",
		},
		{
			name:    "ERROR: Empty merchant ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// No mock setup needed as validation happens before repository call
			},
			merchantID:         "",
			partnerReferenceNo: "ref-empty",
		},
		{
			name:    "ERROR: No records found",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetFirstActiveStaticQrisByMerchant",
					constant.ValueCtxMockType(), "merchant-456", "ref-456").
					Return((*paymentModel.Payment)(nil), fmt.Errorf("no active static QRIS found for merchant"))
			},
			merchantID:         "merchant-456",
			partnerReferenceNo: "ref-456",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			tc.setupMock(paymentRepo)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			payment, err := paymentSvc.GetFirstActiveStaticQris(ctx, tc.merchantID, tc.partnerReferenceNo)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, expectedPayment.UUID, payment.UUID)
				assert.Equal(t, expectedPayment.MerchantID, payment.MerchantID)
				assert.Equal(t, expectedPayment.Status, payment.Status)
				assert.Equal(t, expectedPayment.Type, payment.Type)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
