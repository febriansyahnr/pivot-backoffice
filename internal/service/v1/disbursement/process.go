package disbursementService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	ledgerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func (s *DisbursementService) Process(ctx context.Context, id string, isRetryTransfer bool) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/Process")
	defer segment.End()

	var (
		err                     error
		orchestratorTransaction = &orchestratorModel.TransactionAndFeeObject{}
		queueTTLLock            = time.Duration(s.config.AppConfig.DisbursementProcessExpireLockSecond) * time.Second
		traceId, _              = ctx.Value(pdkConst.CtxTraceIdKey).(string)
	)

	queueKey := fmt.Sprintf(constant.DisbursementProcessQueueLockFmt, id)
	if ok, errLock := s.redisExt.SetNX(ctx, queueKey, true, queueTTLLock).Result(); errLock != nil {
		s.logger.Error(ctx, "set exclusive queue with key "+queueKey, logger.Error(errLock))
		return pkgErrors.New(httpResponse.HttpErrDatabase, fmt.Errorf("QUEUE: "+constant.InternalErrorFmt, traceId))

	} else if !ok {
		return pkgErrors.New(httpResponse.HttpErrDupCheck, constant.ErrDisbursementIsBeingProcessed)
	}

	defer func() {
		if e := s.redisExt.Del(ctx, queueKey).Err(); e != nil {
			s.logger.Error(ctx, "clears the disbursement process queue lock", logger.Error(e))
		}
	}()

	// Find disbursement by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, id)
	if err != nil {
		return pkgErrors.New(httpResponse.HttpErrDatabase, err)
	} else if disbursement == nil {
		err = constant.ErrDisbursementNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        "Disbursement-Service",
		ReferenceId: disbursement.MerchantID,
		OriginId:    disbursement.UUID,
	})

	// Magic Number Validation
	_, err = util.ValidateMagicNumber(http.MethodGet, disbursement.BeneficiaryAccountNo)
	if s.config.Environment != constant.EnvironmentProduction && err != nil && err.Error() == constant.ErrInsufficientBalance.Error() {
		ctx = context.WithValue(ctx, constant.CtxMockInsufficientBalanceMerchant, true)
	}

	// Validate merchant balance, for ensure there is no money loss. Only validate for normal process (not retry process)
	if !isRetryTransfer {
		// Define bulk ID.
		bulkID := ""
		if disbursement.BulkID != nil {
			bulkID = *disbursement.BulkID
		}

		// Validate balance again.
		if valid, errBalanceValidate := s.validateBalanceAndUpdateIfNotValid(ctx, ctx, []string{disbursement.UUID}, &disbursementModel.ApproveRequest{
			MerchantID: disbursement.MerchantID,
			BulkID:     bulkID,
		}); errBalanceValidate != nil {
			return errBalanceValidate
		} else if !valid {
			s.logger.Error(ctx, constant.ErrInsufficientBalance.Error(), logger.Error(constant.ErrInsufficientBalance))
			return pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrInsufficientBalance)
		}
	}

	// Validate transaction already processed or not, if not exist then create and call snapCore to process it.
	isAlreadyProcessedTransaction, existedTransaction, err := s.validateProcessedTransaction(ctx, disbursement.UUID, isRetryTransfer)
	if err != nil {
		return err
	} else if isAlreadyProcessedTransaction {
		s.logger.Info(ctx, "Disbursement already processed", logger.Any("disbursementID", disbursement.UUID))
		return nil
	}

	if existedTransaction == nil {
		if trxID, feeID, errCreate := s.CreatePendingOrchestratorTransaction(ctx, disbursement); errCreate != nil {
			return errCreate
		} else {
			orchestratorTransaction.TransactionID = trxID
			orchestratorTransaction.FeeID = feeID
			orchestratorTransaction.TransferFeeID = disbursement.MetadataObj.FeeDetail.TransferId
		}
	} else {
		orchestratorTransaction.TransactionID = existedTransaction.TransactionID
		orchestratorTransaction.FeeID = existedTransaction.FeeID
		orchestratorTransaction.TransferFeeID = existedTransaction.TransferFeeID
	}
	orchestratorTransaction.MerchantID = disbursement.MerchantID

	if errProcess := s.ProcessBankTransferAndUpdateTransaction(ctx, disbursement, orchestratorTransaction); errProcess != nil {
		return errProcess
	}

	return nil
}

