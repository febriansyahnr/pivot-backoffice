package cardFundedPayoutService

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) ProcessPendingSubsequentPayments(ctx context.Context, merchantID, referenceID string) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/ProcessPendingSubsequentPayments")
	defer span.End()

	payout, err := s.disbursementRepo.GetDetailForCardFundedPayoutByID(ctx, referenceID)
	if err != nil {
		s.logger.Error(ctx, "Failed when get card-funded payout detail", logger.Error(err))
		return err
	}
	cardFundedPayout := payout.GetCardFundedPayoutDetail()

	request := model.ExecuteSubsequentPaymentRequest{
		MerchantID:  merchantID,
		ReferenceID: referenceID,
		VendorID:    cardFundedPayout.VendorID,
		VendorName:  cardFundedPayout.VendorName,
	}
	if cardFundedPayout.SettlementMethod == constant.PaymentSettlementMethodStandard {
		go s.executeSubsequentPayments(context.WithoutCancel(ctx), request)
	} else {
		// Currently, the INSTANT settlement method is not yet available.
		//
		// Once this feature is enabled, after executing the subsequent transaction, the flow will proceed to
		// payment transaction validation followed by the payout process.
		return s.executeSubsequentPayments(ctx, request)
	}
	return nil
}

func (s *service) executeSubsequentPayments(ctx context.Context, request model.ExecuteSubsequentPaymentRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/executeSubsequentPayments")
	defer span.End()

	payments, err := s.paymentRepo.FindPendingSubsequentCardFundedPayout(ctx, request.MerchantID, request.ReferenceID)
	if err != nil {
		s.logger.Error(ctx, "Failed when finding pending subsequent card-funded payment transactions", logger.Error(err))
		return err
	} else if len(payments) == 0 {
		s.logger.Info(ctx, fmt.Sprintf("Transaction with reference ID %s has no subsequent transactions", request.ReferenceID))
		return nil
	}

	certPEM, err := s.creditCardSvc.GetCardEncryptionPublicKey(ctx, request.MerchantID)
	if err != nil {
		return err
	}

	const workerCount = 5

	wg := new(sync.WaitGroup)
	pool, err := ants.NewPoolWithFuncGeneric(workerCount, func(payment model.CardFundedPayment) {
		defer wg.Done()

		fnCtx := context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
			From:        "Card-Funded-Payout",
			ReferenceId: request.MerchantID,
			OriginId:    payment.ID,
		})
		s.logger.Info(fnCtx, "Processing subsequent payment for payment ID "+payment.ID, logger.Any("detail", payment))

		payloadBytes, _ := json.Marshal(payment.ToCardAuthenticationRequest(request.VendorID, request.VendorName))

		encryptedPayload, encryptErr := s.cryptoProvider.EncryptPKCS7(certPEM, payloadBytes)
		if encryptErr != nil {
			s.logger.Error(fnCtx, "Failed when encrypt request payload for card authentication", logger.Error(encryptErr))
			return
		}

		authRequest := creditcardCoreProcessorModel.AuthenticationRequest{
			MerchantID:       payment.MerchantID,
			PaymentID:        payment.ID,
			EncryptedPayload: encryptedPayload,
		}
		if _, authErr := s.creditCardSvc.Authentication(fnCtx, authRequest); authErr != nil {
			s.failPaymentAndChargeOnError(
				fnCtx,
				payment.ID,
				payment.ChargeID,
				constant.CreditCardAuthorizationFailed,
				"Failed to process authorization request",
			)
			s.logger.Error(fnCtx, "Failed when execute card authentication request", logger.Error(authErr))
		}
	})
	if err != nil {
		s.logger.Error(ctx, "Failed when preparing worker pool for pending subsequent payments", logger.Error(err))
		return err
	}
	defer pool.Release()

	for _, payment := range payments {
		wg.Add(1)
		pool.Invoke(payment)
	}

	wg.Wait()

	return nil
}

func (s *service) failPaymentAndChargeOnError(ctx context.Context, paymentID, chargeID, reasonType, reasonDesc string) {
	defer func() { _ = s.recordPaymentFailedStatus(ctx, paymentID) }()

	err := s.paymentRepo.UpdatePaymentStatusWithReason(ctx, paymentID, paymentModel.UpdatePaymentStatusWithReasonRequest{
		Status:            constant.UnifiedPaymentSessionStatusCancelled,
		ReasonType:        util.ValueToPtr(reasonType),
		ReasonDescription: util.ValueToPtr(reasonDesc),
	})
	if err != nil {
		s.logger.Warn(ctx, "Failed to update payment status with failure reason", logger.Error(err))
		return
	}
	err = s.accountTransactionRepo.UpdateStatusAccountTransaction(ctx, chargeID, constant.StatusFailed, nil, nil)
	if err != nil {
		s.logger.Warn(ctx, "Failed to update charge status", logger.Error(err))
		return
	}
}
