package callbackService

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	data := make([]callbackModel.Callback, 0)
	data = append(data, callbackModel.Callback{
		UUID: uuid.New(),
	})
	expectedResponse := &commonModel.PaginationResponse{
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
		MockSetup      func(mockRepo *repositoryMocks.ICallbackRepository)
	}{
		{
			Name:           "SUCCESS: getList",
			WantErr:        false,
			ExpectedResult: expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.GetListCallbackFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
		},
		{
			Name:           "FAILED: got error on repository when getList",
			WantErr:        true,
			ExpectedResult: expectedResponse,
			ExpectedError:  "failed to getList",
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.GetListCallbackFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("failed to getList"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewICallbackRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCache := redisExtMocks.NewIRedisExt(t)
			ctx := context.Background()
			tc.MockSetup(mockRepo)

			cbService := New(mockLogger, mockCache, mockRepo, nil, nil)
			response, err := cbService.GetList(ctx, &callbackModel.GetListCallbackFilterRequest{}, 1, 20)
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
