package cardFundedPayoutService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) ProcessInitialCardFundedPayoutAuthFailure(ctx context.Context, merchantID, referenceID string) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/ProcessInitialCardFundedPayoutAuthFailure")
	defer span.End()

	// Get payout detail
	payout, err := s.disbursementRepo.GetDetailForCardFundedPayoutByID(ctx, referenceID)
	if err != nil {
		s.logger.Error(ctx, "Failed when get card-funded payout detail", logger.Error(err))
		return err
	} else if payout == nil {
		s.logger.Warn(ctx, "No data found when get card-funded payout detail", logger.String("referenceID", referenceID))
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	// Find and fail all pending subsequent payment
	payments, err := s.paymentRepo.FindPendingSubsequentCardFundedPayout(ctx, merchantID, referenceID)
	if err != nil {
		s.logger.Error(ctx, "Failed when finding pending subsequent card-funded payment transactions", logger.Error(err))
		return err
	}

	// Create payout ledger with failed status
	orchestratorRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 util.GenerateUUID(),
		ReferenceID:          payout.UUID,
		Type:                 orchestratorModel.TypeDisbursement,
		MerchantID:           util.ParseUUID(payout.MerchantID),
		Currency:             payout.Currency,
		Debit:                payout.Amount.InexactFloat64(),
		Channel:              constant.ChannelBankTransfer,
		Status:               constant.TransferStatusFailed,
		Remarks:              util.ValueOfPtr(payout.Remark),
		ReasonType:           util.ValueToPtr(constant.ReasonTypeOtherReason),
		ReasonDescription:    util.ValueToPtr("Failed to process the payout due to a charge failure"),
		TransactionTimestamp: time.Now().UTC(),
		Usecase:              constant.TypePaymentFundedPayout,
	}
	if err = s.orchestratorSvc.PostAccountTransaction(ctx, orchestratorRequest); err != nil {
		return fmt.Errorf("failed post account transaction: %w", err)
	}

	// Record failed payout history
	_ = s.recordStatusHistory(ctx, payout.UUID, constant.DisbursementStatusHistoryFailed, constant.StatusHistoryActorSystem, "")

	// Cancel subsequent payment if any
	if len(payments) > 0 {
		s.logger.Info(ctx, fmt.Sprintf("Transaction with reference ID %s has subsequent transactions, all transactions will be canceled", referenceID))
		for _, payment := range payments {
			s.failPaymentAndChargeOnError(
				ctx,
				payment.ID,
				payment.ChargeID,
				constant.CreditCardAuthenticationFailed,
				"Failed to process due to a first-payment charge failure",
			)
		}
	}

	return nil
}
