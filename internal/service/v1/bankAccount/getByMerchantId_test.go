package bankAccount

import (
	"context"
	"testing"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestGetByMerchantID(t *testing.T) {
	traceId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	createReq := &bankAccount.CreateBankAccountRequest{
		UUID:                   "01922291-3f2e-734f-a65b-e3595f1318f5",
		MerchantID:             "merchant-id",
		BeneficiaryAccountNo:   "1234567890",
		BeneficiaryAccountName: "John Doe",
		BeneficiaryBankCode:    "123",
		BeneficiaryBankName:    "Bank ABC",
		CreatedBy:              "John Doe",
	}

	dataMockType := mock.AnythingOfType("*bankAccount.BankAccount")

	tests := []struct {
		name       string
		request    *bankAccount.CreateBankAccountRequest
		setupMocks func(repo *repoMocks.IBankAccountRepository)
		wantErr    bool
	}{
		{
			name:    "SUCCESS: Bank account created",
			request: createReq,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), createReq.MerchantID).
					Return(nil, nil)

				repo.
					On("Create", c.ValueCtxMockType(), dataMockType).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Bank account not created",
			request: createReq,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), createReq.MerchantID).
					Return(nil, nil)

				repo.
					On("Create", c.ValueCtxMockType(), dataMockType).
					Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: GetByMerchantID error",
			request: createReq,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), createReq.MerchantID).
					Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Bank account already exist",
			request: createReq,
			setupMocks: func(repo *repoMocks.IBankAccountRepository) {
				repo.
					On("GetByMerchantID", c.ValueCtxMockType(), createReq.MerchantID).
					Return(&bankAccount.BankAccount{
						UUID:       "01922291-3f2e-734f-a65b-e3595f1318f5",
						MerchantID: "merchant-id",
					}, nil)
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
			_, err := svc.Create(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("BankAccount.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
