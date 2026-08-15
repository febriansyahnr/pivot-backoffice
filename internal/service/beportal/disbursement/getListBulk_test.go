package disbursementService

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetListBulkBulk(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}
	data := make([]disbursementModel.BulkDisbursement, 0)
	data = append(data, disbursementModel.BulkDisbursement{
		UUID: uuid.NewString(),
	})
	expectedResponse := commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		Name           string
		WantErr        bool
		ExpectedResult *commonModel.PaginationResponse
		ExpectedError  string
		MockSetup      func(mockRepo *repositoryMocks.IDisbursementRepository)
	}{
		{
			Name:           "SUCCESS: GetListBulk",
			WantErr:        false,
			ExpectedResult: &expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.On(
					"GetListBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetBulkDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(&expectedResponse, nil)
			},
		},
		{
			Name:           "FAILED: got error on repository when GetListBulk",
			WantErr:        true,
			ExpectedResult: &expectedResponse,
			ExpectedError:  "failed to GetListBulk",
			MockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.On(
					"GetListBulk",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementModel.GetBulkDisbursementFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("failed to GetListBulk"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)
			ctx := context.Background()
			tc.MockSetup(mockRepo)

			svc := New(&conf, pdkLoggerMock, merchantRepo, mockRepo, nil, nil)
			response, err := svc.GetListBulk(ctx, &disbursementModel.GetBulkDisbursementFilterRequest{}, 1, 20)
			if tc.WantErr {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
