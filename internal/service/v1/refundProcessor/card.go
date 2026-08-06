package refundProcessorService

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (s *CardStrategy) Process(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refundProcessor/CardStrategyProcess")
	defer span.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        "Refund",
		OriginId:    request.RefundID,
		ReferenceId: request.MerchantID,
	})

	refundResp, err := s.creditCardRepo.Refund(ctx, &creditcardCoreProcessorModel.RefundRequest{
		MerchantID:              request.MerchantID,
		ClientTransactionID:     request.PaymentClientReferenceID,
		AcquirerTransactionID:   request.PaymentProcessorID,
		RefundClientReferenceID: request.ClientReferenceID,
		Currency:                request.Currency,
		Amount:                  request.Amount,
	})
	if err != nil {
		return err
	} else if refundResp == nil {
		s.logger.Warn(ctx, "[CardStrategyProcess] Refund response data empty")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	if refundResp.Status != constant.CreditCardStatusSuccess {
		s.logger.Warn(ctx, "[CardStrategyProcess] Refund response status is not success")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrFailedToRefundPayment)
	}

	return nil
}