func (s *DisbursementService) validateProcessedTransaction(ctx context.Context, disbursementID string, isRetryTransfer bool) (
	isProcessedTransaction bool,
	orchestratorTransaction *orchestratorModel.TransactionAndFeeObject,
	err error,
) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/validateProcessedTransaction")
	defer segment.End()

	// Check existed transaction by disbursement.UUID
	existedTransaction, err := s.orchestratorSvc.FindByReference(ctx, disbursementID, constant.TypeDisbursement)
	if err != nil {
		return false, nil, err
	} else if existedTransaction == nil {
		return false, nil, nil
	}

	orchestratorTransaction = &orchestratorModel.TransactionAndFeeObject{
		MerchantID:    existedTransaction.MerchantID.String(),
		TransactionID: existedTransaction.UUID.String(),
	}

	var feeMetadata feeModel.FeeMetadataObject
	existedFee, err := s.orchestratorSvc.FindByReference(ctx, disbursementID, constant.TypeFee)
	if err != nil {
		return false, nil, err
	} else if existedFee != nil {
		orchestratorTransaction.FeeID = existedFee.UUID.String()

		_ = json.Unmarshal(existedFee.AdditionalInfo.JSONText, &feeMetadata)

		orchestratorTransaction.TransferFeeID = feeMetadata.TransferId
	}

	if existedTransaction.Status != constant.StatusPending {
		s.logger.Info(ctx, "Transaction already processed", logger.Any("account_transaction", existedTransaction))
		return true, orchestratorTransaction, nil
	}

	// Extract forceFailed from context (default to false)
	forceFailed := false
	if val := ctx.Value(constant.CtxForceFailed); val != nil {
		forceFailed, _ = val.(bool)
	}

	// Check to snapCore bank transfer data by account transaction id
	existedBankTransfer, err := s.routingProcessorSvc.GetTransferByID(ctx, existedTransaction, forceFailed)
	if err != nil && existedBankTransfer == nil {
		return false, orchestratorTransaction, err
	}

	if existedTransaction.ProcessorReference == constant.FlipPGProcessor {
		// check Flip Escrow balance before retry
		flipBalance, err := s.routingProcessorSvc.GetFlipEscrowBalance(ctx, existedTransaction.ProcessorReference)
		if err != nil {
			return false, orchestratorTransaction, err
		}

		if flipBalance.BalanceAmount < existedTransaction.Debit {
			return false, orchestratorTransaction, constant.ErrInsufficientBalance
		}

		if existedBankTransfer.Status == "" {
			// valid to retry transfer, because balance is enough
			existedBankTransfer.Status = constant.SnapCoreBankTransferStatusFailed
		}
	}

	// Existed in snapCore bank transfer
	if existedBankTransfer != nil {
		s.logger.Info(ctx, "Found existed bank transfer in snap core",
			logger.Any("snapCore bankTransfer", existedBankTransfer),
			logger.String("bankTransferStatus", existedBankTransfer.Status),
			logger.Bool("isRetryTransfer", isRetryTransfer))

		// Force to retry transfer process for FAILED transfer
		if isRetryTransfer && existedBankTransfer.Status == constant.SnapCoreBankTransferStatusFailed {
			s.logger.Info(ctx, "Retrying disbursement for failed bank transfer", logger.Any("snapCore bankTransfer", existedBankTransfer))
			return false, orchestratorTransaction, nil
		} else if isRetryTransfer && existedBankTransfer.Status != constant.SnapCoreBankTransferStatusFailed {
			s.logger.Error(ctx, "Retry disbursement is not allowed because processor transaction not in failed status", logger.Any("account_transaction", existedTransaction))
			return false, orchestratorTransaction, constant.ErrRetryDisbursementIsNotAllowed
		}

		var (
			reasonType *string
			reasonDesc *string
		)

		if existedTransaction.ReasonType.Valid {
			reasonType = &existedTransaction.ReasonType.String
		}
		if existedTransaction.ReasonDescription.Valid {
			reasonDesc = &existedTransaction.ReasonDescription.String
		}

		// Build account transaction status from snap core resp.
		accountTransactionStatus := constant.StatusPending
		switch existedBankTransfer.Status {
		case constant.SnapCoreBankTransferStatusSuccess:
			accountTransactionStatus = constant.StatusSuccess

			if feeMetadata.DeductionType == constant.MerchantFeeDeductionTypeAutomated {
				orchestratorTransaction.FeeID = ""
			}

			// Increment deferred LADDER tiering counter now that the payout has succeeded
			if feeMetadata.LadderCounterKey != "" {
				s.feeSvc.IncrementLadderCounter(ctx, feeMetadata.LadderCounterKey, feeMetadata.LadderCounterIncrement)
			}

		case constant.SnapCoreBankTransferStatusFailed:
			accountTransactionStatus = constant.StatusFailed

			reasonOther := constant.ReasonTypeOtherReason
			reasonType = &reasonOther
			reasonDesc = &existedBankTransfer.ResponseMessage
		}

		// Set status to PENDING due to SnapCoreResponseCodeInsufficientFund
		if existedBankTransfer.Status == constant.SnapCoreBankTransferStatusFailed && util.IsPatternMatch(constant.SnapCoreResponseCodeInsufficientFundPattern, existedBankTransfer.ResponseCode) {
			accountTransactionStatus = constant.StatusPending

			reasonInsufficientFund := constant.ReasonTypeInsufficientEscrowFund
			reasonType = &reasonInsufficientFund
			reasonDesc = &existedBankTransfer.ResponseMessage
		}

		// Begin Tx
		ctx, errTrx := s.disbursementRepo.BeginTransaction(ctx)
		if errTrx != nil {
			return true, orchestratorTransaction, errTrx
		}

		isCompleted := false
		defer func() {
			if !isCompleted {
				if errTrx = s.disbursementRepo.RollbackTransaction(ctx); errTrx != nil {
					return
				}
			}
		}()

		// Update transaction with certain status
		if errUpdate := s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, accountTransactionStatus, reasonType, reasonDesc, disbursementID); errUpdate != nil {
			return true, orchestratorTransaction, err
		}

		// Commit Tx
		if errTrx = s.disbursementRepo.CommitTransaction(ctx); errTrx != nil {
			return true, orchestratorTransaction, errTrx
		}
		isCompleted = true

		s.logger.Info(ctx, "Transaction already created and processed in bank", logger.Any("snapCore bankTransfer", existedBankTransfer))
		return true, orchestratorTransaction, nil
	}

	return false, orchestratorTransaction, nil
}

