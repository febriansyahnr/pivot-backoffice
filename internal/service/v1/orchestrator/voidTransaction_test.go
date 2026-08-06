package orchestrator_service

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVoidCreditcardTransaction(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"VoidTransaction",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			err := accTrxSvc.VoidCreditcardTransaction(ctx, &orchestratorModel.VoidTransactionRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}

}
