package cardFundedPayoutService_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestProcessFinishCardFundedPayoutSettlement(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")
	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	log := logMock.NewILogger(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	statusHistoryRepo := repoMocks.NewIStatusHistoriesRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	snapCoreRepo := repoMocks.NewISnapCoreRepository(t)
	cacheClient := redisMock.NewIRedisExt(t)
	mutex := redisMock.NewIMutexer(t)

	cfg := &config.Config{}
	service := New(cfg, log,
		WithDisbursementRepository(disbursementRepo),
		WithPaymentRepository(paymentRepo),
		WithStatusHistoriesRepository(statusHistoryRepo),
		WithOrchestratorService(orchestratorSvc),
		WithSnapCoreRepository(snapCoreRepo),
		WithCacheClient(cacheClient),
	)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	referenceID := "REF/PAYOUT/202603/0001" // NOSONAR
	paymentID := "payment-id-1"             // NOSONAR

	validPayment := &paymentModel.Payment{
		UUID:        paymentID,
		ReferenceID: &referenceID,
		MerchantID:  merchantID,
		Currency:    constant.CurrencyIDR,
		Amount:      decimal.NewFromFloat(1_000_000.00),              // NOSONAR
		Fee:         util.ValueToPtr(decimal.NewFromFloat(5_000.00)), // NOSONAR
		TotalAmount: decimal.NewFromFloat(1_005_000.00),              // NOSONAR
		Status:      constant.UnifiedPaymentSessionStatusPaid,
		Type:        constant.PaymentTypeCardFundedPayout,
	}

	validFundingSummary := &model.CardFundedPayoutFundingSummary{
		PayoutID:               referenceID,
		MerchantID:             merchantID,
		TotalPayment:           1_005_000.00, // NOSONAR
		TotalSuccessSettlement: 1_005_000.00, // NOSONAR
		TotalFee:               5_000.00,     // NOSONAR
	}

	validPayout := &disbursementModel.Disbursement{
		UUID:                   referenceID,
		ReferenceID:            referenceID,
		MerchantID:             merchantID,
		Currency:               constant.CurrencyIDR,
		Amount:                 decimal.NewFromFloat(1_000_000.00),              // NOSONAR
		Fee:                    util.ValueToPtr(decimal.NewFromFloat(5_000.00)), // NOSONAR
		TotalAmount:            decimal.NewFromFloat(1_005_000.00),              // NOSONAR
		BeneficiaryBankCode:    "014",                                           // NOSONAR
		BeneficiaryBankName:    util.ValueToPtr("Bank BCA"),                     // NOSONAR
		BeneficiaryAccountNo:   "1234567890",                                    // NOSONAR
		BeneficiaryAccountName: "John Doe",                                      // NOSONAR
	}

	validPayoutWithManualAction := &disbursementModel.Disbursement{
		UUID:                   referenceID,
		ReferenceID:            referenceID,
		MerchantID:             merchantID,
		Currency:               constant.CurrencyIDR,
		Amount:                 decimal.NewFromFloat(1_000_000.00),              // NOSONAR
		Fee:                    util.ValueToPtr(decimal.NewFromFloat(5_000.00)), // NOSONAR
		TotalAmount:            decimal.NewFromFloat(1_005_000.00),              // NOSONAR
		BeneficiaryBankCode:    "002",                                           // NOSONAR
		BeneficiaryBankName:    util.ValueToPtr("BANK RAKYAT INDONESIA"),        // NOSONAR
		BeneficiaryAccountNo:   "999966660001",                                  // NOSONAR
		BeneficiaryAccountName: "John Doe",                                      // NOSONAR
	}

	successBankTransferResp := &snapCoreModel.BankTransferResponseData{
		UUID:   "snap-core-uuid-1",
		Status: constant.SnapCoreBankTransferStatusSuccess,
		Amount: commonModel.Amount{
			Value:    "1000000",
			Currency: constant.CurrencyIDR,
		},
	}

	failedBankTransferResp := &snapCoreModel.BankTransferResponseData{
		UUID:            "snap-core-uuid-1",
		Status:          constant.SnapCoreBankTransferStatusFailed,
		ResponseCode:    "500",
		ResponseMessage: "Insufficient funds",
	}

	tests := []struct {
		name      string
		request   *model.ProcessFinishCardFundedPayoutSettlementRequest
		setupMock func()
		wantError error
	}{
		{
			name:    "ERROR: GetPaymentById returns error",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Once().Return(nil, assert.AnError)
			},
			wantError: fmt.Errorf("failed get payment details: %w", assert.AnError),
		},
		{
			name:    "ERROR: Payment is not a card-funded payout",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				invalidPayment := &paymentModel.Payment{
					UUID:        paymentID,
					ReferenceID: &referenceID,
					MerchantID:  merchantID,
					Type:        constant.PaymentTypeVirtualTerminal,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Once().Return(invalidPayment, nil)
			},
			wantError: errors.New("payment is not a card-funded payout"),
		},
		{
			name:    "ERROR: Failed to acquire lock",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(validPayment, nil)
				cacheClient.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", mock.Anything).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("failed to acquire lock: %w", assert.AnError),
		},
		{
			name:    "ERROR: GetCardFundedPayoutFundingSummary returns error",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				mutex.On("LockContext", mock.Anything).Return(nil)
				mutex.On("UnlockContext", mock.Anything).Once().Return(false, assert.AnError)
				paymentRepo.On("GetCardFundedPayoutFundingSummary", mock.Anything, merchantID, referenceID, 14).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "failed to release distributed lock for card-funded payout", mock.Anything).Once().Return()
			},
			wantError: fmt.Errorf("failed get funding summary: %w", assert.AnError),
		},
		{
			name:    "ERROR: Total payment amount is zero",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				zeroFundingSummary := &model.CardFundedPayoutFundingSummary{
					PayoutID:     referenceID,
					MerchantID:   merchantID,
					TotalPayment: 0,
				}
				mutex.On("UnlockContext", mock.Anything).Return(true, nil)
				paymentRepo.On("GetCardFundedPayoutFundingSummary", mock.Anything, merchantID, referenceID, 14).Once().Return(zeroFundingSummary, nil)
			},
			wantError: errors.New("total payment amount is zero"),
		},
		{
			name:    "SUCCESS: Transactions not finalized - has pending settlement",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				pendingFundingSummary := &model.CardFundedPayoutFundingSummary{
					PayoutID:               referenceID,
					MerchantID:             merchantID,
					TotalPayment:           1_005_000.00,
					TotalWaiting:           0,
					TotalFailed:            0,
					TotalPendingSettlement: 1_005_000.00,
					TotalFee:               5_000.00,
				}
				paymentRepo.On("GetCardFundedPayoutFundingSummary", mock.Anything, merchantID, referenceID, 14).Once().Return(pendingFundingSummary, nil)
				log.On("Info", mock.Anything, fmt.Sprintf("Cannot process card-funded payout transaction for ID %s: related transactions are not finalized", referenceID), mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:    "ERROR: GetDetailForCardFundedPayoutByID returns error",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				paymentRepo.On("GetCardFundedPayoutFundingSummary", mock.Anything, merchantID, referenceID, 14).Return(validFundingSummary, nil)
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, referenceID).Once().Return(nil, assert.AnError)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
			},
			wantError: fmt.Errorf("failed get payout detail: %w", assert.AnError),
		},
		{
			name:    "ERROR: PostAccountTransaction returns error",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, referenceID).Once().Return(validPayoutWithManualAction, nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
			},
			wantError: fmt.Errorf("failed post account transaction: %w", assert.AnError),
		},
		{
			name:    "SUCCESS: Waiting for manual action",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				validPayoutWithManualAction.BeneficiaryAccountNo = "999966660002"

				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, referenceID).Once().Return(validPayoutWithManualAction, nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.Processor == constant.ManualProcessor &&
						*req.ReasonType == constant.ReasonTypeWaitingManualAction &&
						*req.ReasonDescription == constant.ReasonDescWaitingManualAction
				})).Once().Return(nil)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
				log.On("Info", mock.Anything, fmt.Sprintf("Payout ID %s triggered but requires manual action due to whitelist; process will be handled by ops team", validPayoutWithManualAction.UUID), mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:    "SUCCESS: Bank transfer success",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				disbursementRepo.On("GetDetailForCardFundedPayoutByID", mock.Anything, referenceID).Return(validPayout, nil)
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Times(2).Return(nil)
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.MatchedBy(func(req *orchestratorModel.CreateAccountTransactionRequest) bool {
					return req.Processor == "" && req.ReasonType == nil && req.ReasonDescription == nil
				})).Return(nil)
				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Once().Return(successBankTransferResp, nil)
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, validPayout.UUID, successBankTransferResp.UUID, successBankTransferResp.BankReferenceNo).Once().Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
				log.On("Info", mock.Anything, "Bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:    "SUCCESS: Bank transfer success with update additional info error",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Times(2).Return(nil)
				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Once().Return(successBankTransferResp, nil)
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, validPayout.UUID, successBankTransferResp.UUID, successBankTransferResp.BankReferenceNo).Once().Return(assert.AnError)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(assert.AnError)
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Update payout processor reference id and bank reference no", mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Update account transactions additional info", mock.Anything).Once().Return()
				log.On("Info", mock.Anything, "Bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:    "SUCCESS: Bank transfer success with update status error",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Times(2).Return(nil)
				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Once().Return(successBankTransferResp, nil)
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, validPayout.UUID, successBankTransferResp.UUID, successBankTransferResp.BankReferenceNo).Once().Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusSuccess, mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Update status account transaction (success)", mock.Anything).Once().Return()
				log.On("Info", mock.Anything, "Bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:    "SUCCESS: Bank transfer failed with status update",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Times(2).Return(nil)
				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Once().Return(failedBankTransferResp, assert.AnError)
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, validPayout.UUID, successBankTransferResp.UUID, successBankTransferResp.BankReferenceNo).Once().Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
				log.On("Info", mock.Anything, "Bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:    "SUCCESS: Bank transfer failed but update status error",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Times(2).Return(nil)
				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Once().Return(failedBankTransferResp, assert.AnError)
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo", mock.Anything, validPayout.UUID, successBankTransferResp.UUID, successBankTransferResp.BankReferenceNo).Once().Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
				log.On("Error", mock.Anything, "Update status account transaction (failed)", mock.Anything).Once().Return()
				log.On("Info", mock.Anything, "Bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name:    "SUCCESS: Bank transfer returns nil response",
			request: &model.ProcessFinishCardFundedPayoutSettlementRequest{MerchantID: merchantID, ReferenceID: paymentID},
			setupMock: func() {
				statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Times(2).Return(nil)
				snapCoreRepo.On("BankTransfer", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusFailed, mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Info", mock.Anything, "Card-funded payout funding summary for payout ID "+referenceID, mock.Anything).Once().Return()
				log.On("Info", mock.Anything, "Bank transfer status for card-funded payout transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			log.On("Info", mock.Anything, "Checking transaction settlement for payment ID "+paymentID).Once().Return()

			err := service.ProcessFinishCardFundedPayoutSettlement(t.Context(), test.request)
			require.Equal(t, test.wantError, err)

			log.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
			statusHistoryRepo.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
			snapCoreRepo.AssertExpectations(t)
			cacheClient.AssertExpectations(t)
			mutex.AssertExpectations(t)
		})
	}
}
