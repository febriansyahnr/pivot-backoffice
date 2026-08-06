package ledgerService

import (
	"context"
	"errors"
	"testing"

	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockPkg "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestGetLedgerDetail(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(repo *mockRepo.IAccountTransactionRepository)
		referenceId string
		wantErr     bool
	}{
		{
			name: "success",
			setup: func(repo *mockRepo.IAccountTransactionRepository) {
				repo.On("GetLedgerDetail", mock.Anything, "valid-reference-id").Return([]orchestrator_model.AccountTransaction{}, nil)
			},
			referenceId: "valid-reference-id",
			wantErr:     false,
		},
		{
			name: "error",
			setup: func(repo *mockRepo.IAccountTransactionRepository) {
				repo.On("GetLedgerDetail", mock.Anything, "invalid-reference-id").Return(nil, errors.New("error"))
			},
			referenceId: "invalid-reference-id",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			atRepo := mockRepo.NewIAccountTransactionRepository(t)
			accRepo := mockRepo.NewIAccountRepository(t)
			merchantService := mockSvc.NewIMerchantService(t)
			accSvc := mockSvc.NewIAccountService(t)
			logger, _ := mockPkg.NewZapLogger(mockPkg.Config{})
			tc.setup(atRepo)
			s := New(logger, atRepo, accRepo, merchantService, nil, accSvc)
			_, err := s.GetLedgerDetail(context.Background(), tc.referenceId)
			if tc.wantErr {
				if err == nil {
					t.Errorf("want error, got nil")
				}
			}
			atRepo.AssertExpectations(t)
		})
	}
}
