package cardFundedPayoutService_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateBankTransferStatus(t *testing.T) {
	log := logMock.NewILogger(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)

	service := New(nil, log,
		WithDisbursementRepository(disbursementRepo),
		WithOrchestratorService(orchestratorSvc),
		WithStatusHistoriesRepository(statusHistoryRepo),
	)

	referenceID := "REF/PAYOUT/202603/0001" // NOSONAR
	transactionUUID := uuid.MustParse("a1b2c3d4-e5f6-7890-abcd-ef1234567890")

	validPayout := &disbursementModel.Disbursement{
		UUID:                   referenceID,
		ReferenceID:            referenceID,
		MerchantID:             "12f513ca-d538-412a-92a2-6a02344d9b6c",
		Currency:               constant.CurrencyIDR,
		Amount:                 decimal.NewFromFloat(1_000_000.00),              // NOSONAR
		Fee:                    util.ValueToPtr(decimal.NewFromFloat(5_000.00)), // NOSONAR
		TotalAmount:            decimal.NewFromFloat(1_005_000.00),              // NOSONAR
		BeneficiaryBankCode:    "BCA",                                           // NOSONAR
		BeneficiaryBankName:    util.ValueToPtr("Bank BCA"),                     // NOSONAR
		BeneficiaryAccountNo:   "1234567890",                                    // NOSONAR
		BeneficiaryAccountName: "John Doe",                                      // NOSONAR
	}

	validPendingTransaction := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:        transactionUUID,
		ReferenceID: referenceID,
		Status:      constant.StatusPending,
	}

	tests := []struct {
		name      string
		request   *routingProcessorModel.BankTransferResponseData
		setupMock func()
		wantError error
	}{
		{
			name: "SUCCESS: Transaction status is final (not pending) - status update ignored",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction: &orchestratorModel.AccountTransactionWithUseCase{
					UUID:        transactionUUID,
					ReferenceID: referenceID,
					Status:      constant.StatusSuccess,
				},
			},
			setupMock: func() {
				log.On("Info", mock.Anything, "Transaction status is final, status cannot be updated. Bank transfer status update ignored", mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR: GetDetailForCardFundedPayoutByID returns error",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
			},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, referenceID).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed when get detail for card-funded payout", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR: Payout not found (nil)",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
			},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, referenceID).Once().Return(nil, nil)
			},
			wantError: constant.ErrPayoutIsNotFound,
		},
		{
			name: "ERROR: UpdateProcessorReferenceIdAndBankReferenceNo returns error",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
				BankReferenceNo:    "bank-reference-no-1", // NOSONAR
			},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, referenceID).Return(validPayout, nil)
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, validPayout.UUID, "transfer-uuid-1", "bank-reference-no-1").Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("failed update payout processor reference id and bank reference no: %w", assert.AnError),
		},
		{
			name: "ERROR: UpdateProcessorAndReconReferenceByID returns error",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
				BankReferenceNo:    "bank-reference-no-1", // NOSONAR
			},
			setupMock: func() {
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, validPayout.UUID, "transfer-uuid-1", "bank-reference-no-1").Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, transactionUUID.String(), "processor-ref-1", "transfer-uuid-1", mock.Anything).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("failed update processor and recon reference: %w", assert.AnError),
		},
		{
			name: "ERROR: UpdateStatusAccountTransaction returns error",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
				BankReferenceNo:    "bank-reference-no-1", // NOSONAR
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, transactionUUID.String(), constant.StatusSuccess, (*string)(nil), (*string)(nil)).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("failed update status account transaction: %w", assert.AnError),
		},
		{
			name: "SUCCESS: Bank transfer success - status updated to SUCCESS",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
				BankReferenceNo:    "bank-reference-no-1", // NOSONAR
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, transactionUUID.String(), constant.StatusSuccess, (*string)(nil), (*string)(nil)).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, "Update bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: Bank transfer failed - status updated to FAILED",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusFailed,
				ResponseCode:       "500",                // NOSONAR
				ResponseMessage:    "Insufficient funds", // NOSONAR
				ProcessorReference: "processor-ref-1",    // NOSONAR
				Transaction:        validPendingTransaction,
				BankReferenceNo:    "bank-reference-no-1", // NOSONAR
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, transactionUUID.String(), constant.StatusFailed, mock.Anything, mock.Anything).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, "Update bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: Bank transfer pending - status remains PENDING (no status history recorded)",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusPending,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
				BankReferenceNo:    "bank-reference-no-1", // NOSONAR
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, transactionUUID.String(), constant.StatusPending, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: Bank transfer success with recordStatusHistory error (error ignored)",
			request: &routingProcessorModel.BankTransferResponseData{
				UUID:               "transfer-uuid-1", // NOSONAR
				Status:             constant.SnapCoreBankTransferStatusSuccess,
				ProcessorReference: "processor-ref-1", // NOSONAR
				Transaction:        validPendingTransaction,
				BankReferenceNo:    "bank-reference-no-1", // NOSONAR
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, transactionUUID.String(), constant.StatusSuccess, (*string)(nil), (*string)(nil)).Once().Return(nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Info", mock.Anything, "Update bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			err := service.UpdateBankTransferStatus(t.Context(), test.request)
			require.Equal(t, test.wantError, err)

			log.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
		})
	}
}

