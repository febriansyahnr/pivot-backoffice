package withdrawalService_test

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"

	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTriggeringAutoWithdrawalProcess(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	rmq := rabbitmqExtMock.NewRabbitMQExt(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	orchestratorSvc := serviceMock.NewIOrchestratorService(t)

	service := New(
		logger, nil, orchestratorSvc, nil, nil,
		WithMerchantRepository(merchantRepo), WithRabbitMQClient(rmq),
		WithWithdrawalConfig(&config.WithdrawalConfig{
			MinAmount:            25_000,
			MaxAmount:            250_000,
			AutoWithdrawalWorker: 10,
		}),
	)

	merchantId := "69e5d41e-9578-4988-9948-77dade452e44"
	merchants := []merchant.MerchantWithActiveAutoWithdrawalStatus{
		{
			MerchantId:           merchantId,
			MerchantName:         "PT Dummy Bersama",
			AccountName:          c.TypePayment,
			BeneficiaryBankCode:  "002",
			BeneficiaryAccountNo: "111111111111",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult int64
	}{
		{
			name: "ERROR:Get list of merchants",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsWithActiveAutoWithdrawalStatus", c.ValueCtxMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsWithActiveAutoWithdrawalStatus", c.ValueCtxMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR:Get Available merchant balance",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsWithActiveAutoWithdrawalStatus", c.ValueCtxMockType(),
				).Return(merchants, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.CancelCtxMockType(), merchantId, c.TypePayment,
				).Once().Return(0.0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:No data processed",
			setupMock: func() {
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.CancelCtxMockType(), merchantId, c.TypePayment,
				).Once().Return(20_000.0, nil)
			},
		},
		{
			name:      "ERROR:Context canceled",
			setupMock: func() { cancel() },
			wantErr:   context.Canceled,
		},
		{
			name: "ERROR:Failed to publish message",
			setupMock: func() {
				ctx = context.Background()

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.CancelCtxMockType(), merchantId, c.TypePayment,
				).Return(25_000.0, nil)

				rmq.On(
					"PublishWithDelay", c.CancelCtxMockType(), c.StringMockType(), mock.Anything, mock.AnythingOfType("time.Duration"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				rmq.On(
					"PublishWithDelay", c.CancelCtxMockType(), c.StringMockType(), mock.Anything, mock.AnythingOfType("time.Duration"),
				).Once().Return(nil)
			},
			wantResult: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.TriggeringAutoWithdrawalProcess(ctx)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestForceAutoWithdrawalProcess(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	rmq := rabbitmqExtMock.NewRabbitMQExt(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	orchestratorSvc := serviceMock.NewIOrchestratorService(t)
	accountTrxRepo := repoMocks.NewIAccountTransactionRepository(t)

	balanceWithdrawalConfig := config.BalanceWithdrawalConfig{
		WithdrawalAfterInactivityDays:   4,
		NotificationAfterInactivityDays: 3,
	}
	config := &config.WithdrawalConfig{
		AutoWithdrawalWorker:         1,
		MinAmount:                    10_000,
		PaymentBalanceConfig:         balanceWithdrawalConfig,
		DisbursementBalanceConfig:    balanceWithdrawalConfig,
		VirtualTerminalBalanceConfig: balanceWithdrawalConfig,
	}

	service := New(
		logger, nil, orchestratorSvc, nil, nil,
		WithRabbitMQClient(rmq),
		WithWithdrawalConfig(config),
		WithMerchantRepository(merchantRepo),
		WithAccountTransactionRepository(accountTrxRepo),
	)

	merchantId := "33afe27b-4b93-4476-8583-a5e5002ad85c"
	accountName := "DISBURSEMENT"
	merchants := []merchant.MerchantWithdrawalDetails{
		{MerchantId: merchantId, MerchantName: "Test1", AccountName: accountName},
	}
	tests := []struct {
		name       string
		date       time.Time
		setupMock  func()
		wantResult *merchant.ForceAutoWithdrawalProcessResponse
		wantErr    error
	}{
		{
			name: "ERROR:Get merchants list",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Non-whitelisted merchant",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return([]merchant.MerchantWithdrawalDetails{
					{MerchantId: "61a8fb3-5e3c-47f0-8f2a-d8743dc8b506", MerchantName: "Non-Whitelist", AccountName: "PAYMENT"},
				}, nil)
			},
			wantResult: &merchant.ForceAutoWithdrawalProcessResponse{
				Total: 1, Skip: 1,
			},
		},
		{
			name: "ERROR:Get available merchant balance",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return(merchants, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, accountName,
				).Once().Return(0.0, c.ErrSomeErrorForUnitTest)
			},
			wantResult: &merchant.ForceAutoWithdrawalProcessResponse{
				Total: 1, Failed: 1,
			},
		},
		{
			name: "SUCCESS:Merchant balance is less then min transaction",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return(merchants, nil)
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, accountName,
				).Once().Return(9_000.00, nil)
			},
			wantResult: &merchant.ForceAutoWithdrawalProcessResponse{
				Total: 1, Skip: 1,
			},
		},
		{
			name: "ERROR:Gat last transaction date",
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return(merchants, nil)
				orchestratorSvc.On(
					"GetAvailableMerchantBalance", c.ValueCtxMockType(), merchantId, mock.Anything,
				).Return(11_000.00, nil)

				accountTrxRepo.On(
					"GetLastTransactionByAccountName", c.ValueCtxMockType(), merchantId, accountName,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantResult: &merchant.ForceAutoWithdrawalProcessResponse{
				Total: 1, Failed: 1,
			},
		},
		{
			name: "SUCCESS:Active merchant",
			date: time.Date(2024, 12, 2, 17, 2, 45, 0, time.UTC),
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return(merchants, nil)
				accountTrxRepo.On(
					"GetLastTransactionByAccountName", c.ValueCtxMockType(), merchantId, accountName,
				).Once().Return(new(time.Date(2024, 12, 2, 12, 2, 45, 0, time.UTC)), nil)

			},
			wantResult: &merchant.ForceAutoWithdrawalProcessResponse{
				Total: 1, Skip: 1,
			},
		},
		{
			name: "SUCCESS:Notify dormant merchant",
			date: time.Date(2024, 12, 2, 17, 2, 45, 0, time.UTC),
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return([]merchant.MerchantWithdrawalDetails{
					{MerchantId: merchantId, MerchantName: "Test1", AccountName: c.AccountNameVirtualTerminal},
				}, nil)
				accountTrxRepo.On(
					"GetLastTransactionByAccountName", c.ValueCtxMockType(), merchantId, c.AccountNameVirtualTerminal,
				).Once().Return(new(time.Date(2024, 11, 29, 17, 2, 45, 0, time.UTC)), nil)
				rmq.On(
					"Publish", c.ValueCtxMockType(), rabbitMqExt.CommServiceEmailRoutingKey, mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
			wantResult: &merchant.ForceAutoWithdrawalProcessResponse{
				Total: 1, Notify: 1,
			},
		},
		{
			name: "SUCCESS:Publish auto withdrawal process",
			date: time.Date(2024, 12, 2, 17, 2, 45, 0, time.UTC),
			setupMock: func() {
				merchantRepo.On(
					"GetListOfMerchantsToForceTheAutoWithdrawalProcess", c.ValueCtxMockType(),
				).Once().Return([]merchant.MerchantWithdrawalDetails{
					{MerchantId: merchantId, MerchantName: "Test1", AccountName: c.AccountNamePayment},
				}, nil)
				accountTrxRepo.On(
					"GetLastTransactionByAccountName", c.ValueCtxMockType(), merchantId, c.AccountNamePayment,
				).Once().Return(new(time.Date(2024, 11, 28, 17, 2, 45, 0, time.UTC)), nil)
				rmq.On(
					"Publish", c.ValueCtxMockType(), rabbitMqExt.WithdrawalProcessRoutingKey, mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
			wantResult: &merchant.ForceAutoWithdrawalProcessResponse{
				Total: 1, Dormant: 1,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.ForceAutoWithdrawalProcess(context.Background(), test.date)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
