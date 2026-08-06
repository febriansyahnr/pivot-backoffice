package refundProcessorService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	snapQReModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

// Process is responsible for handling the refund process for QRIS payments.
// It initiates a refund request to the snap core repository with the payment processor ID and reason.
// The method uses OpenTelemetry for tracing and logs the refund operations with appropriate log levels.
// this func will ignore the acquirer and give the responsibility to the snap core
// to handle the refund process
func (s *QRISStrategy) Process(ctx context.Context, request *refundModel.RefundProcessRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refundProcessor/QRISStrategyProcess")
	defer span.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        "Refund",
		OriginId:    request.RefundID,
		ReferenceId: request.MerchantID,
	})

	payload := &snapQReModel.QRMPMRefundRequest{
		QRID:   request.PaymentProcessorID,
		Reason: request.Reason,
		AdditionalInfo: map[string]interface{}{
			"clientReferenceId": request.ClientReferenceID,
		},
		Amount: commonModel.Amount{
			Currency: request.Currency,
			Value:    fmt.Sprintf("%.2f", request.Amount),
		},
	}
	resp, err := s.snapCoreRepo.RefundQRMPM(ctx, payload)
	if err != nil {
		s.logger.Error(ctx, "[QRISStrategyProcess] failed to refund QRIS", pdkLogger.Error(err))
		return err
	}

	if resp == nil {
		s.logger.Error(ctx, "[QRISStrategyProcess] failed to refund QRIS due nil response from processor")
		return constant.ErrFailedToRefundPayment
	}

	s.logger.Info(ctx, "success refund QRIS",
		pdkLogger.String("refundId", request.RefundID),
		pdkLogger.String("paymentId", request.PaymentProcessorID),
		pdkLogger.String("refundNo", resp.RefundNo),
	)

	return nil
}
