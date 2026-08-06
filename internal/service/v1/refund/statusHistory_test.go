package refundService

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRecordStatusHistory(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	tests := []struct {
		name         string
		request      *statusHistoryModel.RecordRefundStatusHistoryRequest
		modifierMock func()
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: Record status with known status",
			request: &statusHistoryModel.RecordRefundStatusHistoryRequest{
				RefundID: "test-refund-id",
				Status:   constant.RefundStatusHistoryPending,
				Actor:    "test-user",
			},
			modifierMock: func() {
				statusHistoriesRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Record status with unknown status",
			request: &statusHistoryModel.RecordRefundStatusHistoryRequest{
				RefundID: "test-refund-id",
				Status:   "UNKNOWN_STATUS",
				Actor:    "test-user",
			},
			modifierMock: func() {
				statusHistoriesRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Record success status",
			request: &statusHistoryModel.RecordRefundStatusHistoryRequest{
				RefundID: "test-refund-id",
				Status:   constant.RefundStatusHistorySuccess,
				Actor:    constant.StatusHistoryActorSystem,
			},
			modifierMock: func() {
				statusHistoriesRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Record waiting bank transfer status",
			request: &statusHistoryModel.RecordRefundStatusHistoryRequest{
				RefundID: "test-refund-id",
				Status:   constant.RefundStatusHistoryWaitingBankTransfer,
				Actor:    constant.StatusHistoryActorSystem,
			},
			modifierMock: func() {
				statusHistoriesRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: JSON marshal success - basic structs should always marshal",
			request: &statusHistoryModel.RecordRefundStatusHistoryRequest{
				RefundID: "test-refund-id",
				Status:   constant.RefundStatusHistoryPending,
				Actor:    constant.StatusHistoryActorUser,
			},
			modifierMock: func() {
				statusHistoriesRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Repository insert error - should be logged but not returned",
			request: &statusHistoryModel.RecordRefundStatusHistoryRequest{
				RefundID: "test-refund-id",
				Status:   constant.RefundStatusHistoryPending,
				Actor:    constant.StatusHistoryActorUser,
			},
			modifierMock: func() {
				statusHistoriesRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: false, // Repository error is logged but function still returns nil
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.modifierMock()

			svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			concreteSvc := svc.(*RefundService)
			err := concreteSvc.recordStatusHistory(ctx, tc.request)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrMsg != "" {
					assert.ErrorContains(t, err, tc.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecordRefundPending(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	refundID := "test-refund-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			// Verify the status history object has correct values
			assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
			assert.Equal(t, refundID, sh.ReferenceID)
			assert.Equal(t, constant.RefundStatusHistoryPending, sh.Status)

			// Verify metadata
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, "Refund Created", metadata.Label)

			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*RefundService)
	concreteSvc.recordRefundPending(ctx, refundID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordRefundWaitingBankTransfer(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	refundID := "test-refund-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.RefundStatusHistoryWaitingBankTransfer, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*RefundService)
	concreteSvc.recordRefundWaitingBankTransfer(ctx, refundID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordRefundSuccess(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	refundID := "test-refund-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.RefundStatusHistorySuccess, sh.Status)
			assert.Equal(t, "Success", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*RefundService)
	concreteSvc.recordRefundSuccess(ctx, refundID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordRefundFailed(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	refundID := "test-refund-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.RefundStatusHistoryFailed, sh.Status)
			assert.Equal(t, "Failed", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*RefundService)
	concreteSvc.recordRefundFailed(ctx, refundID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordRefundCancelled(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	refundID := "test-refund-id"
	actor := "user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.RefundStatusHistoryCancelled, sh.Status)
			assert.Equal(t, "Cancelled", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*RefundService)
	concreteSvc.recordRefundCancelled(ctx, refundID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordRefundStatusHistory(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	tests := []struct {
		name         string
		refundID     string
		actor        string
		statusType   string
		expectCall   bool
		modifierMock func(repo *repositoryMocks.IStatusHistoriesRepository)
	}{
		{
			name:       "SUCCESS: Record pending status",
			refundID:   "test-refund-id",
			actor:      constant.StatusHistoryActorSystem,
			statusType: constant.RefundStatusHistoryPending,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
						assert.Equal(t, "test-refund-id", sh.ReferenceID)
						assert.Equal(t, constant.RefundStatusHistoryPending, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record waiting bank transfer status",
			refundID:   "test-refund-id",
			actor:      constant.StatusHistoryActorSystem,
			statusType: constant.RefundStatusHistoryWaitingBankTransfer,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
						assert.Equal(t, "test-refund-id", sh.ReferenceID)
						assert.Equal(t, constant.RefundStatusHistoryWaitingBankTransfer, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record success status",
			refundID:   "test-refund-id",
			actor:      constant.StatusHistoryActorProcessor,
			statusType: constant.RefundStatusHistorySuccess,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
						assert.Equal(t, "test-refund-id", sh.ReferenceID)
						assert.Equal(t, constant.RefundStatusHistorySuccess, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record failed status",
			refundID:   "test-refund-id",
			actor:      constant.StatusHistoryActorProcessor,
			statusType: constant.RefundStatusHistoryFailed,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
						assert.Equal(t, "test-refund-id", sh.ReferenceID)
						assert.Equal(t, constant.RefundStatusHistoryFailed, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record cancelled status",
			refundID:   "test-refund-id",
			actor:      constant.StatusHistoryActorUser,
			statusType: constant.RefundStatusHistoryCancelled,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
						assert.Equal(t, "test-refund-id", sh.ReferenceID)
						assert.Equal(t, constant.RefundStatusHistoryCancelled, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Unknown status type - handled by generic method",
			refundID:   "test-refund-id",
			actor:      "test-user",
			statusType: "UNKNOWN_STATUS",
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
						assert.Equal(t, "test-refund-id", sh.ReferenceID)
						assert.Equal(t, "UNKNOWN_STATUS", sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Empty status type - handled by generic method",
			refundID:   "test-refund-id",
			actor:      "test-user",
			statusType: "",
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypeRefund, sh.ReferenceType)
						assert.Equal(t, "test-refund-id", sh.ReferenceID)
						assert.Equal(t, "", sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a fresh mock repository for each test case
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			tc.modifierMock(statusHistoriesRepo)

			svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			concreteSvc := svc.(*RefundService)
			concreteSvc.RecordRefundStatusHistory(ctx, tc.refundID, tc.actor, tc.statusType)

			if tc.expectCall {
				statusHistoriesRepo.AssertExpectations(t)
			}
		})
	}
}

func TestRecordRefundStatusHistory_WithNilRepository(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	refundID := "test-refund-id"
	actor := "test-user"
	statusType := constant.RefundStatusHistoryPending

	// Create service without status history repository (nil)
	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, nil)
	concreteSvc := svc.(*RefundService)

	// Should not panic and should return immediately
	concreteSvc.RecordRefundStatusHistory(ctx, refundID, actor, statusType)
}