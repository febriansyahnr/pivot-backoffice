package disbursementService

import (
	"context"
	"database/sql"
	"testing"
	"time"

	rabbitMQMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessPayoutAlert(t *testing.T) {
	validInput := &disbursementModel.PayoutTransactionAlertRequest{
		DisbursementID: uuid.NewString(),
		BankProcessor:  "BNI",
		TransferType:   "INSTANT",
	}
	conf := config.Config{
		Environment: c.EnvironmentStaging,
		SlackConfig: config.SlackConfig{
			PayoutAlertWebHookURL: "https://hooks.slack.com/test",
		},
	}

	responseTransferLog := &snapCoreModel.BankTransferCheckStatusResponseData{
		UUID:         uuid.NewString(),
		BankAcquirer: "BNI",
		TransferType: "INTRABANK",
		Status:       c.StatusFailed,
		TransferLogs: []snapCoreModel.TransferLog{
			{
				Status:    c.StatusPending,
				Bank:      "PERMATA",
				Order:     1,
				Action:    "INTRABANK_transfer",
				UUID:      uuid.NewString(),
				CreatedAt: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			},
			{
				Status:    c.StatusFailed,
				Bank:      "PERMATA",
				Order:     1,
				Action:    "INQUIRY_STATUS_transfer",
				UUID:      uuid.NewString(),
				CreatedAt: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			},
		},
	}

	testCases := []struct {
		name       string
		mocksSetup func(
			disbursementRepo *repositoryMocks.IDisbursementRepository,
			merchantRepo *repositoryMocks.IMerchantRepository,
			orchestratorSvc *serviceMocks.IOrchestratorService,
			rabbitMqExt *rabbitMQMock.RabbitMQExt,
			snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
		)
		input       *disbursementModel.PayoutTransactionAlertRequest
		expectError bool
	}{
		{
			name: "VALID:Success transaction no need to send alert",
			mocksSetup: func(
				disbursementRepo *repositoryMocks.IDisbursementRepository,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				rabbitMqExt *rabbitMQMock.RabbitMQExt,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
			) {
				orchestratorSvc.On(
					"FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&orchestraModel.AccountTransactionWithUseCase{
					Status: c.StatusSuccess,
				}, nil)
			},
			input:       validInput,
			expectError: false,
		},
		{
			name: "VALID:Pending transaction sends alert and republishes",
			mocksSetup: func(
				disbursementRepo *repositoryMocks.IDisbursementRepository,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				rabbitMqExt *rabbitMQMock.RabbitMQExt,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
			) {
				orchestratorSvc.On(
					"FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&orchestraModel.AccountTransactionWithUseCase{
					Status:               c.StatusPending,
					TransactionTimestamp: time.Now(),
				}, nil)

				// Mock for sendPayoutTransactionAlert
				disbursementRepo.On(
					"FindByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:                   validInput.DisbursementID,
						MerchantID:             uuid.NewString(),
						ReferenceID:            "test-ref",
						Amount:                 decimal.NewFromInt(100000),
						BeneficiaryAccountNo:   "1234567890",
						BeneficiaryAccountName: "John Doe",
						Remark:                 stringPtr("Test remark"),
						BankReferenceNo:        stringPtr("BNK12345"),
						BeneficiaryBankName:    stringPtr("Bank Test"),
					},
				}, nil)

				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchantModel.Merchant{
					Name: "Test Merchant",
				}, nil)

				rabbitMqExt.On(
					"Publish", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)

				snapCoreRepo.On(
					"CheckStatusByExternalId", c.ValueCtxMockType(), c.StringMockType(), true,
				).Return(responseTransferLog, nil)
			},
			input:       validInput,
			expectError: false,
		},
		{
			name: "VALID:Failed transaction sends alert without republishing",
			mocksSetup: func(
				disbursementRepo *repositoryMocks.IDisbursementRepository,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				rabbitMqExt *rabbitMQMock.RabbitMQExt,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
			) {
				orchestratorSvc.On(
					"FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&orchestraModel.AccountTransactionWithUseCase{
					Status:               c.StatusFailed,
					TransactionTimestamp: time.Now(),
					ReasonDescription:    sql.NullString{String: "Transaction failed", Valid: true},
				}, nil)

				// Mock for sendPayoutTransactionAlert
				disbursementRepo.On(
					"FindByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:                   validInput.DisbursementID,
						MerchantID:             uuid.NewString(),
						ReferenceID:            "test-ref",
						Amount:                 decimal.NewFromInt(100000),
						BeneficiaryAccountNo:   "1234567890",
						BeneficiaryAccountName: "John Doe",
						Remark:                 stringPtr("Test remark"),
						BankReferenceNo:        stringPtr("BNK12345"),
						BeneficiaryBankName:    stringPtr("Bank Test"),
					},
				}, nil)

				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchantModel.Merchant{
					Name: "Test Merchant",
				}, nil)

				rabbitMqExt.On(
					"Publish", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)

				snapCoreRepo.On(
					"CheckStatusByExternalId", c.ValueCtxMockType(), c.StringMockType(), true,
				).Return(responseTransferLog, nil)
			},
			input:       validInput,
			expectError: false,
		},
		{
			name: "VALID: Success transaction by check status not sent slack alert",
			mocksSetup: func(
				disbursementRepo *repositoryMocks.IDisbursementRepository,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				rabbitMqExt *rabbitMQMock.RabbitMQExt,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				orchestratorSvc.On(
					"FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&orchestraModel.AccountTransactionWithUseCase{
					Status:               c.StatusPending,
					TransactionTimestamp: time.Now(),
					ReasonDescription:    sql.NullString{String: "Transaction pending", Valid: true},
				}, nil)

				// Mock for sendPayoutTransactionAlert
				disbursementRepo.On(
					"FindByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:                   validInput.DisbursementID,
						MerchantID:             uuid.NewString(),
						ReferenceID:            "test-ref",
						Amount:                 decimal.NewFromInt(100000),
						BeneficiaryAccountNo:   "1234567890",
						BeneficiaryAccountName: "John Doe",
						Remark:                 stringPtr("Test remark"),
						BankReferenceNo:        stringPtr("BNK12345"),
						BeneficiaryBankName:    stringPtr("Bank Test"),
					},
				}, nil)

				successresponseTransferLog := responseTransferLog
				successresponseTransferLog.Status = c.StatusSuccess
				successresponseTransferLog.TransferLogs = append(successresponseTransferLog.TransferLogs, snapCoreModel.TransferLog{
					Status:    c.StatusSuccess,
					Bank:      "PERMATA",
					Order:     2,
					Action:    "INQUIRY_STATUS_transfer",
					UUID:      uuid.NewString(),
					CreatedAt: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				})

				snapCoreRepo.On(
					"CheckStatusByExternalId", c.ValueCtxMockType(), c.StringMockType(), true,
				).Return(successresponseTransferLog, nil)

				orchestratorSvc.On("UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType()).Return(nil)

				statusHistoriesRepo.On("Insert", c.ValueCtxMockType(), mock.Anything).Return(nil)
			},
			input:       validInput,
			expectError: false,
		},
		{
			name: "INVALID:Transaction not found",
			mocksSetup: func(
				disbursementRepo *repositoryMocks.IDisbursementRepository,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				rabbitMqExt *rabbitMQMock.RabbitMQExt,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
			) {
				orchestratorSvc.On(
					"FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil, nil)
			},
			input:       validInput,
			expectError: true,
		},
		{
			name: "INVALID:FindByReference error",
			mocksSetup: func(
				disbursementRepo *repositoryMocks.IDisbursementRepository,
				merchantRepo *repositoryMocks.IMerchantRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
				rabbitMqExt *rabbitMQMock.RabbitMQExt,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
			) {
				orchestratorSvc.On(
					"FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input:       validInput,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			merchantRepoMock := repositoryMocks.NewIMerchantRepository(t)
			snapCoreRepoMock := repositoryMocks.NewISnapCoreRepository(t)
			orchSvcMock := serviceMocks.NewIOrchestratorService(t)
			beneficiaryAccSvcMock := serviceMocks.NewIBeneficiaryAccountService(t)
			rabbitMqExtMock := rabbitMQMock.NewRabbitMQExt(t)
			statusHistoriesRepoMock := repositoryMocks.NewIStatusHistoriesRepository(t)

			tc.mocksSetup(disbursementRepoMock, merchantRepoMock, orchSvcMock, rabbitMqExtMock, snapCoreRepoMock, statusHistoriesRepoMock)

			svc := New(
				&conf, pdkLoggerMock, merchantRepoMock, disbursementRepoMock, snapCoreRepoMock, nil,
				WithOrchestratorService(orchSvcMock), WithBeneficiaryAccService(beneficiaryAccSvcMock),
				WithRabbitMQClient(rabbitMqExtMock),
				WithStatusHistoriesRepository(statusHistoriesRepoMock),
			)

			ctx := context.Background()
			err := svc.ProcessPayoutAlert(ctx, tc.input)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			disbursementRepoMock.AssertExpectations(t)
			merchantRepoMock.AssertExpectations(t)
			snapCoreRepoMock.AssertExpectations(t)
			orchSvcMock.AssertExpectations(t)
			beneficiaryAccSvcMock.AssertExpectations(t)
			rabbitMqExtMock.AssertExpectations(t)
		})
	}
}
