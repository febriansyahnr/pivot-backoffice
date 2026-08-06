package merchantTopUp_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchantTopUp"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessMerchantTopUpWithVirtualAccount(t *testing.T) {
	log := loggerMock.NewILogger(t)
	repo := repoMocks.NewIMerchantTopUpRepository(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	rmq := rmqMock.NewRabbitMQExt(t)
	internalSvc := serviceMocks.NewIMerchantTopUpService(t)
	feeSvc := serviceMocks.NewIFeeService(t)

	service := New(&config.Config{}, log, nil, repo, nil,
		WithMerchantService(merchantSvc),
		WithOrchestratorService(orchestratorSvc),
		WithRabbitMQClient(rmq),
		WithInternalService(internalSvc),
		WithFeeService(feeSvc),
	)

	vaNumber := "123456789"

	tests := []struct {
		name       string
		status     string
		paidAmount string
		setupMock  func()
		wantErr    error
	}{
		{
			name:      "Transaction status failed",
			status:    c.StatusFailed,
			setupMock: func() { /* Empty */ },
		},
		{
			name: "ERROR:Reference number is not found",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(nil, nil)
			},
			wantErr: c.ErrMerchantTopUpReferenceNotFound,
		},
		{
			name: "ERROR:Get available merchant balance",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(&merchantTopUp.MerchantTopUp{AccountName: c.TypeWallet}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, c.TypeWallet,
				).Once().Return(0.0, c.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while getting available merchant balance", mock.Anything,
				).Once().Return()
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(&merchantTopUp.MerchantTopUp{AccountName: c.TypeWallet}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, c.TypeWallet,
				).Once().Return(15_000.0, nil)

				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while find merchant by id", mock.Anything,
				).Once().Return()
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(&merchantTopUp.MerchantTopUp{AccountName: c.TypeWallet}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, c.TypeWallet,
				).Once().Return(15_000.0, nil)

				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantErr: c.ErrMerchantNotFound,
		},
		{
			name:       "ERROR:Paid amount non numeric value",
			paidAmount: "ABC",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(&merchantTopUp.MerchantTopUp{AccountName: c.TypeWallet}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, c.TypeWallet,
				).Once().Return(15_000.0, nil)

				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchant.Merchant{}, nil)
			},
			wantErr: &strconv.NumError{Func: "ParseFloat", Num: "ABC", Err: strconv.ErrSyntax},
		},
		{
			name: "ERROR:Create account transaction",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(&merchantTopUp.MerchantTopUp{AccountName: c.TypeWallet}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, c.TypeWallet,
				).Once().Return(15_000.0, nil)

				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchant.Merchant{}, nil)

				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Once().Return(0.0, nil, nil)

				orchestratorSvc.On("CreateAccountTransaction", mock.Anything, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while create account transaction", mock.Anything,
				).Once().Return()
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS: But error send callback",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(&merchantTopUp.MerchantTopUp{AccountName: c.TypeWallet}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, c.TypeWallet,
				).Once().Return(15_000.0, nil)

				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchant.Merchant{}, nil)

				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Once().Return(0.0, nil, nil)

				orchestratorSvc.On("CreateAccountTransaction", mock.Anything, mock.Anything).Once().Return(nil)

				internalSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Warn", mock.Anything, "Failed while send callback", mock.Anything).Once().Return()
				rmq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On(
					"GetByReferenceNumber", mock.Anything, vaNumber,
				).Once().Return(&merchantTopUp.MerchantTopUp{AccountName: c.TypeWallet}, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, c.TypeWallet,
				).Once().Return(15_000.0, nil)

				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchant.Merchant{}, nil)

				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Once().Return(0.0, nil, nil)

				orchestratorSvc.On("CreateAccountTransaction", mock.Anything, mock.Anything).Once().Return(nil)

				internalSvc.On("SendCallback", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rmq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.status == "" {
				test.status = paymentConstant.VirtualAccountStatusPaid
			}
			if test.paidAmount == "" {
				test.paidAmount = "10000.00"
			}
			request := &paymentModel.VirtualAccountPaymentNotificationRequest{
				Number: vaNumber,
				Status: test.status,
				PaidAmount: commonModel.Amount{
					Currency: "IDR", Value: test.paidAmount,
				},
				TrxDatetime: time.Now().UTC(),
			}
			test.setupMock()
			log.On("Info", mock.Anything, "Process Merchant Top Up VA", mock.Anything).Once().Return()

			assert.Equal(t, test.wantErr, service.ProcessMerchantTopUpWithVirtualAccount(context.Background(), request))
		})
	}
}
