package orchestrator_service

import (
	"context"
	"testing"

	"errors"

	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetWalletCustomersTotalBalance(t *testing.T) {

	testCases := []struct {
		name    string
		setup   func(repo *mocks.IAccountTransactionRepository)
		wantErr bool
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			setup: func(mockRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On(
					"GetWalletCustomersTotalBalance",
					mock.Anything,
					mock.Anything,
				).Return(float64(0), nil)
			},
		},
		{
			name: "FAILED: got error on repository when getList",
			setup: func(mockRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On(
					"GetWalletCustomersTotalBalance",
					mock.Anything,
					mock.Anything,
				).Return(float64(0), errors.New("failed to getList"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLog, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			ctx := context.Background()
			tc.setup(mockRepo)

			orchService := New(mockLog, nil, mockRepo, nil)
			response, err := orchService.GetWalletCustomersTotalBalance(ctx, &orchestrator_model.GetWalletTotalBalanceRequest{})
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)

		})
	}
}
