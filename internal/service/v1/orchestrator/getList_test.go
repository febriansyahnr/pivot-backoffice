package orchestrator_service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	data := make([]orchestrator_model.AccountTransaction, 0)
	data = append(data, orchestrator_model.AccountTransaction{
		UUID: uuid.New(),
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
		MockSetup      func(mockRepo *repositoryMocks.IAccountTransactionRepository)
	}{
		{
			Name:           "SUCCESS: getList",
			WantErr:        false,
			ExpectedResult: &expectedResponse,
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.TransactionHistoryFilterRequest"),
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
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.TransactionHistoryFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("failed to getList"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLog, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			ctx := context.Background()
			tc.MockSetup(mockRepo)

			orchService := New(mockLog, nil, mockRepo, nil)
			response, err := orchService.GetList(ctx, &orchestrator_model.TransactionHistoryFilterRequest{}, 1, 20)
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
