package paymentService_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestFilterStaticVaList(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	expectedResponse := &commonModel.PaginationResponse{
		Data: []paymentModel.StaticVaListResponse{
			{
				UUID:        "payment-123",
				ReferenceID: "VAname234",
				VaNumber:    "1234567890987654",
				VaBank:      "BCA",
				VaBankLogo:  "https://storage.googleapis.com/path/to/bca-logo.png",
				VaName:      "VAname234",
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
		request   paymentModel.StaticVaFilterRequest
	}{
		{
			name:    "SUCCESS: Filter static VA list",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("FilterStaticVaList",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
		},
		{
			name:    "SUCCESS: Filter static VA list with query",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("FilterStaticVaList",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				ID:         "VAname234",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
		},
		{
			name:    "SUCCESS: Filter static VA list with status",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("FilterStaticVaList",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
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
				repo.On("FilterStaticVaList",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaFilterRequest")).
					Return((*commonModel.PaginationResponse)(nil), constant.ErrSomeErrorForUnitTest)
			},
			request: paymentModel.StaticVaFilterRequest{
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
			response, err := paymentSvc.FilterStaticVaList(ctx, tc.request)
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

func TestGetStaticVaDetail(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	expectedResponse := &paymentModel.StaticVaDetailResponse{
		UUID:        "payment-123",
		ReferenceID: "VAname234",
		VaNumber:    "1234567890987654",
		VaBank:      "BCA",
		VaBankLogo:  "https://storage.googleapis.com/path/to/bca-logo.png",
		VaName:      "VAname234",
		VaIssuer:    "PT BANK CENTRAL ASIA TBK",
		VaType:      "Open Virtual Account",
		Status:      "ACTIVE",
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(*repositoryMocks.IPaymentRepository)
		request   paymentModel.StaticVaDetailRequest
	}{
		{
			name:    "SUCCESS: Get static VA detail",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticVaDetail",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaDetailRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticVaDetailRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
			},
		},
		{
			name:    "ERROR: Repository error",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticVaDetail",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaDetailRequest")).
					Return((*paymentModel.StaticVaDetailResponse)(nil), constant.ErrSomeErrorForUnitTest)
			},
			request: paymentModel.StaticVaDetailRequest{
				PaymentID:  "invalid-payment",
				MerchantID: "merchant-123",
			},
		},
		{
			name:    "ERROR: Empty payment ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// No mock setup needed as validation happens before repository call
			},
			request: paymentModel.StaticVaDetailRequest{
				PaymentID:  "",
				MerchantID: "merchant-123",
			},
		},
		{
			name:    "ERROR: Empty merchant ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// No mock setup needed as validation happens before repository call
			},
			request: paymentModel.StaticVaDetailRequest{
				PaymentID:  "payment-123",
				MerchantID: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			tc.setupMock(paymentRepo)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			response, err := paymentSvc.GetStaticVaDetail(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.IsType(t, &paymentModel.StaticVaDetailResponse{}, response)
			}

		})
	}
}

func TestGetStaticVaTransactions(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	expectedResponse := &commonModel.PaginationResponse{
		Data: []paymentModel.StaticVaTransactionItem{
			{
				UUID:        "tx-123",
				ReferenceID: "QA202501151023",
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
		request   paymentModel.StaticVaTransactionFilterRequest
	}{
		{
			name:    "SUCCESS: Get static VA transactions",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticVaTransactions",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaTransactionFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "paymentDate",
			},
		},
		{
			name:    "SUCCESS: Get static VA transactions with status filter",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticVaTransactions",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaTransactionFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Status:     "SUCCESS",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "paymentDate",
			},
		},
		{
			name:    "SUCCESS: Get static VA transactions with ID filter",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				repo.On("GetStaticVaTransactions",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaTransactionFilterRequest")).
					Return(expectedResponse, nil)
			},
			request: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				ID:         "tx-123",
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
				repo.On("GetStaticVaTransactions",
					constant.ValueCtxMockType(), mock.AnythingOfType("paymentModel.StaticVaTransactionFilterRequest")).
					Return((*commonModel.PaginationResponse)(nil), constant.ErrSomeErrorForUnitTest)
			},
			request: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "invalid-payment",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
		},
		{
			name:    "ERROR: Empty payment ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// No mock setup needed as validation happens before repository call
			},
			request: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
		},
		{
			name:    "ERROR: Empty merchant ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository) {
				// No mock setup needed as validation happens before repository call
			},
			request: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "",
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
			response, err := paymentSvc.GetStaticVaTransactions(ctx, tc.request)
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

func TestDeactivateStaticVa(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name       string
		wantErr    bool
		setupMock  func(*repositoryMocks.IPaymentRepository, *repositoryMocks.ISnapCoreRepository)
		paymentID  string
		merchantID string
		request    paymentModel.StaticVaUpdateStatusRequest
	}{
		{
			name:    "SUCCESS: Deactivate static VA",
			wantErr: false,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				processorRefNumber := "1234567890"
				metadata := map[string]any{
					"methodDetail": map[string]any{
						"processorReferenceId": "processor-uuid-123",
					},
				}
				// Mock getting current payment status
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{
						Status:                   "ACTIVE",
						ProcessorReferenceNumber: &processorRefNumber,
						Metadata:                 &metadata,
					}, nil)
				// Mock updating payment status
				repo.On("UpdatePaymentStatus",
					constant.ValueCtxMockType(), "payment-123", "merchant-123", "INACTIVE", mock.AnythingOfType("time.Time")).
					Return(nil)
				// Mock snapCore delete VA
				snapCoreRepo.On("DeleteVirtualAccount",
					constant.ValueCtxMockType(), mock.MatchedBy(func(req *snapCoreModel.DeleteVirtualAccountRequest) bool {
						return req.Number == "1234567890" && req.UUID == "processor-uuid-123"
					})).
					Return(nil, nil)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Payment not currently ACTIVE",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				// Mock getting current payment status as already INACTIVE
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{Status: "INACTIVE"}, nil)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Repository error on getting payment",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return((*paymentModel.Payment)(nil), constant.ErrSomeErrorForUnitTest)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Repository error on updating status",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				processorRefNumber := "1234567890"
				metadata := map[string]any{
					"methodDetail": map[string]any{
						"processorReferenceId": "processor-uuid-123",
					},
				}
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{
						Status:                   "ACTIVE",
						ProcessorReferenceNumber: &processorRefNumber,
						Metadata:                 &metadata,
					}, nil)
				// snap core delete VA is called first and should succeed
				snapCoreRepo.On("DeleteVirtualAccount",
					constant.ValueCtxMockType(), mock.MatchedBy(func(req *snapCoreModel.DeleteVirtualAccountRequest) bool {
						return req.Number == "1234567890" && req.UUID == "processor-uuid-123"
					})).
					Return(nil, nil)
				// DB update is called after snap core succeeds, but fails
				repo.On("UpdatePaymentStatus",
					constant.ValueCtxMockType(), "payment-123", "merchant-123", "INACTIVE", mock.AnythingOfType("time.Time")).
					Return(constant.ErrSomeErrorForUnitTest)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Empty payment ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				// No mock setup needed as validation happens before repository call
			},
			paymentID:  "",
			merchantID: "merchant-123",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Empty merchant ID",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				// No mock setup needed as validation happens before repository call
			},
			paymentID:  "payment-123",
			merchantID: "",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: SnapCore delete VA error - DB should NOT be updated",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				processorRefNumber := "1234567890"
				metadata := map[string]any{
					"methodDetail": map[string]any{
						"processorReferenceId": "processor-uuid-123",
					},
				}
				// Mock getting current payment status
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{
						Status:                   "ACTIVE",
						ProcessorReferenceNumber: &processorRefNumber,
						Metadata:                 &metadata,
					}, nil)
				// Mock snapCore delete VA error
				// Note: UpdatePaymentStatus should NOT be called when DeleteVirtualAccount fails
				snapCoreRepo.On("DeleteVirtualAccount",
					constant.ValueCtxMockType(), mock.MatchedBy(func(req *snapCoreModel.DeleteVirtualAccountRequest) bool {
						return req.Number == "1234567890" && req.UUID == "processor-uuid-123"
					})).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
		{
			name:    "ERROR: Missing processor reference ID in metadata - DB should NOT be updated",
			wantErr: true,
			setupMock: func(repo *repositoryMocks.IPaymentRepository, snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				processorRefNumber := "1234567890"
				metadata := map[string]any{
					"methodDetail": map[string]any{
						// Missing processorReferenceId
					},
				}
				// Mock getting current payment status
				repo.On("GetPaymentByIdAndMerchantId",
					constant.ValueCtxMockType(), "payment-123", "merchant-123").
					Return(&paymentModel.Payment{
						Status:                   "ACTIVE",
						ProcessorReferenceNumber: &processorRefNumber,
						Metadata:                 &metadata,
					}, nil)
				// Note: Neither DeleteVirtualAccount nor UpdatePaymentStatus should be called
				// when metadata extraction fails
			},
			paymentID:  "payment-123",
			merchantID: "merchant-123",
			request: paymentModel.StaticVaUpdateStatusRequest{
				Status: "INACTIVE",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			tc.setupMock(paymentRepo, snapCoreRepo)
			paymentSvc := New(paymentRepo, logger, snapCoreRepo, nil, nil, nil, nil)

			ctx := context.Background()
			err := paymentSvc.DeactivateStaticVa(ctx, tc.paymentID, tc.merchantID, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
