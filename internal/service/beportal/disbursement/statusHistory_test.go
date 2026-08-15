package disbursementService

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
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

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	tests := []struct {
		name         string
		request      *statusHistoryModel.RecordDisbursementStatusHistoryRequest
		modifierMock func()
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: Record status with known status",
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				DisbursementID: "test-disbursement-id",
				Status:         constant.DisbursementStatusHistoryWaiting,
				Actor:          "test-user",
				ReasonType:     "",
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
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				DisbursementID: "test-disbursement-id",
				Status:         "UNKNOWN_STATUS",
				Actor:          "test-user",
				ReasonType:     "",
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
			name: "SUCCESS: Record failed status with reason type",
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				DisbursementID: "test-disbursement-id",
				Status:         constant.DisbursementStatusHistoryFailed,
				Actor:          "system",
				ReasonType:     constant.DisbursementReasonTypeInvalidAccount,
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
			name: "SUCCESS: Record processing status with reason type",
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				DisbursementID: "test-disbursement-id",
				Status:         constant.DisbursementStatusHistoryProcessing,
				Actor:          "system",
				ReasonType:     constant.DisbursementReasonTypeDelayed,
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
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				DisbursementID: "test-disbursement-id",
				Status:         constant.DisbursementStatusHistoryWaiting,
				Actor:          "test-user",
				ReasonType:     "",
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
			request: &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
				DisbursementID: "test-disbursement-id",
				Status:         constant.DisbursementStatusHistoryWaiting,
				Actor:          "test-user",
				ReasonType:     "",
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

			svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			concreteSvc := svc.(*DisbursementService)
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

func TestRecordDisbursementWaiting(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	ctx := context.Background()
	disbursementID := "test-disbursement-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			// Verify the status history object has correct values
			assert.Equal(t, constant.TypeDisbursement, sh.ReferenceType)
			assert.Equal(t, disbursementID, sh.ReferenceID)
			assert.Equal(t, constant.DisbursementStatusHistoryWaiting, sh.Status)

			// Verify metadata
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, "Payout Created", metadata.Label)

			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*DisbursementService)
	concreteSvc.recordDisbursementWaiting(ctx, disbursementID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordDisbursementApproved(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	ctx := context.Background()
	disbursementID := "test-disbursement-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.DisbursementStatusHistoryApproved, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*DisbursementService)
	concreteSvc.recordDisbursementApproved(ctx, disbursementID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordDisbursementWaitingForTopUp(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	ctx := context.Background()
	disbursementID := "test-disbursement-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			// Note: Actor should be system type, not the passed actor
			assert.Equal(t, constant.UserSystemType, metadata.Actor)
			assert.Equal(t, constant.DisbursementStatusHistoryWaitingForTopUp, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*DisbursementService)
	concreteSvc.recordDisbursementWaitingForTopUp(ctx, disbursementID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordDisbursementRejected(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	ctx := context.Background()
	disbursementID := "test-disbursement-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.DisbursementStatusHistoryRejected, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*DisbursementService)
	concreteSvc.recordDisbursementRejected(ctx, disbursementID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordDisbursementProcessing(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	ctx := context.Background()
	disbursementID := "test-disbursement-id"
	reasonType := constant.DisbursementReasonTypeDelayed

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, constant.UserSystemType, metadata.Actor)
			assert.Equal(t, constant.DisbursementStatusHistoryProcessing, sh.Status)
			// Should have specific description for delayed reason
			assert.Contains(t, metadata.Description, "longer than expected")
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*DisbursementService)
	concreteSvc.recordDisbursementProcessing(ctx, disbursementID, reasonType)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordDisbursementSuccess(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	ctx := context.Background()
	disbursementID := "test-disbursement-id"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, constant.UserSystemType, metadata.Actor)
			assert.Equal(t, constant.DisbursementStatusHistorySuccess, sh.Status)
			assert.Equal(t, "Success", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*DisbursementService)
	concreteSvc.recordDisbursementSuccess(ctx, disbursementID)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordDisbursementFailed(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}

	ctx := context.Background()
	disbursementID := "test-disbursement-id"
	reasonType := constant.DisbursementReasonTypeInvalidAccount

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, constant.UserSystemType, metadata.Actor)
			assert.Equal(t, constant.DisbursementStatusHistoryFailed, sh.Status)
			// Should have specific description for invalid account reason
			assert.Contains(t, metadata.Description, "account number is invalid")
			assert.Contains(t, metadata.Recommendation, "correct and active")
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*DisbursementService)
	concreteSvc.recordDisbursementFailed(ctx, disbursementID, reasonType)

	statusHistoriesRepo.AssertExpectations(t)
}
