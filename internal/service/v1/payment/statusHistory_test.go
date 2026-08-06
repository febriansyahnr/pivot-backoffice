package paymentService

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
		request      *statusHistoryModel.RecordPaymentStatusHistoryRequest
		modifierMock func()
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: Record status with known status",
			request: &statusHistoryModel.RecordPaymentStatusHistoryRequest{
				PaymentID:  "test-payment-id",
				Status:     constant.PaymentStatusHistoryPending,
				Actor:      "test-user",
				ReasonType: "",
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
			request: &statusHistoryModel.RecordPaymentStatusHistoryRequest{
				PaymentID:  "test-payment-id",
				Status:     "UNKNOWN_STATUS",
				Actor:      "test-user",
				ReasonType: "",
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
			request: &statusHistoryModel.RecordPaymentStatusHistoryRequest{
				PaymentID:  "test-payment-id",
				Status:     constant.PaymentStatusHistorySuccess,
				Actor:      constant.StatusHistoryActorSystem,
				ReasonType: "",
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
			request: &statusHistoryModel.RecordPaymentStatusHistoryRequest{
				PaymentID:  "test-payment-id",
				Status:     constant.PaymentStatusHistoryProcessing,
				Actor:      constant.StatusHistoryActorSystem,
				ReasonType: "",
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
			request: &statusHistoryModel.RecordPaymentStatusHistoryRequest{
				PaymentID:  "test-payment-id",
				Status:     constant.PaymentStatusHistoryPending,
				Actor:      constant.StatusHistoryActorUser,
				ReasonType: "",
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
			request: &statusHistoryModel.RecordPaymentStatusHistoryRequest{
				PaymentID:  "test-payment-id",
				Status:     constant.PaymentStatusHistoryPending,
				Actor:      constant.StatusHistoryActorUser,
				ReasonType: "",
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

			svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			err := svc.recordStatusHistory(ctx, tc.request)

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

func TestRecordPaymentPending(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			// Verify the status history object has correct values
			assert.Equal(t, constant.TypePayment, sh.ReferenceType)
			assert.Equal(t, paymentID, sh.ReferenceID)
			assert.Equal(t, constant.PaymentStatusHistoryPending, sh.Status)

			// Verify metadata
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, "Payment Created", metadata.Label)

			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentPending(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentSuccess(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistorySuccess, sh.Status)
			assert.Equal(t, "Success", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentSuccess(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentProcessing(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryProcessing, sh.Status)
			assert.Equal(t, "Processing", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentProcessing(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentVoid(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryVoid, sh.Status)
			assert.Equal(t, "Void", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentVoid(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentExpired(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryExpired, sh.Status)
			assert.Equal(t, "Expired", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentExpired(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentCancelled(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := "user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryCancelled, sh.Status)
			assert.Equal(t, "Cancelled", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentCancelled(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentRequireAction(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryRequireAction, sh.Status)
			assert.Equal(t, "Requires Action", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentRequireAction(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentRequirePaymentMethod(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryRequirePaymentMethod, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentRequirePaymentMethod(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentRequireConfirmation(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := "test-user"

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryRequireConfirmation, sh.Status)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentRequireConfirmation(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}

func TestRecordPaymentPaid(t *testing.T) {
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ctx := context.Background()
	paymentID := "test-payment-id"
	actor := constant.StatusHistoryActorSystem

	statusHistoriesRepo.On(
		"Insert",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(sh *statusHistoryModel.StatusHistory) bool {
			var metadata statusHistoryModel.StatusHistoryMetadata
			err := json.Unmarshal(sh.Metadata.JSONText, &metadata)
			require.NoError(t, err)
			assert.Equal(t, actor, metadata.Actor)
			assert.Equal(t, constant.PaymentStatusHistoryPaid, sh.Status)
			assert.Equal(t, "Paid", metadata.Label)
			return true
		}),
	).Return(nil).Once()

	svc := New(nil, mockLogger, nil, nil, nil, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
	svc.recordPaymentPaid(ctx, paymentID, actor)

	statusHistoriesRepo.AssertExpectations(t)
}