func (s *DisbursementService) CreatePendingOrchestratorTransaction(
	ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction,
) (orchestratorTransactionID, orchestratorFeeID string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/createPendingOrchestratorTransaction")
	defer segment.End()

	transactionUUID := uuid.New()
	feeUUID := uuid.New()
	orchestratorTransactionID = transactionUUID.String()
	orchestratorFeeID = feeUUID.String()
	disbursement.MetadataObj.FeeDetail.LinkedTransactionId = orchestratorTransactionID

	// Parse merchant ID
	merchantUUID, err := uuid.Parse(disbursement.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error parsing merchant id", logger.Error(err))
		return "", "", pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	remark := ""
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}

	// Begin Tx
	ctx, errTrx := s.disbursementRepo.BeginTransaction(ctx)
	if errTrx != nil {
		return "", "", errTrx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if errTrx = s.disbursementRepo.RollbackTransaction(ctx); errTrx != nil {
				return
			}
		}
	}()

	// Build orchestrator transaction request
	var reasonType *string
	if disbursement.CutOffTimeStatusOngoing {
		reasonType = util.ValueToPtr(constant.ReasonTypePayoutCutOffTime)
	}

	transactionAmount, _ := disbursement.Amount.Float64()
	orchestratorRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 transactionUUID,
		ReferenceID:          disbursement.UUID,
		Type:                 orchestratorModel.TypeDisbursement,
		MerchantID:           merchantUUID,
		Currency:             disbursement.Currency,
		Credit:               0.00,
		Debit:                transactionAmount,
		Channel:              constant.ChannelBankTransfer,
		Status:               constant.StatusPending,
		ReasonType:           reasonType,
		Remarks:              remark,
		TransactionTimestamp: disbursement.CreatedAt,
	}

	// Set transaction status pending
	if err = s.orchestratorSvc.PostAccountTransaction(ctx, orchestratorRequest); err != nil {
		return "", "", err
	}

	// Create processing status history
	s.recordDisbursementProcessing(ctx, disbursement.UUID, "")

	merchantId := merchantUUID
	feeAmount, _ := disbursement.Fee.Float64() // Fees charged by Harsya

	// Transaction On Behalf Of Sub-Merchant
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		merchantId, _ = uuid.Parse(parentMerchantId)
		disbursement.MetadataObj.FeeDetail.Notes = "ON-BEHALF"

		// Fees charged by the main merchant
		if feeAmount > 0 {
			transferRequest := &transfer.TransferRequest{
				SourceMerchantID: merchantUUID,        // Sub-Merchant
				RecipientID:      merchantId.String(), // Main-Merchant
				ReferenceID:      disbursement.UUID,
				TransferType:     constant.MoneyFlowDirect,
				Amount:           feeAmount,
				Remarks:          fmt.Sprintf("Disbursement Fee Transfer - Ref: %s", disbursement.ReferenceID),
				ParentMerchantID: merchantId, // Main-Merchant
				Usecase:          constant.TypeDisbursement,
			}
			newCtx := context.WithValue(ctx, constant.CtxSetPendingTransaction, true)

			transferResult, err := s.transferSvc.Transfer(newCtx, transferRequest)
			if err != nil {
				return "", "", err
			}
			disbursement.MetadataObj.FeeDetail.TransferId = transferResult.UUID.String()
		}

		// Fees charged by Harsya
		feeAmount = disbursement.MetadataObj.FeeDetail.FinalAmount
	}

	feeTrxRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 feeUUID,
		ReferenceID:          disbursement.UUID,
		Type:                 orchestratorModel.TypeFee,
		MerchantID:           merchantId,
		Currency:             disbursement.Currency,
		Credit:               0.00,
		Debit:                feeAmount,
		Channel:              "",
		Status:               constant.StatusPending,
		ReasonType:           reasonType,
		Remarks:              remark,
		TransactionTimestamp: disbursement.CreatedAt,
		Usecase:              constant.TypeDisbursement,
	}
	feeTrxRequest.AdditionalInfo.Valid = true
	feeTrxRequest.AdditionalInfo.JSONText, _ = json.Marshal(disbursement.MetadataObj.FeeDetail)

	// Set transaction status pending
	if err = s.orchestratorSvc.PostAccountTransaction(ctx, feeTrxRequest); err != nil {
		return "", "", err
	}

	// Commit Tx
	if errTrx = s.disbursementRepo.CommitTransaction(ctx); errTrx != nil {
		return "", "", errTrx
	}
	isCompleted = true

	return orchestratorTransactionID, orchestratorFeeID, nil
}

