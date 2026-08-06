package callbackService_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	callbackService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetCallbackLogList(t *testing.T) {
	data := make([]callbackModel.CallbackLogWithMaster, 0)
	data = append(data, callbackModel.CallbackLogWithMaster{
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
		ExpectedError  error
		MockSetup      func(mockRepo *repositoryMocks.ICallbackRepository)
	}{
		{
			Name:           "SUCCESS: getList",
			WantErr:        false,
			ExpectedResult: expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"FindMerchantCallbackHistory", mock.Anything, constant.PtrGetListCallbackLogFilterRequestMockType(), mock.Anything, mock.Anything,
				).Return(expectedResponse, nil)
			},
		},
		{
			Name:          "ERROR: getList",
			WantErr:       true,
			ExpectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"FindMerchantCallbackHistory", mock.Anything, constant.PtrGetListCallbackLogFilterRequestMockType(), mock.Anything, mock.Anything,
				).Return(nil, assert.AnError)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			mockRepo := repositoryMocks.NewICallbackRepository(t)

			tc.MockSetup(mockRepo)

			cbService := callbackService.New(nil, nil, mockRepo, nil, nil)
			response, err := cbService.GetCallbackLogList(t.Context(), &callbackModel.GetListCallbackLogFilterRequest{}, 1, 20)
			if tc.WantErr {
				require.Nil(t, response)
				require.Equal(t, tc.ExpectedError, err)

			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
