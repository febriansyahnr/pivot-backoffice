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

func TestUpdate(t *testing.T) {
	traceID := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)

	accountMockType := mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount")
	uuidValue := "uuid-123"

	activeStatus := c.StatusActive
	existingAccount := &payoutManualProcessingAccountModel.PayoutManualProcessingAccount{
		UUID:          uuidValue,
		MerchantID:    "merchant-123",
		BankCode:      "BCA",
		AccountNumber: "1234567890",
		Status:        c.StatusActive,
	}

	tests := []struct {
		name       string
		request    *payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest
		setupMocks func(repo *repoMocks.IPayoutManualProcessingAccountRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: Update account",
			request: &payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest{
				UUID:      uuidValue,
				Status:    &activeStatus,
				UpdatedBy: "John Doe",
			},
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("GetByUUID", c.ValueCtxMockType(), uuidValue).
					Return(existingAccount, nil)
				repo.
					On("Update", c.ValueCtxMockType(), accountMockType).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: GetByUUID returns error",
			request: &payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest{
				UUID:      uuidValue,
				Status:    &activeStatus,
				UpdatedBy: "John Doe",
			},
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("GetByUUID", c.ValueCtxMockType(), uuidValue).
					Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Account not found",
			request: &payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest{
				UUID:      uuidValue,
				Status:    &activeStatus,
				UpdatedBy: "John Doe",
			},
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("GetByUUID", c.ValueCtxMockType(), uuidValue).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid status",
			request: &payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest{
				UUID:      uuidValue,
				Status:    nil,
				UpdatedBy: "John Doe",
			},
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("GetByUUID", c.ValueCtxMockType(), uuidValue).
					Return(existingAccount, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Failed to update account",
			request: &payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest{
				UUID:      uuidValue,
				Status:    &activeStatus,
				UpdatedBy: "John Doe",
			},
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("GetByUUID", c.ValueCtxMockType(), uuidValue).
					Return(existingAccount, nil)
				repo.
					On("Update", c.ValueCtxMockType(), accountMockType).
					Return(c.ErrSomeErrorForUnitTest)
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
			_, err := svc.Update(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("PayoutManualProcessingAccount.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
