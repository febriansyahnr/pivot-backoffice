package cardFundedPayoutService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/go-redsync/redsync/v4"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const maxPaymentCreatedDays = 14

func (s *service) ProcessFinishCardFundedPayoutSettlement(ctx context.Context, request *model.ProcessFinishCardFundedPayoutSettlementRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/ProcessFinishCardFundedPayoutSettlement")
	defer span.End()

	s.logger.Info(ctx, "Checking transaction settlement for payment ID "+request.ReferenceID)

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed get payment details: %w", err)
	}

	if payment.Type != constant.PaymentTypeCardFundedPayout {
		return errors.New("payment is not a card-funded payout")
	}

	lockKey := fmt.Sprintf(constant.LockKeyCardFundedPayoutFinishSettlementProcess, *payment.ReferenceID)

	mutex := s.cacheClient.NewMutex(
		lockKey,
		redsync.WithExpiry(60*time.Second),
		redsync.WithRetryDelay(80*time.Millisecond),
		redsync.WithFailFast(true),
		redsync.WithTries(256),
	)
	if err = mutex.LockContext(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer func() {
		if _, unlockErr := mutex.UnlockContext(ctx); unlockErr != nil {
			s.logger.Error(ctx, "failed to release distributed lock for card-funded payout", logger.Error(unlockErr))
		}
	}()

	fundingSummary, err := s.paymentRepo.GetCardFundedPayoutFundingSummary(ctx, payment.MerchantID, *payment.ReferenceID, maxPaymentCreatedDays)
	if err != nil {
		return fmt.Errorf("failed get funding summary: %w", err)
	}

	if fundingSummary.TotalPayment == 0 {
		return errors.New("total payment amount is zero")

	} else if fundingSummary.TotalFailed+fundingSummary.TotalSuccessSettlement != fundingSummary.TotalPayment {
		s.logger.Info(ctx, fmt.Sprintf("Cannot process card-funded payout transaction for ID %s: related transactions are not finalized", fundingSummary.PayoutID), logger.Any("detail", fundingSummary))
		return nil
	}

	s.logger.Info(ctx, "Card-funded payout funding summary for payout ID "+fundingSummary.PayoutID, logger.Any("detail", fundingSummary))

	payout, err := s.disbursementRepo.GetDetailForCardFundedPayoutByID(ctx, *payment.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed get payout detail: %w", err)
	}

	_ = s.recordStatusHistory(ctx, payout.UUID, constant.DisbursementStatusHistoryProcessing, constant.StatusHistoryActorSystem, "")

	var (
		now           = time.Now().UTC()
		transactionID = util.GenerateUUID()
		payoutAmount  = fundingSummary.TotalSuccessSettlement - fundingSummary.TotalFee

		accTrxStatus, reasonType, reasonDesc, errMessage = constant.StatusPending, "", "", ""
	)

	orchestratorRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 transactionID,
		ReferenceID:          payout.UUID,
		Type:                 orchestratorModel.TypeDisbursement,
		MerchantID:           util.ParseUUID(payout.MerchantID),
		Currency:             payout.Currency,
		Debit:                payoutAmount,
		Channel:              constant.ChannelBankTransfer,
		Status:               accTrxStatus,
		Remarks:              util.ValueOfPtr(payout.Remark),
		TransactionTimestamp: now,
		Usecase:              constant.TypePaymentFundedPayout,
	}

	isManualProcessingAccount := constant.IsCardFundedPayoutManualProcessingAccount(payout.BeneficiaryBankCode, payout.BeneficiaryAccountNo)
	if isManualProcessingAccount {
		orchestratorRequest.ReasonType = util.ValueToPtr(constant.ReasonTypeWaitingManualAction)
		orchestratorRequest.ReasonDescription = util.ValueToPtr(constant.ReasonDescWaitingManualAction)
		orchestratorRequest.Processor = constant.ManualProcessor
	}
	if err = s.orchestratorSvc.PostAccountTransaction(ctx, orchestratorRequest); err != nil {
		return fmt.Errorf("failed post account transaction: %w", err)
	}

	if isManualProcessingAccount {
		s.logger.Info(
			ctx, fmt.Sprintf("Payout ID %s triggered but requires manual action due to whitelist; process will be handled by ops team", payout.UUID),
			logger.Any("payout", map[string]any{
				"id":            payout.UUID,
				"merchantId":    payout.MerchantID,
				"bankCode":      payout.BeneficiaryBankCode,
				"bankName":      payout.BeneficiaryBankName,
				"accountNumber": payout.BeneficiaryAccountNo,
				"accountName":   payout.BeneficiaryAccountName,
				"currency":      payout.Currency,
				"amount":        fmt.Sprintf("%.2f", payoutAmount),
			}),
		)
		return nil
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        "Card-Funded-Payout",
		OriginId:    payout.UUID,
		ReferenceId: payout.MerchantID,
	})

	snapCoreResp, err := s.snapCoreRepo.BankTransfer(ctx, &snapCoreModel.BankTransferRequest{
		BTBeneficiaryRequest: snapCoreModel.BTBeneficiaryRequest{
			BeneficiaryBankCode:    payout.BeneficiaryBankCode,
			BeneficiaryAccountNo:   payout.BeneficiaryAccountNo,
			BeneficiaryAccountName: payout.BeneficiaryAccountName,
		},
		Currency: payout.Currency,
		Amount: commonModel.Amount{
			Currency: payout.Currency,
			Value:    fmt.Sprintf("%.f", payoutAmount),
		},
		Remark:               util.ValueOfPtr(payout.Remark),
		PurposeOfTransaction: snapCoreModel.DefaultPurchaseOfTransaction,
		TransactionDate:      now,
	}, &snapCoreModel.BankTransferHeaderRequest{
		ExternalId: transactionID.String(),
		MerchantId: payout.MerchantID,
	})
	if snapCoreResp != nil && snapCoreResp.UUID != "" {
		if updateErr := s.disbursementRepo.UpdateProcessorReferenceIdAndBankReferenceNo(ctx, payout.UUID, snapCoreResp.UUID, snapCoreResp.BankReferenceNo); updateErr != nil {
			s.logger.Error(ctx, "Update payout processor reference id and bank reference no", logger.Error(updateErr))
		}
		if updateErr := s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctx, transactionID.String(), constant.SnapCoreProcessor, snapCoreResp.UUID, snapCoreResp.GetReconReferenceNo()); updateErr != nil {
			s.logger.Error(ctx, "Update account transactions additional info", logger.Error(updateErr))
		}
	}
	if err != nil {
		accTrxStatus, reasonType, reasonDesc, errMessage = constant.StatusFailed, constant.ReasonTypeOtherReason, "", err.Error()
		if snapCoreResp != nil {
			accTrxStatus, reasonType, reasonDesc = snapCoreResp.MappingAccountTransactionErrStatus()
		}
		if updateErr := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, transactionID.String(), accTrxStatus, &reasonType, &reasonDesc); updateErr != nil {
			s.logger.Error(ctx, "Update status account transaction (failed)", logger.Error(updateErr))
		}

	} else if snapCoreResp != nil && snapCoreResp.Status == constant.SnapCoreBankTransferStatusSuccess {
		if updateErr := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, transactionID.String(), constant.StatusSuccess, nil, nil); updateErr == nil {
			accTrxStatus = constant.StatusSuccess
		} else {
			s.logger.Error(ctx, "Update status account transaction (success)", logger.Error(updateErr))
		}
	}

	_ = s.recordStatusHistory(ctx, payout.UUID, accTrxStatus, constant.StatusHistoryActorSystem, "")

	s.logger.Info(
		ctx, "Bank transfer status for card-funded payout transaction",
		logger.Any("details", map[string]string{
			"id":                     payout.UUID,
			"merchantId":             payout.MerchantID,
			"beneficiaryBankName":    util.ValueOfPtr(payout.BeneficiaryBankName),
			"beneficiaryAccountNo":   payout.BeneficiaryAccountNo,
			"beneficiaryAccountName": payout.BeneficiaryAccountName,
			"currency":               payout.Currency,
			"amount":                 fmt.Sprintf("%.2f", payoutAmount),
		}),
		logger.String("status", accTrxStatus), logger.String("reasonType", reasonType), logger.String("reasonDescription", reasonDesc), logger.String("error", errMessage),
	)

	return nil
}
