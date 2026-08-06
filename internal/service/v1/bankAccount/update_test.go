package bankAccount

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	traceId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	request := &bankAccount.UpdateBankAccountRequest{
		UUID:                   "01922291-3f2e-734f-a65b-e3595f1318f5",
		MerchantID:             "merchant-id",
		BeneficiaryAccountNo:   "1234567890",
		BeneficiaryAccountName: "John Doe",
		BeneficiaryBankCode:    "123",
		BeneficiaryBankName:    "Bank ABC",
		UpdatedBy:              "kaI",
	}

	dataMockType := mock.AnythingOfType("*bankAccount.BankAccount")

	tests := []struct {
		name       string
		request    *bankAccount.UpdateBankAccountRequest
		setupMocks func(repo *repoMocks.IBankAccountRepository)
		wantErr    bool
	}{
		{
			name:    "SUCCESS: Bank account updated",
			request: request,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), request.MerchantID).
					Return(&bankAccount.BankAccount{
						UUID:       "01922291-3f2e-734f-a65b-e3595f1318f5",
						MerchantID: "merchant-id",
						CreatedBy:  "user-id",
						CreatedAt:  time.Now(),
						UpdatedBy:  "user-id",
					}, nil)

				repo.
					On("Update", c.ValueCtxMockType(), dataMockType).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Bank account not updated",
			request: request,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), request.MerchantID).
					Return(&bankAccount.BankAccount{
						UUID:       "01922291-3f2e-734f-a65b-e3595f1318f5",
						MerchantID: "merchant-id",
						CreatedBy:  "user-id",
						CreatedAt:  time.Now(),
						UpdatedBy:  "user-id",
					}, nil)

				repo.
					On("Update", c.ValueCtxMockType(), dataMockType).
					Return(c.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name:    "ERROR: GetByMerchantID error",
			request: request,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), request.MerchantID).
					Return(nil, c.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name:    "ERROR: Bank account not found",
			request: request,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), request.MerchantID).
					Return(nil, nil)

			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			bankAccRepo := repoMocks.NewIBankAccountRepository(t)

			tt.setupMocks(bankAccRepo)

			svc := New(bankAccRepo, logger)
			_, err := svc.Update(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("BankAccount.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