func (s *DisbursementService) ValidateBankAccountAndUpdateTransaction(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction, orchestratorTransaction *orchestratorModel.TransactionAndFeeObject) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/validateBankAccountAndUpdateTransaction")
	defer segment.End()

	// Find derived merchant ID
	merchantID := disbursement.MerchantID
	if derivedMerchantID, _ := ctx.Value(constant.CtxDerivedMerchantID).(string); derivedMerchantID != "" {
		merchantID = derivedMerchantID
	}

	beneficiaryAccount, err := s.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &beneficiaryAccountModel.CheckAccountRequest{
		MerchantID:           merchantID,
		BeneficiaryAccountNo: disbursement.BeneficiaryAccountNo,
		BeneficiaryBankCode:  disbursement.BeneficiaryBankCode,
		AdditionalInfo:       map[string]any{},
	})
	if err != nil {
		// Begin Tx
		ctx, errTrx := s.disbursementRepo.BeginTransaction(ctx)
		if errTrx != nil {
			return errTrx
		}

		isCompleted := false
		defer func() {
			if !isCompleted {
				if errTrx = s.disbursementRepo.RollbackTransaction(ctx); errTrx != nil {
					return
				}
			}
		}()

		// Update transaction with failed status
		reasonType := constant.ReasonTypeBeneficiaryAccountReason
		reasonDesc := constant.ReasonDescInvalidBeneficiaryAccount
		if errors.Is(err, constant.ErrInactiveAccount) {
			reasonDesc = constant.SnapCoreResponseInactiveAccountMessage
		} else if errors.Is(err, constant.ErrDormantAccount) {
			reasonDesc = constant.SnapCoreResponseDormantAccountMessage
		} else if errors.Is(err, constant.ErrInvalidAccount) {
			reasonDesc = constant.SnapCoreResponseInvalidAccountMessage
		}

		if errUpdate := s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, constant.StatusFailed, &reasonType, &reasonDesc, disbursement.UUID); errUpdate != nil {
			return errUpdate
		}

		// Commit Tx
		if errTrx = s.disbursementRepo.CommitTransaction(ctx); errTrx != nil {
			return errTrx
		}
		isCompleted = true

		return err
	}

	errBeneficiaryLimit := s.validateBeneficiaryLimit(ctx, merchantID, disbursement, beneficiaryAccount)
	if errBeneficiaryLimit == nil {
		return nil
	}

	// Begin Tx
	ctx, errTrx := s.disbursementRepo.BeginTransaction(ctx)
	if errTrx != nil {
		return errTrx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if errTrx = s.disbursementRepo.RollbackTransaction(ctx); errTrx != nil {
				return
			}
		}
	}()

	// Update transaction with failed status
	reasonType := constant.ReasonTypeDeclinedBeneficiaryRestriction
	reasonDesc := constant.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions

	if errUpdate := s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, constant.StatusFailed, &reasonType, &reasonDesc, disbursement.UUID); errUpdate != nil {
		return errUpdate
	}

	// Commit Tx
	if errTrx = s.disbursementRepo.CommitTransaction(ctx); errTrx != nil {
		return errTrx
	}
	isCompleted = true

	return errBeneficiaryLimit
}

