package orchestrator_service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"

	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindByReference(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"FindByReference",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&orchestrator_model.AccountTransactionWithUseCase{}, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Service error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"FindByReference",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			_, err := accTrxSvc.FindByReference(ctx, uuid.NewString(), "type")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}
