package disbursementService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateBulk(t *testing.T) {
	validRequest := &disbursementModel.CreateBulkDisbursementRequest{}
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.IDisbursementRepository)
		input     *disbursementModel.CreateBulkDisbursementRequest
	}{
		{
			name:    "SUCCESS: Create",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.On(
					"InsertBulkDisbursement",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.BulkDisbursement"),
				).Return(nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: Got error in Insert",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.On(
					"InsertBulkDisbursement",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.BulkDisbursement"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)
			// mockLog, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			ctx := context.Background()
			tc.mockSetup(mockRepo)

			svc := New(&conf, pdkLoggerMock, merchantRepo, mockRepo, nil, nil)
			response, err := svc.CreateBulk(ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			mockRepo.AssertExpectations(t)
			//
		})
	}
}
