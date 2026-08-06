package withdrawalService_test

import (
	"context"
	"errors"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPreparation(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	bankAccRepo := repoMocks.NewIBankAccountRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)

	service := New(logger, nil, orchestratorSvc, bankAccRepo, nil)

	merchantId := uuid.NewString()
	listBankAccount := []bankAccount.BankAccountResponse{{
		BeneficiaryBankCode:    "002",
		BeneficiaryBankName:    "BANK DUMMY",
		BeneficiaryAccountNo:   "000000000001",
		BeneficiaryAccountName: "JOHN WICK",
	}}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *withdrawal.PreparationResponse
	}{
		{
			name: "ERROR:Get list bank account",
			setupMock: func() {
				bankAccRepo.On(
					"GetListBankAccount", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest) // NOSONAR
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Bank account not found",
			setupMock: func() {
				bankAccRepo.On(
					"GetListBankAccount", c.ValueCtxMockType(), merchantId,
				).Once().Return(nil, nil) // NOSONAR
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank account not found")), // NOSONAR
		},
		{
			name: "ERROR:Get available balance",
			setupMock: func() {
				bankAccRepo.On("GetListBankAccount", c.ValueCtxMockType(), merchantId).Return(listBankAccount, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, c.TypePayment,
				).Once().Return(0.0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest, // NOSONAR
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, c.TypePayment,
				).Return(1_250_000.00, nil)
			},
			wantResult: &withdrawal.PreparationResponse{
				MerchantId:       merchantId,
				AccountName:      c.TypePayment,
				AvailableBalance: 1_250_000,
				BankAccounts:     listBankAccount,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.Preparation(context.Background(), &withdrawal.PreparationRequest{
				MerchantId:  merchantId,
				AccountName: c.TypePayment,
			})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
