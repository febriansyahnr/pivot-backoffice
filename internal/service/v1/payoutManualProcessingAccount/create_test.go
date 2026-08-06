package payoutManualProcessingAccount_test

import (
	"context"
	"errors"
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

func TestCreate(t *testing.T) {
	traceID := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)

	createReq := &payoutManualProcessingAccountModel.CreatePayoutManualProcessingAccountRequest{
		MerchantID:    "merchant-123",
		BankCode:      "BCA",
		AccountNumber: "1234567890",
		UpdatedBy:     "John Doe",
	}

	dataMockType := mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount")

	tests := []struct {
		name       string
		request    *payoutManualProcessingAccountModel.CreatePayoutManualProcessingAccountRequest
		setupMocks func(repo *repoMocks.IPayoutManualProcessingAccountRepository)
		wantErr    bool
	}{
		{
			name:    "SUCCESS: Payout manual processing account created",
			request: createReq,
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("Create", c.ValueCtxMockType(), dataMockType).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Duplicate entry already exists",
			request: createReq,
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("Create", c.ValueCtxMockType(), dataMockType).
					Return(errors.New("Error 1062: Duplicate entry 'merchant-123-BCA-1234567890' for key 'uniq_merchant_bank_account'"))
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Failed to create account",
			request: createReq,
			setupMocks: func(repo *repoMocks.IPayoutManualProcessingAccountRepository) {
				repo.
					On("Create", c.ValueCtxMockType(), dataMockType).
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
			_, err := svc.Create(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("PayoutManualProcessingAccount.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