func (s *DisbursementService) ProcessBankTransferAndUpdateTransaction(
	ctx context.Context,
	disbursement *disbursementModel.DisbursementWithTransaction,
	orchestratorTransaction *orchestratorModel.TransactionAndFeeObject,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/processBankTransferAndUpdateTransaction")
	defer segment.End()

	// check forbidden access
	if err := s.merchantForbiddenUsecaseSvc.CheckUseCase(ctx, disbursement.MerchantID, constant.ReferenceDisbursement); err != nil {
		s.logger.Error(ctx, "unable to process disbursement transaction due to merchant forbidden usecase.", logger.Error(err), logger.Any("merchant_id", disbursement.MerchantID), logger.Any("transaction_ids", orchestratorTransaction))
		reasonType := constant.ReasonTypeBlockedByHarsya
		reasonDesc := fmt.Sprintf(constant.ReasonDescBlockedByHarsya, constant.ReferenceDisbursement)

		trxErr := s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, constant.StatusFailed, &reasonType, &reasonDesc, disbursement.UUID)
		if trxErr != nil {
			s.logger.Error(ctx, "unable to update transaction status", logger.Error(trxErr), logger.Any("merchant_id", disbursement.MerchantID), logger.Any("transaction_ids", orchestratorTransaction))
		}
		_ = s.self.DecrDailyTransactionLimit(ctx, disbursement.MerchantID, disbursement.Amount.InexactFloat64())
		_ = s.self.DecrBeneficiaryPayoutLimit(ctx, disbursement.MerchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.Amount.InexactFloat64())
		return err
	}

	// Check if trx is whitelisted
	var isWhitelisted bool
	err := s.redisExt.Get(ctx, fmt.Sprintf("backend-portal:trx-merchant-whitelist:%s", disbursement.ReferenceID)).Scan(&isWhitelisted)
	if err != nil && !errors.Is(err, redis.Nil) {
		s.logger.Error(ctx, "getting trx merchant whitelist", logger.Error(err))
		// Expected error not do anything
	}

	s.logger.Info(
		ctx,
		"check trx merchant whitelist",
		logger.String("merchant_id", disbursement.MerchantID), logger.String("transaction_id", disbursement.ReferenceID), logger.Bool("is_whitelisted", isWhitelisted),
	)

	// if isWhitelisted, then do not process to processor, and make it failed
	if isWhitelisted {
		// Begin Tx
		ctx, trxErr := s.disbursementRepo.BeginTransaction(ctx)
		if trxErr != nil {
			return trxErr
		}

		isCompleted, isFailedTrx := false, false
		defer func() {
			if isCompleted && isFailedTrx {
				_ = s.self.DecrDailyTransactionLimit(context.Background(), disbursement.MerchantID, disbursement.Amount.InexactFloat64())
				_ = s.self.DecrBeneficiaryPayoutLimit(context.Background(), disbursement.MerchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.Amount.InexactFloat64())

				// Send failed alert
				s.sendPayoutTransactionAlert(context.Background(), &disbursementModel.PayoutTransactionAlertRequest{
					DisbursementID: disbursement.UUID,
					BankProcessor:  "",
					TransferType:   "",
					BankRefNo:      "",
				})
			}
		}()
		defer func() {
			if !isCompleted {
				if trxErr = s.disbursementRepo.RollbackTransaction(ctx); trxErr != nil {
					return
				}
			}
		}()

		reasonType := constant.ReasonTypeBlockedByBankWhitelisted
		reasonDesc := constant.ReasonDescBlockedByBankWhitelisted
		if err := s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, constant.StatusFailed, &reasonType, &reasonDesc, disbursement.UUID); err != nil {
			s.logger.Error(ctx, "unable to update transaction status", logger.Error(err), logger.Any("merchant_id", disbursement.MerchantID), logger.Any("transaction_ids", orchestratorTransaction))
			return err
		}

		err := s.redisExt.Del(ctx, fmt.Sprintf("backend-portal:trx-merchant-whitelist:%s", disbursement.ReferenceID)).Err()
		if err != nil {
			s.logger.Error(ctx, "delete exclusive queue with key "+fmt.Sprintf("backend-portal:trx-merchant-whitelist:%s", disbursement.ReferenceID), logger.Error(err))
		}

		// Commit Tx
		if trxErr = s.disbursementRepo.CommitTransaction(ctx); trxErr != nil {
			return trxErr
		}
		isCompleted, isFailedTrx = true, true
		return nil
	}

	if err := s.self.ExternalFDS(ctx, disbursement, orchestratorTransaction); err != nil {
		// Restore limit deducted due to unprocessed bank transfer transaction.
		_ = s.self.DecrDailyTransactionLimit(context.Background(), disbursement.MerchantID, disbursement.Amount.InexactFloat64())
		_ = s.self.DecrBeneficiaryPayoutLimit(context.Background(), disbursement.MerchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.Amount.InexactFloat64())
		// If the external FDS returns an error indicating that the transaction is blocked by FDS, this condition is not treated as an error response.
		// This is because the transaction has been successfully processed; however, the FDS assessment indicates that the transaction is considered high risk for further processing.
		if errors.Is(err, constant.ErrBlockedByFDS) {
			return nil
		}
		return err
	}

	// Request fund transfer to snap-core
	transactionRemark := util.ValueOfPtr(disbursement.Remark)
	// when the remark from merchant was empty
	// then replace with last 12 digit of disbursement uuid
	if transactionRemark == "" {
		transactionRemark = disbursement.UUID[24:]
	}

	// Find the source from bank account by merchant id
	bankAccount, err := s.bankAccountRepo.GetByMerchantID(ctx, disbursement.MerchantID)
	sourceAccount := routingProcessorModel.SubjectRequest{}
	if err != nil || bankAccount == nil {
		sourceAccount.Name = disbursement.SenderName
	} else {
		sourceAccount.Name = bankAccount.BeneficiaryAccountName
		sourceAccount.AccountNo = bankAccount.BeneficiaryAccountNo
		sourceAccount.BankCode = bankAccount.BeneficiaryBankCode
	}

	// Check if manual processing account exists for this merchant
	isManual, err := s.handleManualProcessingAccount(ctx, disbursement, orchestratorTransaction)
	if err != nil {
		return err
	} else if isManual {
		return nil
	}

	triggerTransfer, err := s.routingProcessorSvc.BankTransfer(ctx, &routingProcessorModel.BankTransferRequest{
		HeaderRequest: snapCoreModel.BankTransferHeaderRequest{
			ExternalId: orchestratorTransaction.TransactionID,
			MerchantId: disbursement.MerchantID,
		},
		Beneficiary: routingProcessorModel.SubjectRequest{
			BankCode:    disbursement.BeneficiaryBankCode,
			AccountNo:   disbursement.BeneficiaryAccountNo,
			AccountName: disbursement.BeneficiaryAccountName,
		},
		Source:               sourceAccount,
		Amount:               commonModel.Amount{Currency: disbursement.Currency, Value: disbursement.Amount.StringFixed(2)},
		Currency:             disbursement.Currency,
		Remark:               transactionRemark,
		PurposeOfTransaction: snapCoreModel.DefaultPurchaseOfTransaction,
		TransactionDate:      disbursement.CreatedAt,
	})
	if err != nil && errors.Is(err, constant.ErrDoubleDisbursementIndication) {
		return err
	}

	// Begin Tx
	ctx, trxErr := s.disbursementRepo.BeginTransaction(ctx)
	if trxErr != nil {
		return trxErr
	}

	isCompleted, isFailedTrx, isPendingTrx := false, false, false
	bankProcessor, transferType, bankRefNo := "", "", ""
	defer func() {
		if isCompleted {
			if isFailedTrx {
				_ = s.self.DecrDailyTransactionLimit(context.Background(), disbursement.MerchantID, disbursement.Amount.InexactFloat64())
				_ = s.self.DecrBeneficiaryPayoutLimit(context.Background(), disbursement.MerchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo, disbursement.Amount.InexactFloat64())

				// Send failed alert
				s.sendPayoutTransactionAlert(context.Background(), &disbursementModel.PayoutTransactionAlertRequest{
					DisbursementID: disbursement.UUID,
					BankProcessor:  bankProcessor,
					TransferType:   transferType,
					BankRefNo:      bankRefNo,
				})
			}

			if isPendingTrx {
				// Send pending alert
				errPub := s.rabbitMqExt.PublishWithTTL(
					context.Background(),
					rabbitMqExt.PayoutAlertProcessingRoutingKey,
					&disbursementModel.PayoutTransactionAlertRequest{
						DisbursementID: disbursement.UUID,
						BankProcessor:  bankProcessor,
						TransferType:   transferType,
						BankRefNo:      bankRefNo,
					},
					time.Duration(s.config.DisbursementConfig.SchedulePayoutAlertInMinute)*time.Minute,
				)
				if errPub != nil {
					s.logger.Error(ctx, "error publish payout alert", logger.Error(errPub))
				}
			}
		}
	}()
	defer func() {
		if !isCompleted {
			if trxErr = s.disbursementRepo.RollbackTransaction(ctx); trxErr != nil {
				return
			}
		}
	}()

	if err != nil {
		// Update transaction and fee status
		reasonType := constant.ReasonTypeOtherReason
		reasonDesc := ""
		accountTransactionStatus := constant.StatusFailed
		processorReferenceName := ""

		// Set status to PENDING due to SnapCoreResponseCodeInsufficientFund
		if triggerTransfer != nil {
			processorReferenceName = triggerTransfer.ProcessorReference
			bankProcessor, transferType = triggerTransfer.BankProcessor, triggerTransfer.TransferType

			if triggerTransfer.Status == constant.SnapCoreBankTransferStatusPending {
				accountTransactionStatus = constant.StatusPending
				reasonDesc = triggerTransfer.ResponseMessage
				isPendingTrx = true

			} else if triggerTransfer.Status == constant.SnapCoreBankTransferStatusFailed &&
				util.IsPatternMatch(constant.SnapCoreResponseCodeInsufficientFundPattern, triggerTransfer.ResponseCode) {
				reasonType = constant.ReasonTypeInsufficientEscrowFund
				reasonDesc = triggerTransfer.ResponseMessage
				accountTransactionStatus = constant.StatusPending
			} else if util.IsPatternMatch(constant.SnapCoreResponseCodeInactiveAccountPattern, triggerTransfer.ResponseCode) {
				reasonType = constant.ReasonTypeBeneficiaryAccountReason
				reasonDesc = constant.SnapCoreResponseInactiveAccountMessage
			} else if util.IsPatternMatch(constant.SnapCoreResponseCodeDormantAccountPattern, triggerTransfer.ResponseCode) {
				reasonType = constant.ReasonTypeBeneficiaryAccountReason
				reasonDesc = constant.SnapCoreResponseDormantAccountMessage
			} else if util.IsPatternMatch(constant.SnapCoreResponseCodeInvalidAccountPattern, triggerTransfer.ResponseCode) {
				reasonType = constant.ReasonTypeBeneficiaryAccountReason
				reasonDesc = constant.SnapCoreResponseInvalidAccountMessage
			} else if triggerTransfer.ResponseMessage != "" {
				reasonDesc = triggerTransfer.ResponseMessage
			}
		}

		if errUpdate := s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, accountTransactionStatus, &reasonType, &reasonDesc, disbursement.UUID); errUpdate != nil {
			return errUpdate
		}

		if processorReferenceName != "" {
			if errUpdateProcRef := s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctx, orchestratorTransaction.TransactionID, processorReferenceName, triggerTransfer.UUID, triggerTransfer.GetReconReferenceNo()); errUpdateProcRef != nil {
				return errUpdateProcRef
			}
		}

		// Commit Tx
		if trxErr = s.disbursementRepo.CommitTransaction(ctx); trxErr != nil {
			return trxErr
		}
		isCompleted = true
		isFailedTrx = accountTransactionStatus == constant.StatusFailed
		isPendingTrx = accountTransactionStatus == constant.StatusPending

		return err
	}

	// Update disbursement set processorReferenceId = triggerTransfer.UUID
	if errUpdate := s.disbursementRepo.UpdateProcessorReferenceIdAndBankReferenceNo(ctx, disbursement.UUID, triggerTransfer.UUID, triggerTransfer.BankReferenceNo); errUpdate != nil {
		return errUpdate
	}

	if errUpdateProcRef := s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctx, orchestratorTransaction.TransactionID, triggerTransfer.ProcessorReference, triggerTransfer.UUID, triggerTransfer.GetReconReferenceNo()); errUpdateProcRef != nil {
		return errUpdateProcRef
	}

	if triggerTransfer.TransactionDate.IsZero() {
		triggerTransfer.TransactionDate = time.Now().UTC()
	}
	if errUpdTrxTime := s.updateTransactionTimestamp(ctx, orchestratorTransaction, triggerTransfer.TransactionDate); errUpdTrxTime != nil {
		return errUpdTrxTime
	}

	// Update transaction and fee status
	if disbursement.MetadataObj.FeeDetail.DeductionType == constant.MerchantFeeDeductionTypeAutomated {
		orchestratorTransaction.FeeID = ""
	}

	// Increment deferred LADDER tiering counter on sync success
	if triggerTransfer.Status == constant.StatusSuccess {
		feeDetail := disbursement.MetadataObj.FeeDetail
		if feeDetail.LadderCounterKey != "" {
			s.feeSvc.IncrementLadderCounter(ctx, feeDetail.LadderCounterKey, feeDetail.LadderCounterIncrement)
		}
	}

	if triggerTransfer.Status != constant.StatusPending {
		if errUpdate := s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, triggerTransfer.Status, nil, nil, disbursement.UUID); errUpdate != nil {
			return errUpdate
		}
	}

	// Commit Tx
	if trxErr = s.disbursementRepo.CommitTransaction(ctx); trxErr != nil {
		return trxErr
	}
	isCompleted = true
	isFailedTrx = triggerTransfer.Status == constant.StatusFailed
	isPendingTrx = triggerTransfer.Status == constant.StatusPending

	return nil
}

