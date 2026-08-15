package disbursementService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindByID(t *testing.T) {
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	// General status history mock that will handle any calls
	statusHistoriesRepo.On("GetByReference", mock.Anything, mock.Anything, mock.Anything).Return([]*statusHistoryModel.StatusHistory{}, nil).Maybe()

	tests := []struct {
		name         string
		modifierMock func()
		wantErr      string
	}{
		{
			name: "ERROR: FindByID service",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR: Disbursement not found",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(nil, nil)
			},
			wantErr: constant.ErrDisbursementNotFound.Error(),
		},
		{
			name: "SUCCESS",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MetadataObj: disbursementModel.Metadata{},
					},
				}, nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.modifierMock()

			svc := New(&conf, pdkLoggerMock, nil, disbursementRepo, nil, nil, WithStatusHistoriesRepository(statusHistoriesRepo))
			_, err := svc.FindByID(ctx, uuid.NewString())
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
