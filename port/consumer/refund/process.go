package refundConsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *refundConsumer) RefundProcess(ctx context.Context, body []byte, _ string) (err error) {
	ctx, span := otelTracer.Start(ctx, "port/consumer/refund/RefundProcess")
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from RefundProcess", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var (
		request          refundModel.RefundProcessRequest
		metricAttributes = map[string]any{}
	)
	if err := json.Unmarshal(body, &request); err != nil {
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, err)
	}

	defer func() {
		metricData := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentRefundProcessed,
			Attributes:           metricAttributes,
		}
		if err != nil {
			errType, errDetail := pkgErr.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "error when record processed payment refund custom metric", logger.Error(errMetric))
		}

	}()

	// Find Refund
	refund, err := c.refundSvc.FindByID(ctx, request.RefundID)
	if err != nil {
		return err
	}

	// Pass refund data to refund process request
	request.Refund = refund
	metricAttributes["merchantId"] = request.Refund.MerchantID

	// Find refund ledger by ID
	refundLedger, err := c.orchestratorSvc.FindByReference(ctx, request.RefundID, constant.TypeRefund)
	if err != nil {
		return err
	} else if refundLedger == nil {
		c.logger.Warn(ctx, "[RefundProcess] refund ledger is not found")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	refundOfPaymentFeeAmount := 0.0
	refundOfPaymentFee, err := c.orchestratorSvc.FindByReference(ctx, request.RefundID, constant.TypeFeeRefund)
	if err != nil {
		return err
	} else if refundOfPaymentFee != nil {
		refundOfPaymentFeeAmount = refundOfPaymentFee.Credit
	}

	refundLedgerAdditionalInfo := orchestratorModel.MetadataRefund{}
	_ = json.Unmarshal(refundLedger.AdditionalInfo.JSONText, &refundLedgerAdditionalInfo)

	// Find payment charge
	paymentCharge, err := c.orchestratorSvc.FindByID(ctx, refundLedgerAdditionalInfo.PaymentChargeID)
	if err != nil {
		return err
	} else if paymentCharge == nil {
		c.logger.Warn(ctx, "[RefundProcess] payment charge is not found")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	paymentFee, err := c.orchestratorSvc.FindByReference(ctx, refundLedgerAdditionalInfo.PaymentSessionID, constant.TypeFee)
	if err != nil {
		return err
	} else if paymentFee == nil {
		c.logger.Warn(ctx, "[RefundProcess] payment fee is not found")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	metricAttributes["paymentMethod"] = paymentCharge.Channel
	request.PaymentMethodType = paymentCharge.Channel
	request.PaymentProcessorID = paymentCharge.ProcessorReferenceId
	request.PaymentClientReferenceID = paymentCharge.ClientReferenceID
	request.RefundLedgerID = refundLedger.UUID.String()
	request.RefundLedgerReferenceID = refundLedger.ReferenceID
	if paymentCharge.SettlementStatus.Valid {
		request.PaymentChargeSettlementStatus = paymentCharge.SettlementStatus.String
	}
	request.PaymentChargeID = paymentCharge.UUID.String()
	request.PaymentChargeAmount = paymentCharge.Credit
	request.PaymentFeeID = paymentFee.UUID.String()
	request.RefundOfPaymentFeeAmount = refundOfPaymentFeeAmount

	if err = c.refundProcessorSvc.Process(ctx, &request); err != nil {
		return err
	}

	return nil
}