// handleManualProcessingAccount checks whether the disbursement targets a merchant account flagged
// for manual processing. When flagged, it logs the event and marks the transaction as pending
// (waiting for manual action), and returns handled=true so the caller short-circuits the normal
// bank-transfer flow. Returns handled=false otherwise.
func (s *DisbursementService) handleManualProcessingAccount(
	ctx context.Context,
	disbursement *disbursementModel.DisbursementWithTransaction,
	orchestratorTransaction *orchestratorModel.TransactionAndFeeObject,
) (bool, error) {
	isExists := false
	var err error
	isExists, err = s.payoutManualProcessingAccountRepo.IsManualProcessingAccount(ctx, disbursement.MerchantID, disbursement.BeneficiaryBankCode, disbursement.BeneficiaryAccountNo)
	if err != nil {
		s.logger.Error(ctx, "failed to check manual processing account", logger.Error(err))
		return false, err
	}
	if !isExists {
		return false, nil
	}

	s.logger.Info(
		ctx, fmt.Sprintf("Disbursement ID %s triggered but requires manual action due to whitelist; process will be handled by ops team", disbursement.UUID),
		logger.Any("disbursement", map[string]any{
			"id":            disbursement.UUID,
			"merchantId":    disbursement.MerchantID,
			"bankCode":      disbursement.BeneficiaryBankCode,
			"bankName":      disbursement.BeneficiaryBankName,
			"accountNumber": disbursement.BeneficiaryAccountNo,
			"accountName":   disbursement.BeneficiaryAccountName,
			"currency":      disbursement.Currency,
			"amount":        disbursement.Amount.String(),
		}),
	)

	errUpdate := s.updateTransactionStatus(
		ctx, orchestratorTransaction, constant.StatusPending,
		util.ValueToPtr(constant.ReasonTypeWaitingManualAction),
		util.ValueToPtr(constant.ReasonDescWaitingManualAction),
	)
	if errUpdate != nil {
		return false, errUpdate
	}

	return true, nil
}

