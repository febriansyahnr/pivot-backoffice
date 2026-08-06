package disbursementService

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTranslateSnapcoreStatus(t *testing.T) {
	testCases := []struct {
		name           string
		inputStatus    string
		expectedStatus string
	}{
		{
			name:           "Success status",
			inputStatus:    constant.SnapCoreBankTransferStatusSuccess,
			expectedStatus: constant.StatusSuccess,
		},
		{
			name:           "Failed status",
			inputStatus:    constant.SnapCoreBankTransferStatusFailed,
			expectedStatus: constant.StatusFailed,
		},
		{
			name:           "Pending status",
			inputStatus:    constant.SnapCoreBankTransferStatusPending,
			expectedStatus: constant.StatusPending,
		},
		{
			name:           "Unknown status",
			inputStatus:    "UNKNOWN",
			expectedStatus: constant.StatusFailed,
		},
		{
			name:           "Empty status",
			inputStatus:    "",
			expectedStatus: constant.StatusFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := translateSnapcoreStatus(tc.inputStatus)
			assert.Equal(t, tc.expectedStatus, result)
		})
	}
}

func TestProcessUpdateTransferStatus(t *testing.T) {
	type Mocker struct {
		accountTransaction    *repositoryMocks.IAccountTransactionRepository
		orchestratorSvc       *serviceMocks.IOrchestratorService
		disbursementRepo      *repositoryMocks.IDisbursementRepository
		statusHistoriesRepo   *repositoryMocks.IStatusHistoriesRepository
		merchantRepo          *repositoryMocks.IMerchantRepository
		snapCoreRepoMock      *repositoryMocks.ISnapCoreRepository
		beneficiaryAccSvcMock *serviceMocks.IBeneficiaryAccountService
		forbiddenUsecaseSvc   *serviceMocks.IMerchantForbiddenUseCaseService
		feeSvc                *serviceMocks.IFeeService
		transferSvc           *serviceMocks.ITransferService
		ledgerSvc             *serviceMocks.ILedgerService
	}

	var (
		externalID         = uuid.New()
		disbursementID     = uuid.New()
		transactionID      = util.ParseUUID("71d5e465-4ccb-4b46-b000-651c7dcc7bc4")
		feeID              = util.ParseUUID("942169a4-d56a-48e2-a910-641c5f2ced7c")
		bankTransferUUID   = uuid.New()
		bulkDisbursementID = uuid.NewString()
		transferFeeID      = uuid.NewString()
	)

	testCases := []struct {
		name          string
		setupMocks    func(m Mocker)
		input         *routingProcessorModel.BankTransferResponseData
		expectedError bool
	}{
		{
			name: "Success - Transaction found and updated to success",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									DeductionType: constant.MerchantFeeDeductionTypeDirect,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
						AdditionalInfoObj: orchestratorModel.FeeTransactionMetadataObject{
							FeeMetadataObject: feeModel.FeeMetadataObject{
								TransferId: transferFeeID,
							},
						},
					}, nil)
				m.disbursementRepo.
					On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, disbursementID.String(), bankTransferUUID.String(), "1234567890").
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, transactionID.String(), constant.StatusSuccess, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusSuccess, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
				m.ledgerSvc.On("UpdateTransaction", mock.Anything, mock.Anything).Once().Return(nil)
				m.transferSvc.On("UpdateTransferStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:      externalID.String(),
				Status:          constant.SnapCoreBankTransferStatusSuccess,
				ResponseCode:    "200xx00",
				BankReferenceNo: "1234567890",
				UUID:            bankTransferUUID.String(),
			},
			expectedError: false,
		},
		{
			name: "Success - Transaction found and updated to success",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							BulkID: &bulkDisbursementID,
							UUID:   disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									DeductionType: constant.MerchantFeeDeductionTypeAutomated,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.disbursementRepo.
					On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, disbursementID.String(), bankTransferUUID.String(), "1234567890").
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, transactionID.String(), constant.StatusSuccess, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.
					On("CountStatusInProgressByBulkID", mock.Anything, bulkDisbursementID).
					Return(1)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:      externalID.String(),
				Status:          constant.SnapCoreBankTransferStatusSuccess,
				ResponseCode:    "200xx00",
				BankReferenceNo: "1234567890",
				UUID:            bankTransferUUID.String(),
			},
			expectedError: false,
		},
		{
			name: "Success - Transaction found and updated",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "403xx14",
			},
			expectedError: false,
		},
		{
			name: "Success - Transaction Inactive Account",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "403xx18",
			},
			expectedError: false,
		},
		{
			name: "Success - Transaction Dormant Account",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "403xx09",
			},
			expectedError: false,
		},
		{
			name: "Success - Transaction Invalid Account",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "404xx11",
			},
			expectedError: false,
		},
		{
			name: "Success - CutOff Time reason from additionalInfo sets PENDING and suppresses callback",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							BulkID: &bulkDisbursementID,
							UUID:   disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "202xx00",
				AdditionalInfo: map[string]any{
					"reasonType": "CUT_OFF_TIME",
				},
			},
			expectedError: false,
		},
		{
			name: "Success - No additionalInfo preserves existing behavior",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "403xx14",
			},
			expectedError: false,
		},
		{
			name: "Success - AdditionalInfo with different reasonType preserves existing behavior",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "403xx14",
				AdditionalInfo: map[string]any{
					"reasonType": "SOME_OTHER_REASON",
				},
			},
			expectedError: false,
		},
		{
			name: "Success - LADDER tiering counter incremented on success",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									DeductionType:          constant.MerchantFeeDeductionTypeDirect,
									LadderCounterKey:       "backend-portal:merchant-fee-counter:fee-uuid:2026-03",
									LadderCounterIncrement: 500_000,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
						AdditionalInfoObj: orchestratorModel.FeeTransactionMetadataObject{
							FeeMetadataObject: feeModel.FeeMetadataObject{
								TransferId: transferFeeID,
							},
						},
					}, nil)
				m.disbursementRepo.
					On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, disbursementID.String(), bankTransferUUID.String(), "1234567890").
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, transactionID.String(), constant.StatusSuccess, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusSuccess, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
				m.ledgerSvc.On("UpdateTransaction", mock.Anything, mock.Anything).Once().Return(nil)
				m.transferSvc.On("UpdateTransferStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				m.feeSvc.On("IncrementLadderCounter", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", int64(500_000)).Once()
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:      externalID.String(),
				Status:          constant.SnapCoreBankTransferStatusSuccess,
				ResponseCode:    "200xx00",
				BankReferenceNo: "1234567890",
				UUID:            bankTransferUUID.String(),
			},
			expectedError: false,
		},
		{
			name: "Success - LADDER tiering counter NOT incremented on pending status",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type:                   constant.MerchantFeeDeductionTypeManual,
									LadderCounterKey:       "backend-portal:merchant-fee-counter:fee-uuid:2026-03",
									LadderCounterIncrement: 1,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
				// IncrementLadderCounter should NOT be called -- feeSvc mock will fail if it's called unexpectedly
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "403xx14",
			},
			expectedError: false,
		},
		{
			name: "Error - Commit Transaction",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: feeID,
					}, nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.orchestratorSvc.
					On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusPending, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).
					Return(nil)
				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(errors.New("error"))
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "404xx11",
			},
			expectedError: true,
		},
		{
			name: "Error - Find Account Transaction",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(nil, errors.New("error"))
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "404xx11",
			},
			expectedError: true,
		},
		{
			name: "Error - Find Account Transaction with nil error",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(nil, nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "404xx11",
			},
			expectedError: true,
		},
		{
			name: "Error - Find Transaction Fee",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(nil, errors.New("error"))
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "404xx11",
			},
			expectedError: true,
		},
		{
			name: "Error - Find Transaction Fee with error nil",
			setupMocks: func(m Mocker) {
				m.accountTransaction.
					On("FindByID", mock.Anything, externalID.String()).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:        transactionID,
						ReferenceID: disbursementID.String(),
					}, nil)
				m.disbursementRepo.
					On("FindByID", mock.Anything, disbursementID.String()).
					Return(&disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID: disbursementID.String(),
							MetadataObj: disbursementModel.Metadata{
								FeeDetail: feeModel.FeeMetadataObject{
									Type: constant.MerchantFeeDeductionTypeManual,
								},
							},
						},
					}, nil)
				m.orchestratorSvc.
					On("FindByReference", mock.Anything, disbursementID.String(), constant.TypeFee).
					Return(nil, nil)
			},
			input: &routingProcessorModel.BankTransferResponseData{
				ExternalID:   externalID.String(),
				Status:       constant.SnapCoreBankTransferStatusPending,
				ResponseCode: "404xx11",
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			snapCoreRepoMock := repositoryMocks.NewISnapCoreRepository(t)
			beneficiaryAccSvcMock := serviceMocks.NewIBeneficiaryAccountService(t)
			forbiddenUsecaseSvc := serviceMocks.NewIMerchantForbiddenUseCaseService(t)
			feeSvc := serviceMocks.NewIFeeService(t)
			db, _ := redismock.NewClientMock()
			transferSvc := serviceMocks.NewITransferService(t)
			ledgerSvc := serviceMocks.NewILedgerService(t)

			// General status history mock that will handle any calls
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			tc.setupMocks(Mocker{
				accountTransaction:    accountTransactionRepo,
				orchestratorSvc:       orchestratorSvc,
				disbursementRepo:      disbursementRepo,
				statusHistoriesRepo:   statusHistoriesRepo,
				merchantRepo:          merchantRepo,
				snapCoreRepoMock:      snapCoreRepoMock,
				beneficiaryAccSvcMock: beneficiaryAccSvcMock,
				forbiddenUsecaseSvc:   forbiddenUsecaseSvc,
				feeSvc:                feeSvc,
				transferSvc:           transferSvc,
				ledgerSvc:             ledgerSvc,
			})
			conf := config.Config{
				Environment: constant.EnvironmentStaging,
			}
			svc := New(
				&conf, pdkLoggerMock, merchantRepo, disbursementRepo, snapCoreRepoMock, nil,
				WithStatusHistoriesRepository(statusHistoriesRepo),
				WithOrchestratorService(orchestratorSvc),
				WithBeneficiaryAccService(beneficiaryAccSvcMock),
				WithMerchantForbiddenUseCaseService(forbiddenUsecaseSvc),
				WithRedisClient(redisExt.WrapRedisClient(db, nil)),
				WithFeeService(feeSvc),
				WithAccountTransactionRepository(accountTransactionRepo),
				WithTransferService(transferSvc), WithLedgerService(ledgerSvc),
			)

			err := svc.ProcessUpdateTransferStatus(context.Background(), tc.input)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			ledgerSvc.AssertExpectations(t)
			transferSvc.AssertExpectations(t)
		})
	}
}
