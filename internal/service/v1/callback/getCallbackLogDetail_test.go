package callbackService_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	callbackService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetCallbackLogDetail(t *testing.T) {
	merchantID := uuid.NewString()
	expectedResponse := &callbackModel.CallbackLogWithMaster{
		UUID:       uuid.New(),
		MerchantID: merchantID,
	}

	testCases := []struct {
		Name           string
		MerchantID     string
		WantErr        bool
		ExpectedResult *callbackModel.CallbackLogWithMaster
		MockSetup      func(mockRepo *repositoryMocks.ICallbackRepository)
	}{
		{
			Name:           "SUCCESS",
			MerchantID:     merchantID,
			WantErr:        false,
			ExpectedResult: expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(expectedResponse, nil)
			},
		},
		{
			Name:           "ERROR: Data not found",
			MerchantID:     merchantID,
			WantErr:        true,
			ExpectedResult: expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
		},
		{
			Name:           "ERROR: Service error",
			MerchantID:     merchantID,
			WantErr:        true,
			ExpectedResult: expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			Name:           "ERROR: Merchant not match",
			MerchantID:     "",
			WantErr:        true,
			ExpectedResult: expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.ICallbackRepository) {
				mockRepo.On(
					"GetCallbackLogByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(expectedResponse, nil)
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

			cbService := callbackService.New(mockLogger, mockCache, mockRepo, nil, nil)
			response, err := cbService.GetCallbackLogDetail(ctx, uuid.NewString(), tc.MerchantID)
			if tc.WantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