func (s *DisbursementService) updateTransactionStatus(
	ctx context.Context,
	orchestratorTransaction *orchestratorModel.TransactionAndFeeObject,
	status string,
	reasonType, reasonDescription *string,
) error {
	return s.updateTransactionStatusWithHistory(ctx, orchestratorTransaction, status, reasonType, reasonDescription, "")
}

func (s *DisbursementService) updateTransactionStatusWithHistory(
	ctx context.Context,
	orchestratorTransaction *orchestratorModel.TransactionAndFeeObject,
	status string,
	reasonType, reasonDescription *string,
	disbursementID string,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/updateTransactionStatusWithHistory")
	defer segment.End()

	// Update transaction status
	if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, orchestratorTransaction.TransactionID, status, reasonType, reasonDescription); errUpdate != nil {
		return errUpdate
	}

	// Record status history if disbursementID is provided
	if disbursementID != "" {
		switch status {
		case constant.StatusPending:
			transactionReasonType := ""
			if reasonType != nil {
				transactionReasonType = *reasonType
			}
			s.recordDisbursementProcessing(ctx, disbursementID, s.determineProcessingReasonType(transactionReasonType))
		case constant.StatusSuccess:
			s.recordDisbursementSuccess(ctx, disbursementID)
		case constant.StatusFailed:
			// Get disbursement object to determine failure reason type
			transactionReasonType, transactionReasonDesc := "", ""
			if reasonType != nil {
				transactionReasonType = *reasonType
			}
			if reasonDescription != nil {
				transactionReasonDesc = *reasonDescription
			}
			s.recordDisbursementFailed(ctx, disbursementID, s.determineFailureReasonType(transactionReasonType, transactionReasonDesc))
		}
	}

	if orchestratorTransaction.FeeID == "" {
		return nil
	}

	// Update fee transaction status
	if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, orchestratorTransaction.FeeID, status, reasonType, reasonDescription); errUpdate != nil {
		return errUpdate
	}

	if orchestratorTransaction.TransferFeeID == "" {
		return nil
	}

	request := &ledgerModel.UpdateLedgerEntryRequest{
		ReferenceID: util.ParseUUID(orchestratorTransaction.TransferFeeID),
		Usecase:     constant.ReferenceDisbursement,
		Status:      status,
		ReasonType:  "null", ReasonDescription: "null",
	}
	if reasonType != nil && *reasonType != "" {
		request.ReasonType = constant.ReasonTypeOtherReason
	}
	if reasonDescription != nil && *reasonDescription != "" {
		request.ReasonDescription = "Payout: " + *reasonDescription
	}

	if err := s.ledgerSvc.UpdateTransaction(ctx, request); err != nil {
		return err
	}

	var trfReasonDesc *string
	if request.ReasonDescription != "" && request.ReasonDescription != "null" {
		trfReasonDesc = &request.ReasonDescription
	}
	return s.transferSvc.UpdateTransferStatus(
		ctx, orchestratorTransaction.MerchantID, orchestratorTransaction.TransferFeeID, status, trfReasonDesc,
	)
}

