package payoutManualProcessingAccount_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	payoutManualProcessingAccountSvc "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payoutManualProcessingAccount"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestList(t *testing.T) {
	traceID := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)

	query := &payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery{
		Status:   c.StatusActive,
		Page:     1,
		PageSize: 10,
	}

	accounts := []*payoutManualProcessingAccountModel.PayoutManualProcessingAccount{
		{
			UUID:          "uuid-123",
			MerchantID:    "merchant-123",
			BankCode:      "BCA",
			AccountNumber: "1234567890",
			Status:        c.StatusActive,
		},
	}

	queryMockType := mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccountQuery")

	tests := []struct {
		name       string
		request    *payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery
		setupMocks func(repo *repoMocks.IPayoutManualProcessingAccountRepository)
		wantErr    bool
	}{
		{
			name:    "SUCCESS: List payout manual processing accounts",
			request: query,
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("List", c.ValueCtxMockType(), queryMockType).
					Return(accounts, 1, nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Failed to list accounts",
			request: query,
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("List", c.ValueCtxMockType(), queryMockType).
					Return(nil, 0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			repo := repoMocks.NewIPayoutManualProcessingAccountRepository(t)

			tt.setupMocks(repo)

			svc := payoutManualProcessingAccountSvc.New(repo, logger)
			res, err := svc.List(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("PayoutManualProcessingAccount.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if res == nil {
					t.Errorf("PayoutManualProcessingAccount.List() expected non-nil response")
				}
			}
		})
	}
}
