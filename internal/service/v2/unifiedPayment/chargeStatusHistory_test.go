package unifiedPaymentService

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
	}

	tests := []struct {
		name         string
		request      *statusHistoryModel.RecordChargeStatusHistoryRequest
		modifierMock func()
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: Record status with known status",
			request: &statusHistoryModel.RecordChargeStatusHistoryRequest{
				ChargeID: "test-charge-id",
				Status:   constant.ChargeStatusHistoryWaitingForUserAction,
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
			request: &statusHistoryModel.RecordChargeStatusHistoryRequest{
				ChargeID: "test-charge-id",
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
			request: &statusHistoryModel.RecordChargeStatusHistoryRequest{
				ChargeID: "test-charge-id",
				Status:   constant.ChargeStatusHistorySuccess,
				Actor:    constant.StatusHistoryActorProcessor,
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
			name: "SUCCESS: Record processing status",
			request: &statusHistoryModel.RecordChargeStatusHistoryRequest{
				ChargeID: "test-charge-id",
				Status:   constant.ChargeStatusHistoryProcessing,
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
			request: &statusHistoryModel.RecordChargeStatusHistoryRequest{
				ChargeID: "test-charge-id",
				Status:   constant.ChargeStatusHistoryWaitingForUserAction,
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
			request: &statusHistoryModel.RecordChargeStatusHistoryRequest{
				ChargeID: "test-charge-id",
				Status:   constant.ChargeStatusHistoryWaitingForUserAction,
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

			svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			concreteSvc := svc.(*UnifiedPaymentService)
			err := concreteSvc.recordChargeStatusHistory(ctx, tc.request)

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

func TestRecordChargeWaitingForUserAction(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			// Verify the status history object has correct values
			assert.Equal(t, constant.TypePayment, sh.ReferenceType)
			assert.Equal(t, chargeID, sh.ReferenceID)
			assert.Equal(t, constant.ChargeStatusHistoryWaitingForUserAction, sh.Status)

			// Verify metadata
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, "Waiting for User Action", metadata.Label)

			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeWaitingForUserAction(ctx, chargeID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeWaitingForAuthentication(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.ChargeStatusHistoryWaitingForAuthentication, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeWaitingForAuthentication(ctx, chargeID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeWaitingForExternalFDS(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.ChargeStatusHistoryWaitingForExternalFDS, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeWaitingForExternalFDS(ctx, chargeID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeProcessing(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.ChargeStatusHistoryProcessing, sh.Status)
			assert.Equal(t, "Processing", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeProcessing(ctx, chargeID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeWaitingForCapture(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.ChargeStatusHistoryWaitingForCapture, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeWaitingForCapture(ctx, chargeID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeSuccess(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, constant.StatusHistoryActorProcessor, metadata.Actor)
			assert.Equal(t, constant.ChargeStatusHistorySuccess, sh.Status)
			assert.Equal(t, "Success", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeSuccess(ctx, chargeID, constant.StatusHistoryActorProcessor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeFailed(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, constant.StatusHistoryActorProcessor, metadata.Actor)
			assert.Equal(t, constant.ChargeStatusHistoryFailed, sh.Status)
			assert.Equal(t, "Failed", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeFailed(ctx, chargeID, constant.StatusHistoryActorProcessor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeExpired(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, constant.StatusHistoryActorSystem, metadata.Actor)
			assert.Equal(t, constant.ChargeStatusHistoryExpired, sh.Status)
			assert.Equal(t, "Expired", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	concreteSvc := svc.(*UnifiedPaymentService)
	concreteSvc.recordChargeExpired(ctx, chargeID, constant.StatusHistoryActorSystem)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordChargeStatusHistory(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	tests := []struct {
		name         string
		chargeID     string
		actor        string
		statusType   string
		expectCall   bool
		modifierMock func(repo *repositoryMocks.IStatusHistoriesRepository)
	}{
		{
			name:       "SUCCESS: Record waiting for user action status",
			chargeID:   "test-charge-id",
			actor:      "test-user",
			statusType: constant.ChargeStatusHistoryWaitingForUserAction,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistoryWaitingForUserAction, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record waiting for authentication status",
			chargeID:   "test-charge-id",
			actor:      "test-user",
			statusType: constant.ChargeStatusHistoryWaitingForAuthentication,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistoryWaitingForAuthentication, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record waiting for external FDS status",
			chargeID:   "test-charge-id",
			actor:      "test-user",
			statusType: constant.ChargeStatusHistoryWaitingForExternalFDS,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistoryWaitingForExternalFDS, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record processing status",
			chargeID:   "test-charge-id",
			actor:      constant.StatusHistoryActorSystem,
			statusType: constant.ChargeStatusHistoryProcessing,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistoryProcessing, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record waiting for capture status",
			chargeID:   "test-charge-id",
			actor:      constant.StatusHistoryActorUser,
			statusType: constant.ChargeStatusHistoryWaitingForCapture,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistoryWaitingForCapture, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record success status",
			chargeID:   "test-charge-id",
			actor:      constant.StatusHistoryActorProcessor,
			statusType: constant.ChargeStatusHistorySuccess,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistorySuccess, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record failed status",
			chargeID:   "test-charge-id",
			actor:      constant.StatusHistoryActorProcessor,
			statusType: constant.ChargeStatusHistoryFailed,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistoryFailed, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Record expired status",
			chargeID:   "test-charge-id",
			actor:      constant.StatusHistoryActorSystem,
			statusType: constant.ChargeStatusHistoryExpired,
			expectCall: true,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				repo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
						assert.Equal(t, constant.TypePayment, sh.ReferenceType)
						assert.Equal(t, "test-charge-id", sh.ReferenceID)
						assert.Equal(t, constant.ChargeStatusHistoryExpired, sh.Status)
						return true
					}),
				).Return(nil).Once()
			},
		},
		{
			name:       "SUCCESS: Unknown status type - no method called",
			chargeID:   "test-charge-id",
			actor:      "test-user",
			statusType: "UNKNOWN_STATUS",
			expectCall: false,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				// No mock expectations since no method should be called
			},
		},
		{
			name:       "SUCCESS: Empty status type - no method called",
			chargeID:   "test-charge-id",
			actor:      "test-user",
			statusType: "",
			expectCall: false,
			modifierMock: func(repo *repositoryMocks.IStatusHistoriesRepository) {
				// No mock expectations since no method should be called
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a fresh mock repository for each test case
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			tc.modifierMock(statusHistoriesRepo)

			svc := New(conf, mockLogger, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			concreteSvc := svc.(*UnifiedPaymentService)
			concreteSvc.RecordChargeStatusHistory(ctx, tc.chargeID, tc.actor, tc.statusType)

			if tc.expectCall {
				statusHistoriesRepo.AssertExpectations(t)
			}
		})
	}
}

func TestRecordChargeStatusHistory_WithNilRepository(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctx := context.Background()
	chargeID := "test-charge-id"
	actor := "test-user"
	statusType := constant.ChargeStatusHistoryWaitingForUserAction

	// Create service without status history repository (nil)
	svc := New(conf, mockLogger, nil, nil, nil)
	concreteSvc := svc.(*UnifiedPaymentService)

	// Should not panic and should return immediately
	concreteSvc.RecordChargeStatusHistory(ctx, chargeID, actor, statusType)
}