// determineFailureReasonType maps failure reasons to specific reason types
func (s *DisbursementService) determineFailureReasonType(transactionReasonType, transactionReasonDesc string) string {
	switch transactionReasonType {
	case constant.ReasonTypeDeclinedBeneficiaryRestriction:
		return constant.DisbursementReasonTypeBeneficiaryLimit
	case constant.ReasonTypeBeneficiaryAccountReason:
		if strings.Contains(transactionReasonDesc, "inactive account") {
			return constant.DisbursementReasonTypeInactiveAccount
		} else if strings.Contains(transactionReasonDesc, "dormant account") {
			return constant.DisbursementReasonTypeDormantAccount
		}

		return constant.DisbursementReasonTypeInvalidAccount
	case constant.ReasonTypeBlockedByHarsya:
		return constant.DisbursementReasonTypeFeatureUnavailable
	case constant.ReasonTypeBlockedByFDS:
		return constant.DisbursementReasonTypeBlockedByFDS
	}

	return constant.DisbursementReasonTypeOther
}

// determineProcessingReasonType maps processing reasons to specific reason types
func (s *DisbursementService) determineProcessingReasonType(transactionReasonType string) string {
	if transactionReasonType == constant.ReasonTypePayoutDelayed {
		return constant.DisbursementReasonTypeDelayed
	}
	if transactionReasonType == constant.ReasonTypePayoutCutOffTime {
		return constant.DisbursementReasonTypeCutOffTime
	}
	return constant.DisbursementReasonTypeOther
}

func (s *DisbursementService) updateTransactionTimestamp(
	ctx context.Context,
	orchestratorTransaction *orchestratorModel.TransactionAndFeeObject,
	transactionTimestamp time.Time,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/updateTransactionTimestamp")
	defer segment.End()

	// Update transaction status
	if errUpdate := s.orchestratorSvc.UpdateTransactionTimestamp(ctx, orchestratorTransaction.TransactionID, transactionTimestamp); errUpdate != nil {
		return errUpdate
	}

	if orchestratorTransaction.FeeID == "" {
		return nil
	}

	return s.orchestratorSvc.UpdateTransactionTimestamp(ctx, orchestratorTransaction.FeeID, transactionTimestamp)
}

func (s *DisbursementService) CreateBankTransfer(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ProcessBankTransfer")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        "Disbursement-Service",
		ReferenceId: disbursement.MerchantID,
		OriginId:    disbursement.UUID,
	})

	_, err := util.ValidateMagicNumber(http.MethodGet, disbursement.BeneficiaryAccountNo)
	if s.config.Environment != constant.EnvironmentProduction && err != nil && err.Error() == constant.ErrInsufficientBalance.Error() {
		ctx = context.WithValue(ctx, constant.CtxMockInsufficientBalanceMerchant, true)
	}

	orchestratorTransaction := &orchestratorModel.TransactionAndFeeObject{
		MerchantID: disbursement.MerchantID,
	}

	disbursementLedger, err := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, constant.TypeDisbursement)
	if err != nil {
		return pkgErrors.New(httpResponse.HttpErrDatabase, err)
	} else if disbursementLedger == nil {
		if s.ffBankTransferCreatePendingTrxIfNotExists(ctx, s.config.Environment) {

			transactionId, _, err := s.self.CreatePendingOrchestratorTransaction(ctx, disbursement)
			if err != nil {
				return pkgErrors.New(httpResponse.HttpErrDatabase, err)
			}
			disbursementLedger = &orchestratorModel.AccountTransactionWithUseCase{
				UUID: util.ParseUUID(transactionId),
			}

		} else {
			return pkgErrors.New(httpResponse.HttpErrDatabase, constant.ErrLedgerDetailNotFound)
		}
	} else if disbursementLedger.Status != constant.StatusPending {
		s.logger.Info(ctx, "Transaction already processed", logger.Any("accountTransaction", disbursementLedger))
		return nil
	}
	orchestratorTransaction.TransactionID = disbursementLedger.UUID.String()

	feeLedger, err := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, constant.TypeFee)
	if err != nil {
		return pkgErrors.New(httpResponse.HttpErrDatabase, err)
	} else if feeLedger != nil {
		orchestratorTransaction.FeeID = feeLedger.UUID.String()

		feeMetadata := feeModel.FeeMetadataObject{}
		_ = json.Unmarshal(feeLedger.AdditionalInfo.JSONText, &feeMetadata)

		orchestratorTransaction.TransferFeeID = feeMetadata.TransferId
	}

	return s.self.ProcessBankTransferAndUpdateTransaction(ctx, disbursement, orchestratorTransaction)
}

func (s *DisbursementService) ffBankTransferCreatePendingTrxIfNotExists(ctx context.Context, env string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ffBankTransferCreatePendingTrxIfNotExists")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(env)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, env)

	result, _ := ffclient.BoolVariation(constant.FeatureFlagKeyBankTransferCreatePendingTrxIfNotExists, attr, false)
	return result
}
