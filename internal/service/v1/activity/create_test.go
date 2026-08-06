package activityService

import (
	"context"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	testCases := []struct {
		Name          string
		WantErr       bool
		ExpectedError string
		MockSetup     func(mockRepo *repositoryMocks.IActivityRepository)
	}{
		{
			Name:    "SUCCESS: create",
			WantErr: false,
			MockSetup: func(mockRepo *repositoryMocks.IActivityRepository) {
				mockRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*activityModel.Activity"),
				).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIActivityRepository(t)
			ctx := context.Background()
			tc.MockSetup(mockRepo)

			activityService := New(mockRepo)
			err := activityService.Create(ctx, &activityModel.Activity{})
			if tc.WantErr {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
