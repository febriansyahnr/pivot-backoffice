package refundProcessorService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func (s *EWalletStrategy) Process(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refundProcessor/EWalletStrategyProcess")
	defer span.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        "Refund",
		OriginId:    request.RefundID,
		ReferenceId: request.MerchantID,
	})

	payload := &ewallet.EWalletRefundRequest{
		TransactionID: request.PaymentProcessorID,
		Amount: commonModel.Amount{
			Currency: request.Currency,
			Value:    fmt.Sprintf("%.2f", request.Amount),
		},
	}
	resp, err := s.snapCoreRepo.RefundEWallet(ctx, payload)
	if err != nil {
		s.logger.Error(ctx, "[EWalletStrategyProcess] failed to refund EWallet", pdkLogger.Error(err))
		return err
	}

	if resp == nil {
		s.logger.Error(ctx, "[EWalletStrategyProcess] failed to refund EWallet due nil response from processor")
		return constant.ErrFailedToRefundPayment
	}

	s.logger.Info(ctx, "success refund EWallet",
		pdkLogger.String("refundId", request.RefundID),
		pdkLogger.String("paymentProcessorId", request.PaymentProcessorID),
		pdkLogger.String("refundNo", resp.RefundNo),
	)

	return nil
}
