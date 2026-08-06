package settlementService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *settlementService) ProcessSettlementHoldOrRelease(ctx context.Context, request *settlementModel.ProcessHoldReleaseSettlementRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/settlement/ProcessSettlementHoldOrRelease")
	defer segment.End()

	payment, err := s.paymentSvc.GetDetailByID(ctx, request.ReferenceID)
	if err != nil {
		s.logger.Error(ctx, "error retrieve payment id", logger.Error(err))
		return err
	}

	isOnHold := false
	if request.Action == constant.SettlementHoldActionHold {
		isOnHold = true
	}
	err = s.accountTransactionRepo.UpdateSettlementHoldByReferenceID(ctx, payment.UUID, isOnHold)
	if err != nil {
		s.logger.Error(ctx, "error when update ledger settlement hold by reference id", logger.Error(err), logger.String("paymentId", request.ReferenceID), logger.String("action", request.Action))
		return err
	}

	errList := []error{}
	if request.Action == constant.SettlementHoldActionRelease && !request.LastActionTime.IsZero() {
		unsettledTrxList, err := s.accountTransactionRepo.GetPastDueSettlementTransactions(ctx, &orchestrator_model.GetPastDueSettlementTransactionsRequest{
			ReferenceID: request.ReferenceID,
			Datetime:    request.LastActionTime,
		})
		if err != nil {
			s.logger.Error(ctx, "error when retrieve past due settlement transaction", logger.Error(err))
			return err
		}
		for _, unsettledTrx := range unsettledTrxList {
			if unsettledTrx.Type == constant.TypeFee {
				err = s.internalSvc.ProcessSettlementTransactionFee(ctx, payment.MerchantID, unsettledTrx.UUID.String())

			} else {
				err = s.internalSvc.ProcessSettlement(ctx, &settlementModel.ProcessSettlementRequest{
					MerchantID:    payment.MerchantID,
					TransactionID: unsettledTrx.UUID.String(),
				})
			}
			if err != nil {
				s.logger.Error(ctx, "error when settle unsettled transaction", logger.Error(err), logger.String("transactionId", unsettledTrx.UUID.String()), logger.String("transactionType", unsettledTrx.Type))
				errList = append(errList, err)
			}
		}

		if len(errList) > 0 {
			return constant.ErrProcessManualSettlement
		}
	}

	return nil
}