func TestUpdatePayoutTransactionStatus(t *testing.T) {
	log := logMock.NewILogger(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)

	service := New(nil, log,
		WithDisbursementRepository(disbursementRepo),
		WithOrchestratorService(orchestratorSvc),
		WithStatusHistoriesRepository(statusHistoryRepo),
	)

	payoutID := "5fc93f16-93d8-4c2c-a0f2-27c48887617a"
	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	referenceID := "REF/PAYOUT/202603/0001" // NOSONAR
	transactionUUID := uuid.MustParse("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	bankReferenceNo := "BNK-REF-001"    // NOSONAR
	reconReferenceNo := "RECON-REF-001" // NOSONAR

	validPayout := &disbursementModel.Disbursement{
		UUID:                   payoutID,
		ReferenceID:            referenceID,
		MerchantID:             merchantID,
		Currency:               constant.CurrencyIDR,
		Amount:                 decimal.NewFromFloat(1_000_000.00),
		Fee:                    util.ValueToPtr(decimal.NewFromFloat(5_000.00)),
		TotalAmount:            decimal.NewFromFloat(1_005_000.00),
		Status:                 constant.DisbursementStatusApproved,
		BeneficiaryBankCode:    "BCA",                          // NOSONAR
		BeneficiaryBankName:    util.ValueToPtr("Bank BCA"),    // NOSONAR
		BeneficiaryAccountNo:   "1234567890",                   // NOSONAR
		BeneficiaryAccountName: "John Doe",                     // NOSONAR
		Remark:                 util.ValueToPtr("Test payout"), // NOSONAR
		MetadataObj: disbursementModel.Metadata{
			CardFundedDetail: &disbursementModel.CardFundedDetailMetadata{
				Card: &disbursementModel.CardFundedDetailMetadataCard{},
			},
		},
	}

	validPendingTransaction := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:        transactionUUID,
		ReferenceID: payoutID,
		Status:      constant.StatusPending,
		Debit:       1_000_000.00,
	}

	ctxTx := context.WithValue(t.Context(), mySqlExt.CtxSqlTx, struct{}{})

	tests := []struct {
		name       string
		request    model.PatchPayoutTransactionStatusRequest
		setupMock  func()
		wantError  error
		wantResult *model.PayoutActionResponse
	}{
		{
			name: "ERROR: GetDetailForCardFundedPayoutByID returns error",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID: payoutID,
			},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to retrieve card-funded payout detail", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: Payout not found (nil)",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID: payoutID,
			},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("payout with ID %s not found", payoutID)),
		},
		{
			name: "ERROR: Payout status is not APPROVED",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID: payoutID,
			},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Once().Return(&disbursementModel.Disbursement{
					UUID:       payoutID,
					MerchantID: merchantID,
					Status:     constant.DisbursementStatusWaiting,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("payout must be in APPROVED status; current status is %s", constant.DisbursementStatusWaiting)),
		},
		{
			name: "ERROR: FindByReference returns error",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID: payoutID,
			},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, payoutID).Return(validPayout, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, payoutID, constant.ReferenceDisbursement).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to find transaction by reference and reference type", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: Transaction not found (nil)",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID: payoutID,
			},
			setupMock: func() {
				orchestratorSvc.On("FindByReference", mock.Anything, payoutID, constant.ReferenceDisbursement).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("transaction with reference id %s and reference type %s not found", payoutID, constant.ReferenceDisbursement)),
		},
		{
			name: "ERROR: Transaction status is not PENDING",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID: payoutID,
			},
			setupMock: func() {
				orchestratorSvc.On("FindByReference", mock.Anything, payoutID, constant.ReferenceDisbursement).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   transactionUUID,
					Status: constant.StatusSuccess,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("transaction must be in PENDING status; current status is %s", constant.StatusSuccess)),
		},
		{
			name: "ERROR: BeginTransaction returns error",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID: payoutID,
			},
			setupMock: func() {
				orchestratorSvc.On("FindByReference", mock.Anything, payoutID, constant.ReferenceDisbursement).Return(validPendingTransaction, nil)
				disbursementRepo.On("BeginTransaction", mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed begin transaction", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: UpdateProcessorReferenceIdAndBankReferenceNo returns error",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID:         payoutID,
				Status:           constant.StatusSuccess,
				BankReferenceNo:  bankReferenceNo,
				ReconReferenceNo: reconReferenceNo,
			},
			setupMock: func() {
				disbursementRepo.On("BeginTransaction", mock.Anything).Return(ctxTx, nil)
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", ctxTx, payoutID, "", bankReferenceNo).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed update payout processor reference id and bank reference no", mock.Anything).Once().Return()
				disbursementRepo.On("RollbackTransaction", ctxTx).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed rollback transaction", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: UpdateProcessorAndReconReferenceByID returns error",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID:         payoutID,
				Status:           constant.StatusSuccess,
				BankReferenceNo:  bankReferenceNo,
				ReconReferenceNo: reconReferenceNo,
			},
			setupMock: func() {
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", ctxTx, payoutID, "", bankReferenceNo).Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", ctxTx, transactionUUID.String(), constant.ManualProcessor, "", reconReferenceNo).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed update processor and recon reference", mock.Anything).Once().Return()
				disbursementRepo.On("RollbackTransaction", ctxTx).Once().Return(nil)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: UpdateStatusAccountTransaction returns error",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID:         payoutID,
				Status:           constant.StatusSuccess,
				BankReferenceNo:  bankReferenceNo,
				ReconReferenceNo: reconReferenceNo,
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", ctxTx, transactionUUID.String(), constant.ManualProcessor, "", reconReferenceNo).Return(nil)
				orchestratorSvc.On("UpdateStatusAccountTransaction", ctxTx, transactionUUID.String(), constant.StatusSuccess, (*string)(nil), (*string)(nil)).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed update status account transaction", mock.Anything).Once().Return()
				disbursementRepo.On("RollbackTransaction", ctxTx).Once().Return(nil)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: CommitTransaction returns error",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID:         payoutID,
				Status:           constant.StatusSuccess,
				BankReferenceNo:  bankReferenceNo,
				ReconReferenceNo: reconReferenceNo,
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateStatusAccountTransaction", ctxTx, transactionUUID.String(), constant.StatusSuccess, (*string)(nil), (*string)(nil)).Once().Return(nil)
				statusHistoryRepo.On("Insert", ctxTx, mock.Anything).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", ctxTx).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed commit transaction", mock.Anything).Once().Return()
				disbursementRepo.On("RollbackTransaction", ctxTx).Once().Return(nil)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS: Update payout transaction status to SUCCESS",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID:         payoutID,
				Status:           constant.StatusSuccess,
				BankReferenceNo:  bankReferenceNo,
				ReconReferenceNo: reconReferenceNo,
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateStatusAccountTransaction", ctxTx, transactionUUID.String(), constant.StatusSuccess, (*string)(nil), (*string)(nil)).Once().Return(nil)
				statusHistoryRepo.On("Insert", ctxTx, mock.Anything).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", ctxTx).Once().Return(nil)
			},
			wantError: nil,
			wantResult: &model.PayoutActionResponse{
				ID:            payoutID,
				ReferenceID:   referenceID,
				BankCode:      "BCA",        // NOSONAR
				BankName:      "Bank BCA",   // NOSONAR
				AccountNumber: "1234567890", // NOSONAR
				AccountName:   "John Doe",   // NOSONAR
				Amount: commonModel.AmountRequest{
					Currency: constant.CurrencyIDR,
					Value:    1_000_000.00, // NOSONAR
				},
				Remarks:          "Test payout", // NOSONAR
				MerchantID:       merchantID,
				Status:           constant.StatusSuccess,
				BankReferenceNo:  bankReferenceNo,
				ReconReferenceNo: reconReferenceNo,
			},
		},
		{
			name: "SUCCESS: Update payout transaction status to FAILED",
			request: model.PatchPayoutTransactionStatusRequest{
				PayoutID:          payoutID,
				Status:            constant.StatusFailed,
				BankReferenceNo:   bankReferenceNo,
				ReconReferenceNo:  reconReferenceNo,
				ReasonType:        util.ValueToPtr("UNABLE_TO_PROCESS"),
				ReasonDescription: util.ValueToPtr("Bank rejected the transaction"),
			},
			setupMock: func() {
				orchestratorSvc.On("UpdateStatusAccountTransaction", ctxTx, transactionUUID.String(), constant.StatusFailed, mock.Anything, mock.Anything).Return(nil)
				statusHistoryRepo.On("Insert", ctxTx, mock.Anything).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", ctxTx).Once().Return(nil)
			},
			wantResult: &model.PayoutActionResponse{
				ID:            payoutID,
				ReferenceID:   referenceID,
				BankCode:      "BCA",        // NOSONAR
				BankName:      "Bank BCA",   // NOSONAR
				AccountNumber: "1234567890", // NOSONAR
				AccountName:   "John Doe",   // NOSONAR
				Amount: commonModel.AmountRequest{
					Currency: constant.CurrencyIDR,
					Value:    1_000_000.00, // NOSONAR
				},
				Remarks:           "Test payout", // NOSONAR
				MerchantID:        merchantID,
				Status:            constant.StatusFailed,
				BankReferenceNo:   bankReferenceNo,
				ReconReferenceNo:  reconReferenceNo,
				ReasonType:        util.ValueToPtr("UNABLE_TO_PROCESS"),
				ReasonDescription: util.ValueToPtr("Bank rejected the transaction"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.UpdatePayoutTransactionStatus(t.Context(), test.request)

			require.Equal(t, test.wantError, err)
			if err == nil {
				require.Equal(t, test.wantResult, result)
			}

			log.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
		})
	}
}
