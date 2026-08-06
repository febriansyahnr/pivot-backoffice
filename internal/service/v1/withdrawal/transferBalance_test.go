package withdrawalService_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransferBalance(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	withdrawalRepo := repoMocks.NewIWithdrawalRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	service := New(logger, withdrawalRepo, orchestratorSvc, nil, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "ERROR:Get GetAvailableMerchantBalance",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(0.0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Insufficient balance",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(0.0, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Begin transaction",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2000000.00, nil)

				withdrawalRepo.On(
					"BeginTransaction",
					c.ValueCtxMockType(),
				).Once().Return(context.Background(), c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Rollback transaction",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2000000.00, nil)

				withdrawalRepo.On(
					"BeginTransaction",
					c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), "tx", "transaction"), nil)

				withdrawalRepo.On(
					"Create",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*withdrawal.Withdrawal"),
				).Once().Return(c.ErrSomeErrorForUnitTest)

				withdrawalRepo.On(
					"RollbackTransaction",
					c.ValueCtxMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Withdrawal create",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2000000.00, nil)

				withdrawalRepo.On(
					"BeginTransaction",
					c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), "tx", "transaction"), nil)

				withdrawalRepo.On(
					"Create",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*withdrawal.Withdrawal"),
				).Once().Return(c.ErrSomeErrorForUnitTest)

				withdrawalRepo.On(
					"RollbackTransaction",
					c.ValueCtxMockType(),
				).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Create source ledger",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2000000.00, nil)

				withdrawalRepo.On(
					"BeginTransaction",
					c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), "tx", "transaction"), nil)

				withdrawalRepo.On(
					"Create",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*withdrawal.Withdrawal"),
				).Return(nil)

				orchestratorSvc.On(
					"PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)

				withdrawalRepo.On(
					"RollbackTransaction",
					c.ValueCtxMockType(),
				).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Create destination ledger",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2000000.00, nil)

				withdrawalRepo.On(
					"BeginTransaction",
					c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), "tx", "transaction"), nil)

				withdrawalRepo.On(
					"Create",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*withdrawal.Withdrawal"),
				).Return(nil)

				orchestratorSvc.On(
					"PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				orchestratorSvc.On(
					"PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)

				withdrawalRepo.On(
					"RollbackTransaction",
					c.ValueCtxMockType(),
				).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Commit transaction",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2000000.00, nil)

				withdrawalRepo.On(
					"BeginTransaction",
					c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), "tx", "transaction"), nil)

				withdrawalRepo.On(
					"Create",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*withdrawal.Withdrawal"),
				).Return(nil)

				orchestratorSvc.On(
					"PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)

				withdrawalRepo.On(
					"CommitTransaction",
					c.ValueCtxMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)

				withdrawalRepo.On(
					"RollbackTransaction",
					c.ValueCtxMockType(),
				).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2000000.00, nil)

				withdrawalRepo.On(
					"BeginTransaction",
					c.ValueCtxMockType(),
				).Return(context.WithValue(context.Background(), "tx", "transaction"), nil)

				withdrawalRepo.On(
					"Create",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*withdrawal.Withdrawal"),
				).Return(nil)

				orchestratorSvc.On(
					"PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)

				withdrawalRepo.On(
					"CommitTransaction",
					c.ValueCtxMockType(),
				).Once().Return(nil)
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			_, err := service.TransferBalance(context.Background(), &withdrawal.WithdrawalTransferBalanceRequest{
				MerchantID:             uuid.NewString(),
				UserID:                 uuid.NewString(),
				SourceAccountName:      c.TypePayment,
				DestinationAccountName: c.TypeDisbursement,
				Amount:                 1000000.00,
			})
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
