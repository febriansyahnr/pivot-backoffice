package refundService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRefundDetail(t *testing.T) {
	testCases := []struct {
		name          string
		request       refundModel.FilterRefundRequest
		mockRepoResp  *commonModel.PaginationResponse
		mockRepoErr   error
		expectedResp  *refundModel.RefundResponse
		expectedError error
	}{
		{
			name: "Success case",
			request: refundModel.FilterRefundRequest{
				UUID: "test-uuid",
			},
			mockRepoResp: &commonModel.PaginationResponse{
				Data: []*refundModel.RefundResponse{
					{ID: "test-uuid", Amount: commonModel.Amount{Value: "1000"}},
				},
			},
			mockRepoErr:   nil,
			expectedResp:  &refundModel.RefundResponse{ID: "test-uuid", Amount: commonModel.Amount{Value: "1000"}},
			expectedError: nil,
		},
		{
			name: "Database error",
			request: refundModel.FilterRefundRequest{
				UUID: "test-uuid",
			},
			mockRepoResp:  nil,
			mockRepoErr:   errors.New("database error"),
			expectedResp:  nil,
			expectedError: pkgErr.New(httpResponse.HttpErrDatabase, errors.New("database error")),
		},
		{
			name: "Refund not found",
			request: refundModel.FilterRefundRequest{
				UUID: "test-uuid",
			},
			mockRepoResp: &commonModel.PaginationResponse{
				Data: []*refundModel.RefundResponse{},
			},
			mockRepoErr:   nil,
			expectedResp:  nil,
			expectedError: pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrRefundNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock dependencies
			mockRepo := repositoryMocks.NewIRefundRepository(t)
			_, mockLogger, _ := test.SetupLogger()

			// Setup expectations
			mockRepo.On("GetRefundList", mock.Anything, tc.request).Return(tc.mockRepoResp, tc.mockRepoErr)

			// Create service with mocked dependencies
			svc := &RefundService{
				refundRepo: mockRepo,
				logger:     mockLogger,
			}

			// Call function
			result, err := svc.GetRefundDetail(context.Background(), tc.request)

			// Assert results
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				mockRepo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResp, result)

			// Verify all expectations
			mockRepo.AssertExpectations(t)
		})
	}
}
func TestGetRefundList(t *testing.T) {
	testCases := []struct {
		name          string
		request       refundModel.FilterRefundRequest
		mockRepoResp  *commonModel.PaginationResponse
		mockRepoErr   error
		expectedResp  *commonModel.PaginationResponse
		expectedError error
	}{
		{
			name: "Success case",
			request: refundModel.FilterRefundRequest{
				Page:    1,
				PerPage: 10,
				SortBy:  "created_at",
				Sort:    "desc",
			},
			mockRepoResp: &commonModel.PaginationResponse{
				Data: []*refundModel.RefundResponse{
					{ID: "refund-1", Amount: commonModel.Amount{Value: "1000"}},
					{ID: "refund-2", Amount: commonModel.Amount{Value: "2000"}},
				},
				Meta: commonModel.Meta{
					Page:       1,
					TotalItems: 2,
					TotalPages: 1,
					PerPage:    10,
				},
			},
			mockRepoErr: nil,
			expectedResp: &commonModel.PaginationResponse{
				Data: []*refundModel.RefundResponse{
					{ID: "refund-1", Amount: commonModel.Amount{Value: "1000"}},
					{ID: "refund-2", Amount: commonModel.Amount{Value: "2000"}},
				},
				Meta: commonModel.Meta{
					Page:       1,
					TotalItems: 2,
					TotalPages: 1,
					PerPage:    10,
				},
			},
			expectedError: nil,
		},
		{
			name: "Empty result case",
			request: refundModel.FilterRefundRequest{
				Page:    1,
				PerPage: 10,
				SortBy:  "created_at",
				Sort:    "desc",
			},
			mockRepoResp: &commonModel.PaginationResponse{
				Data: []*refundModel.RefundResponse{},
				Meta: commonModel.Meta{
					Page:       1,
					TotalItems: 0,
					PerPage:    10,
					TotalPages: 0,
				},
			},
			mockRepoErr: nil,
			expectedResp: &commonModel.PaginationResponse{
				Data: []*refundModel.RefundResponse{},
				Meta: commonModel.Meta{
					Page:       1,
					TotalItems: 0,
					PerPage:    10,
					TotalPages: 0,
				},
			},
			expectedError: nil,
		},
		{
			name: "Database error",
			request: refundModel.FilterRefundRequest{
				Page:    1,
				PerPage: 10,
				SortBy:  "created_at",
				Sort:    "desc",
			},
			mockRepoResp:  nil,
			mockRepoErr:   errors.New("database error"),
			expectedResp:  nil,
			expectedError: pkgErr.New(httpResponse.HttpErrDatabase, errors.New("database error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock dependencies
			mockRepo := repositoryMocks.NewIRefundRepository(t)
			_, mockLogger, _ := test.SetupLogger()

			// Setup expectations
			mockRepo.On("GetRefundList", mock.Anything, tc.request).Return(tc.mockRepoResp, tc.mockRepoErr)

			// Create service with mocked dependencies
			svc := &RefundService{
				refundRepo: mockRepo,
				logger:     mockLogger,
			}

			// Call function
			result, err := svc.GetRefundList(context.Background(), tc.request)

			// Assert results
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				mockRepo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResp, result)

			// Verify all expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetExistingRefundList(t *testing.T) {
	testCases := []struct {
		name          string
		request       refundModel.GetExistingRefundListRequest
		setupMock     func(*repositoryMocks.IRefundRepository)
		expectedResp  []refundModel.RefundResponse
		expectedError error
	}{
		{
			name: "SUCCESS",
			request: refundModel.GetExistingRefundListRequest{
				PaymentID: "payment-1",
				Status:    "SUCCESS",
			},
			setupMock: func(mockRepo *repositoryMocks.IRefundRepository) {
				mockRepo.On("ListByPaymentID", mock.Anything, "payment-1", mock.MatchedBy(func(req refundModel.ListByPaymentIDRequest) bool {
					return req.Status == "SUCCESS"
				})).Return([]refundModel.RefundResponse{
					{ID: "refund-1", Amount: commonModel.Amount{Value: "1000"}},
					{ID: "refund-2", Amount: commonModel.Amount{Value: "2000"}},
				}, nil)
			},
			expectedResp: []refundModel.RefundResponse{
				{ID: "refund-1", Amount: commonModel.Amount{Value: "1000"}},
				{ID: "refund-2", Amount: commonModel.Amount{Value: "2000"}},
			},
			expectedError: nil,
		},
		{
			name: "ERROR",
			request: refundModel.GetExistingRefundListRequest{
				PaymentID: "payment-1",
				Status:    "SUCCESS",
			},
			setupMock: func(mockRepo *repositoryMocks.IRefundRepository) {
				mockRepo.On("ListByPaymentID", mock.Anything, "payment-1", mock.MatchedBy(func(req refundModel.ListByPaymentIDRequest) bool {
					return req.Status == "SUCCESS"
				})).Return(nil, assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIRefundRepository(t)
			_, mockLogger, _ := test.SetupLogger()

			tc.setupMock(mockRepo)

			svc := &RefundService{
				refundRepo: mockRepo,
				logger:     mockLogger,
			}

			result, err := svc.GetExistingRefundList(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				mockRepo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResp, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFindByID(t *testing.T) {
	testCases := []struct {
		name          string
		refundID      string
		mockRepoResp  *refundModel.Refund
		mockRepoErr   error
		expectedResp  *refundModel.Refund
		expectedError error
	}{
		{
			name:          "Success case",
			refundID:      "test-uuid",
			mockRepoResp:  &refundModel.Refund{UUID: "test-uuid", Amount: 1000},
			mockRepoErr:   nil,
			expectedResp:  &refundModel.Refund{UUID: "test-uuid", Amount: 1000},
			expectedError: nil,
		},
		{
			name:          "Database error",
			refundID:      "test-uuid",
			mockRepoResp:  nil,
			mockRepoErr:   errors.New("database error"),
			expectedResp:  nil,
			expectedError: pkgErr.New(httpResponse.HttpErrDatabase, errors.New("database error")),
		},
		{
			name:          "Refund not found",
			refundID:      "test-uuid",
			mockRepoResp:  nil,
			mockRepoErr:   nil,
			expectedResp:  nil,
			expectedError: pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrRefundNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock dependencies
			mockRepo := repositoryMocks.NewIRefundRepository(t)
			_, mockLogger, _ := test.SetupLogger()

			// Setup expectations
			mockRepo.On("FindByID", mock.Anything, tc.refundID).Return(tc.mockRepoResp, tc.mockRepoErr)

			// Create service with mocked dependencies
			svc := &RefundService{
				refundRepo: mockRepo,
				logger:     mockLogger,
			}

			// Call function
			result, err := svc.FindByID(context.Background(), tc.refundID)

			// Assert results
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				mockRepo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResp, result)

			// Verify all expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetRefundDetailWithStatusHistories(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	testCases := []struct {
		name                    string
		request                 refundModel.FilterRefundRequest
		mockRefundRepoResp      *refundModel.RefundResponse
		mockRefundRepoErr       error
		mockStatusHistoryResp   []*statusHistoryModel.StatusHistory
		mockStatusHistoryErr    error
		expectedResp            *refundModel.RefundResponse
		expectedError           error
		expectStatusHistoryCall bool
	}{
		{
			name: "Success with status history",
			request: refundModel.FilterRefundRequest{
				UUID:       "test-uuid",
				MerchantID: "merchant-1",
			},
			mockRefundRepoResp: &refundModel.RefundResponse{
				ID:               "test-uuid",
				PaymentSessionID: "payment-123",
				Amount:           commonModel.Amount{Value: "50000", Currency: "IDR"},
				Status:           constant.RefundStatusSuccess,
			},
			mockRefundRepoErr: nil,
			mockStatusHistoryResp: []*statusHistoryModel.StatusHistory{
				{
					ID:            "history-1",
					ReferenceType: constant.TypeRefund,
					ReferenceID:   "test-uuid",
					Status:        constant.RefundStatusPending,
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label:       "Refund Created",
						Description: "Refund has been created",
					},
					CreatedAt: fixedTime,
				},
				{
					ID:            "history-2",
					ReferenceType: constant.TypeRefund,
					ReferenceID:   "test-uuid",
					Status:        constant.RefundStatusSuccess,
					MetadataObj: &statusHistoryModel.StatusHistoryMetadata{
						Label:       "Success",
						Description: "Refund completed successfully",
					},
					CreatedAt: fixedTime,
				},
			},
			mockStatusHistoryErr:    nil,
			expectStatusHistoryCall: true,
			expectedResp: &refundModel.RefundResponse{
				ID:               "test-uuid",
				PaymentSessionID: "payment-123",
				Amount:           commonModel.Amount{Value: "50000", Currency: "IDR"},
				Status:           constant.RefundStatusSuccess,
			},
			expectedError: nil,
		},
		{
			name: "Success without status history (empty)",
			request: refundModel.FilterRefundRequest{
				UUID:       "test-uuid",
				MerchantID: "merchant-1",
			},
			mockRefundRepoResp: &refundModel.RefundResponse{
				ID:               "test-uuid",
				PaymentSessionID: "payment-123",
				Amount:           commonModel.Amount{Value: "50000", Currency: "IDR"},
				Status:           constant.RefundStatusPending,
			},
			mockRefundRepoErr:       nil,
			mockStatusHistoryResp:   []*statusHistoryModel.StatusHistory{},
			mockStatusHistoryErr:    nil,
			expectStatusHistoryCall: true,
			expectedResp: &refundModel.RefundResponse{
				ID:               "test-uuid",
				PaymentSessionID: "payment-123",
				Amount:           commonModel.Amount{Value: "50000", Currency: "IDR"},
				Status:           constant.RefundStatusPending,
			},
			expectedError: nil,
		},
		{
			name: "Success with status history fetch error (graceful handling)",
			request: refundModel.FilterRefundRequest{
				UUID:       "test-uuid",
				MerchantID: "merchant-1",
			},
			mockRefundRepoResp: &refundModel.RefundResponse{
				ID:               "test-uuid",
				PaymentSessionID: "payment-123",
				Amount:           commonModel.Amount{Value: "50000", Currency: "IDR"},
				Status:           constant.RefundStatusSuccess,
			},
			mockRefundRepoErr:       nil,
			mockStatusHistoryResp:   nil,
			mockStatusHistoryErr:    errors.New("status history fetch error"),
			expectStatusHistoryCall: true,
			expectedResp: &refundModel.RefundResponse{
				ID:               "test-uuid",
				PaymentSessionID: "payment-123",
				Amount:           commonModel.Amount{Value: "50000", Currency: "IDR"},
				Status:           constant.RefundStatusSuccess,
			},
			expectedError: nil,
		},
		{
			name: "Error refund not found",
			request: refundModel.FilterRefundRequest{
				UUID:       "non-existent-uuid",
				MerchantID: "merchant-1",
			},
			mockRefundRepoResp:      nil,
			mockRefundRepoErr:       nil,
			expectStatusHistoryCall: false,
			expectedResp:            nil,
			expectedError:           pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrRefundNotFound),
		},
		{
			name: "Error database error from refund repo",
			request: refundModel.FilterRefundRequest{
				UUID:       "test-uuid",
				MerchantID: "merchant-1",
			},
			mockRefundRepoResp:      nil,
			mockRefundRepoErr:       errors.New("database error"),
			expectStatusHistoryCall: false,
			expectedResp:            nil,
			expectedError:           pkgErr.New(httpResponse.HttpErrDatabase, errors.New("database error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock dependencies
			mockRefundRepo := repositoryMocks.NewIRefundRepository(t)
			mockStatusHistoryRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			_, mockLogger, _ := test.SetupLogger()

			// Setup refund repo expectations - using GetRefundByID
			mockRefundRepo.On("GetRefundByID", mock.Anything, tc.request.UUID, tc.request.MerchantID).Return(tc.mockRefundRepoResp, tc.mockRefundRepoErr)

			// Setup status history repo expectations (only if expected)
			if tc.expectStatusHistoryCall && tc.mockRefundRepoResp != nil {
				mockStatusHistoryRepo.On("GetByReference", mock.Anything, constant.TypeRefund, tc.mockRefundRepoResp.ID).
					Return(tc.mockStatusHistoryResp, tc.mockStatusHistoryErr)
			}

			// Create service with mocked dependencies
			svc := &RefundService{
				refundRepo:          mockRefundRepo,
				statusHistoriesRepo: mockStatusHistoryRepo,
				logger:              mockLogger,
			}

			// Call function
			result, err := svc.GetRefundDetailWithStatusHistories(context.Background(), tc.request)

			// Assert results
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				mockRefundRepo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.expectedResp.ID, result.ID)
			assert.Equal(t, tc.expectedResp.PaymentSessionID, result.PaymentSessionID)
			assert.Equal(t, tc.expectedResp.Status, result.Status)

			// Verify status history if expected
			if tc.expectStatusHistoryCall && tc.mockStatusHistoryErr == nil && len(tc.mockStatusHistoryResp) > 0 {
				assert.NotEmpty(t, result.StatusHistory)
				assert.Len(t, result.StatusHistory, len(tc.mockStatusHistoryResp))
			}

			// Verify all expectations
			mockRefundRepo.AssertExpectations(t)
			if tc.expectStatusHistoryCall {
				mockStatusHistoryRepo.AssertExpectations(t)
			}
		})
	}
}
