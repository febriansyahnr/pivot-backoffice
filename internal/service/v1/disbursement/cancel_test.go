package disbursementService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCancel(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	merchantID := uuid.NewString()
	bulkID1 := uuid.NewString()
	bulkID2 := uuid.NewString()
	batchID1 := uuid.NewString()
	batchID2 := uuid.NewString()
	batchID3 := uuid.NewString()

	// Helper to create pointer to string
	stringPtr := func(s string) *string { return &s }

	// Mock disbursement data with valid reason types (insufficient balance only)
	mockDisbursementsBulk1 := []*disbursementModel.DisbursementWithTransaction{
		{
			Disbursement: disbursementModel.Disbursement{
				UUID:       "disburse-1-1",
				ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
			},
		},
		{
			Disbursement: disbursementModel.Disbursement{
				UUID:       "disburse-1-2",
				ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
			},
		},
	}

	mockDisbursementsBulk2 := []*disbursementModel.DisbursementWithTransaction{
		{
			Disbursement: disbursementModel.Disbursement{
				UUID:       "disburse-2-1",
				ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
			},
		},
	}

	mockDisbursementsIndividualValid := []*disbursementModel.Disbursement{
		{
			UUID:       batchID1,
			ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
		},
		{
			UUID:       batchID2,
			ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository)
		input     *disbursementModel.CancelPayoutRequest
	}{
		{
			name:    "ERROR: Empty payload - no BatchBulkID and no BatchID",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				// No mock setup needed since validation happens before repository calls
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID: merchantID,
			},
		},
		{
			name:    "ERROR: Empty payload - empty arrays for both BatchBulkID and BatchID",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				// No mock setup needed since validation happens before repository calls
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{},
				BatchID:     []string{},
			},
		},
		{
			name:    "ERROR: GetAllDisbursementByBulkID fails for first bulk ID",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					constant.ValueCtxMockType(),
					bulkID1,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{bulkID1, bulkID2},
				BatchID:     []string{},
			},
		},
		{
			name:    "ERROR: GetByIDs fails for individual batch IDs",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetByIDs",
					constant.ValueCtxMockType(),
					[]string{batchID1, batchID2},
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{},
				BatchID:     []string{batchID1, batchID2},
			},
		},
		{
			name:    "ERROR: UpdateReasonByIDs fails",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetByIDs",
					constant.ValueCtxMockType(),
					[]string{batchID1, batchID2},
				).Return(mockDisbursementsIndividualValid, nil) // Return only first 2 with valid reason type

				expectedBatchIDs := []string{batchID1, batchID2}
				disbursementRepo.On(
					"UpdateReasonByIDs",
					constant.ValueCtxMockType(),
					expectedBatchIDs,
					constant.DisbursementReasonTypeCancelled,
					"The resource was cancelled by users",
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{},
				BatchID:     []string{batchID1, batchID2},
			},
		},
		{
			name:    "ERROR: Bulk disbursement has invalid reason type (cannot be cancelled)",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				mockDisbursementsWithInvalidReason := []*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       "disburse-invalid-1",
							ReasonType: stringPtr("INVALID_REASON_TYPE"), // Not insufficient balance
						},
					},
				}

				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					constant.ValueCtxMockType(),
					bulkID1,
				).Return(mockDisbursementsWithInvalidReason, nil)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{bulkID1},
				BatchID:     []string{},
			},
		},
		{
			name:    "ERROR: UpdateBulkDisbursementStatusByID fails for first bulk ID",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					constant.ValueCtxMockType(),
					bulkID1,
				).Return(mockDisbursementsBulk1, nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					constant.ValueCtxMockType(),
					bulkID1,
					constant.BulkDisbursementStatusDone,
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{bulkID1},
				BatchID:     []string{},
			},
		},
		{
			name:    "SUCCESS: Cancel with BatchBulkID only - filters reason types",
			wantErr: false,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					constant.ValueCtxMockType(),
					bulkID1,
				).Return(mockDisbursementsBulk1, nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					constant.ValueCtxMockType(),
					bulkID1,
					constant.BulkDisbursementStatusDone,
				).Return(nil)

				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					constant.ValueCtxMockType(),
					bulkID2,
				).Return(mockDisbursementsBulk2, nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					constant.ValueCtxMockType(),
					bulkID2,
					constant.BulkDisbursementStatusDone,
				).Return(nil)

				// Mock GetByIDs call for empty BatchID array
				disbursementRepo.On(
					"GetByIDs",
					constant.ValueCtxMockType(),
					[]string{},
				).Return([]*disbursementModel.Disbursement{}, nil)

				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

				// Only expect valid disbursements (filtered by reason type)
				expectedBatchIDs := []string{"disburse-1-1", "disburse-1-2", "disburse-2-1"}
				disbursementRepo.On(
					"UpdateReasonByIDs",
					constant.ValueCtxMockType(),
					expectedBatchIDs,
					constant.DisbursementReasonTypeCancelled,
					"The resource was cancelled by users",
				).Return(nil)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{bulkID1, bulkID2},
				BatchID:     []string{},
			},
		},
		{
			name:    "SUCCESS: Cancel with BatchID only - filters reason types",
			wantErr: false,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetByIDs",
					constant.ValueCtxMockType(),
					[]string{batchID1, batchID2, batchID3},
				).Return(mockDisbursementsIndividualValid, nil)

				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

				// Only expect valid disbursements (filtered by reason type)
				expectedBatchIDs := []string{batchID1, batchID2}
				disbursementRepo.On(
					"UpdateReasonByIDs",
					constant.ValueCtxMockType(),
					expectedBatchIDs,
					constant.DisbursementReasonTypeCancelled,
					"The resource was cancelled by users",
				).Return(nil)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{},
				BatchID:     []string{batchID1, batchID2, batchID3},
			},
		},
		{
			name:    "SUCCESS: Cancel with both BatchBulkID and BatchID",
			wantErr: false,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					constant.ValueCtxMockType(),
					bulkID1,
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:       "disburse-bulk-1",
							ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
						},
					},
				}, nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					constant.ValueCtxMockType(),
					bulkID1,
					constant.BulkDisbursementStatusDone,
				).Return(nil)

				disbursementRepo.On(
					"GetByIDs",
					constant.ValueCtxMockType(),
					[]string{batchID1},
				).Return([]*disbursementModel.Disbursement{
					{
						UUID:       batchID1,
						ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
					},
				}, nil)

				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

				expectedBatchIDs := []string{"disburse-bulk-1", batchID1}
				disbursementRepo.On(
					"UpdateReasonByIDs",
					constant.ValueCtxMockType(),
					expectedBatchIDs,
					constant.DisbursementReasonTypeCancelled,
					"The resource was cancelled by users",
				).Return(nil)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{bulkID1},
				BatchID:     []string{batchID1},
			},
		},
		{
			name:    "Error: No valid disbursements to cancel (all filtered out)",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetByIDs",
					constant.ValueCtxMockType(),
					[]string{batchID1},
				).Return([]*disbursementModel.Disbursement{
					{
						UUID:       batchID1,
						ReasonType: stringPtr("OTHER_REASON"), // Will be filtered out
					},
				}, nil)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{},
				BatchID:     []string{batchID1},
			},
		},
		{
			name:    "Error: when the reason type is nil then should return error",
			wantErr: true,
			mockSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, statusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				disbursementRepo.On(
					"GetByIDs",
					constant.ValueCtxMockType(),
					[]string{batchID1, batchID2},
				).Return([]*disbursementModel.Disbursement{
					{
						UUID:       batchID1,
						ReasonType: nil, // Will be filtered out
					},
					{
						UUID:       batchID2,
						ReasonType: stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
					},
				}, nil)
			},
			input: &disbursementModel.CancelPayoutRequest{
				MerchantID:  merchantID,
				BatchBulkID: []string{},
				BatchID:     []string{batchID1, batchID2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			tc.mockSetup(mockRepo, statusHistoriesRepo)

			svc := New(&conf, pdkLoggerMock, nil, mockRepo, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			cancelledPayouts, err := svc.Cancel(context.Background(), tc.input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, cancelledPayouts)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, cancelledPayouts)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
