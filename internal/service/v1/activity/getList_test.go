package activityService

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	data := make([]activityModel.Activity, 0)
	data = append(data, activityModel.Activity{
		ID: uuid.NewString(),
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
		MockSetup      func(mockRepo *repositoryMocks.IActivityRepository)
	}{
		{
			Name:           "SUCCESS: getList",
			WantErr:        false,
			ExpectedResult: &expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.IActivityRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("activityModel.ActivityFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(&expectedResponse, nil)
			},
		},
		{
			Name:           "FAILED: got error on repository when getList",
			WantErr:        true,
			ExpectedResult: &expectedResponse,
			ExpectedError:  "failed to getList",
			MockSetup: func(mockRepo *repositoryMocks.IActivityRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("activityModel.ActivityFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("failed to getList"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIActivityRepository(t)
			ctx := context.Background()
			tc.MockSetup(mockRepo)

			activityService := New(mockRepo)
			response, err := activityService.GetList(ctx, activityModel.ActivityFilterRequest{}, 1, 20)
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
