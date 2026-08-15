package disbursementService_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetBulkDisbursementDetail(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	tests := []struct {
		name         string
		modifierMock func(disbursementRepo *repositoryMocks.IDisbursementRepository)
		wantErr      error
		wantResult   *disbursementModel.BulkDisbursementDetail
	}{
		{
			name: "SUCCESS",
			modifierMock: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				disbursementRepo.On(
					"GetBulkDisbursementDetailByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.BulkDisbursementDetail{
					UUID:   "b7b95825-6ec3-486a-a3cf-62d4ed815757",
					Status: "DONE",
				}, nil)
			},
			wantResult: &disbursementModel.BulkDisbursementDetail{
				UUID:   "b7b95825-6ec3-486a-a3cf-62d4ed815757",
				Status: "DONE",
			},
		},
		{
			name: "ERROR: Database error",
			modifierMock: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				disbursementRepo.On(
					"GetBulkDisbursementDetailByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR: Bulk disbursement not found",
			modifierMock: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				disbursementRepo.On(
					"GetBulkDisbursementDetailByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			tc.modifierMock(disbursementRepo)

			svc := New(&conf, logger, nil, disbursementRepo, nil, nil)
			result, err := svc.GetBulkDisbursementDetail(ctx, uuid.NewString())

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			}
		})
	}
}
