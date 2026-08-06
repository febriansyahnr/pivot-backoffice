package paymentConsumerController

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/ewallet"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/qr_mpm"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"

	"google.golang.org/protobuf/proto"
)

func (c *paymentConsumer) ProcessPaymentNotification(ctx context.Context, body []byte, channel string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/payment/ProcessPaymentNotification")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from ProcessPaymentNotification", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var (
		now              = time.Now()
		monitoringData   []string
		metricAttributes = map[string]any{}
	)

	defer monitor.WriteAndSend(
		ctx, "process-payment-notification", now, nil, err, func() []string {
			return monitoringData
		},
	)
	defer func() {
		metricData := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentPaymentProcessed,
			Attributes:           metricAttributes,
		}
		if err != nil {
			errType, errDetail := pkgError.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "error when record processed payment custom metric", logger.Error(errMetric))
		}
	}()

	switch channel {
	case constant.ChannelVirtualAccount:
		reqProto := &virtualAccount.PaymentNotificationPayload{}

		if err = proto.Unmarshal(body, reqProto); err != nil {
			return pkgError.New(response.HttpErrRequest, err)
		}

		monitoringData = []string{
			fmt.Sprintf("status:%s", reqProto.Status),
			fmt.Sprintf("paymentMethod:%s", channel),
			fmt.Sprintf("processor:%s", reqProto.Processor),
			fmt.Sprintf("processorID:%s", reqProto.ProcessorReferenceID),
		}
		metricAttributes["status"] = reqProto.Status
		metricAttributes["paymentMethod"] = channel
		metricAttributes["acquirer"] = reqProto.Acquirer

		if reqProto.Number == "" {
			return pkgError.New(response.HttpErrRequest, errors.New("invalid request"))
		}

		bankReferenceId := reqProto.ReferenceNo
		if bankReferenceId == "" {
			var additionalData struct {
				BankReferenceId string `json:"bankReferenceId"`
			}
			if err := json.Unmarshal([]byte(reqProto.AdditionalInfo), &additionalData); err == nil {
				bankReferenceId = additionalData.BankReferenceId
			}
		}

		req := paymentModel.VirtualAccountPaymentNotificationRequest{
			Number:   reqProto.Number,
			Acquirer: reqProto.Acquirer,
			Status:   reqProto.Status,
			PaidAmount: commonModel.Amount{
				Currency: reqProto.PaidAmount.Currency,
				Value:    reqProto.PaidAmount.Value,
			},
			Processor:              reqProto.Processor,
			ProcessorID:            reqProto.ProcessorReferenceID,   // VA ID
			ProcessorTransactionID: reqProto.ProcessorTransactionID, // VA Trx ID
			TrxDatetime:            reqProto.TransactionTime.AsTime(),
			AdditionalData:         reqProto.AdditionalInfo,
			BankReferenceId:        bankReferenceId,
		}

		info := map[string]string{
			"number":          reqProto.Number,
			"acquirer":        reqProto.Acquirer,
			"status":          reqProto.Status,
			"paidAmount":      reqProto.PaidAmount.Value,
			"processorID":     reqProto.ProcessorReferenceID,
			"bankReferenceId": bankReferenceId,
		}
		c.logger.Info(ctx, "ProcessPaymentNotification - VA", pdkLogger.Any("info", info))

		if reqProto.ExpiredAt != nil {
			expiredAt := reqProto.ExpiredAt.AsTime()
			req.ExpiredAt = &expiredAt
		}

		// Check and process unified payment v2
		amountInFloat := 0.0
		amountInFloat, _ = strconv.ParseFloat(req.PaidAmount.Value, 64)
		if processedUnifiedPaymentV2, err := c.checkAndProcessUnifiedPaymentV2(ctx, reqProto.AdditionalInfo, &unifiedPaymentModel.PaymentNotificationRequest{
			PaymentMethodType: constant.UnifiedPaymentMethodVA,
			ChargeStatus:      constant.MapProcessorToChargeStatus(req.Status, ""),
			Amount: unifiedPaymentModel.Amount{
				Currency: req.PaidAmount.Currency,
				Value:    amountInFloat,
			},
			TrxDatetime:              req.TrxDatetime,
			Processor:                req.Processor,
			ProcessorID:              req.ProcessorID,
			ProcessorTransactionID:   req.ProcessorTransactionID,
			ProcessorReferenceNumber: req.Number,
			ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              strings.ToUpper(req.Acquirer),
					VirtualAccountNumber: req.Number,
					VirtualAccountName:   "", // TODO: provide virtual account name on payment notif
					ExpiryAt:             util.ValueOfPtr(req.ExpiredAt),
					BankReferenceNo:      req.BankReferenceId,
				},
			},
		}); err != nil && !errors.Is(err, constant.ErrPaymentNotFound) {
			return err
		} else if processedUnifiedPaymentV2 {
			// Successfully process unified payments v2
			return nil
		}

		// If VA number found in payments then process payments
		err = c.paymentSvc.ProcessVirtualAccountPayment(ctx, &req)
		if err != nil && !errors.Is(err, constant.ErrPaymentNotFound) {
			return err

		} else if err == nil {
			// Successfully process payments
			return nil
		}

		// If VA number found in merchant top up reference then process merchant top up
		err = c.merchantTopUpSvc.ProcessMerchantTopUpWithVirtualAccount(ctx, &req)
		if err != nil && !errors.Is(err, constant.ErrMerchantTopUpReferenceNotFound) {
			return err

		} else if err == nil {
			// Successfully process merchant top up
			return nil
		}

		// There is no succeed process
		return pkgError.New(response.HttpErrRequest, fmt.Errorf("why VA payment happen referenceNumber=%s", req.Number))

	case constant.ChannelQris:
		req := &paymentModel.QrisPaymentNotificationRequest{}

		protoReq := qr_mpm.QrisPaymentNotificationRequest{}
		err := proto.Unmarshal(body, &protoReq)
		if err != nil {
			return pkgError.New(response.HttpErrRequest, err)
		}

		req = &paymentModel.QrisPaymentNotificationRequest{
			Acquirer:    protoReq.Acquirer,
			ReferenceNo: protoReq.ReferenceNo,
			Status:      protoReq.Status,
			PaidAmount: commonModel.Amount{
				Currency: protoReq.PaidAmount.Currency,
				Value:    protoReq.PaidAmount.Value,
			},
			Processor:              protoReq.Processor,
			ProcessorID:            protoReq.ProcessorReferenceID,
			ProcessorTransactionID: protoReq.ProcessorTransactionID,
		}
		info := map[string]string{
			"acquirer":    protoReq.Acquirer,
			"status":      protoReq.Status,
			"referenceNo": protoReq.ReferenceNo,
			"paidAmount":  protoReq.PaidAmount.Value,
			"processorID": protoReq.ProcessorReferenceID,
		}
		c.logger.Info(ctx, "ProcessPaymentNotification - Qris", pdkLogger.Any("info", info))

		if protoReq.ExpiredAt != nil {
			expiredAt := protoReq.ExpiredAt.AsTime()
			req.ExpiredAt = &expiredAt
		}
		if protoReq.TransactionTime != nil {
			req.TransactionTime = protoReq.TransactionTime.AsTime()
		}

		monitoringData = []string{
			fmt.Sprintf("status:%s", req.Status),
			fmt.Sprintf("paymentMethod:%s", channel),
		}
		metricAttributes["status"] = protoReq.Status
		metricAttributes["paymentMethod"] = channel
		metricAttributes["acquirer"] = protoReq.Acquirer

		// Check and process unified payment v2
		amountInFloat := 0.0
		amountInFloat, _ = strconv.ParseFloat(req.PaidAmount.Value, 64)
		if processedUnifiedPaymentV2, err := c.checkAndProcessUnifiedPaymentV2(ctx, protoReq.AdditionalInfo, &unifiedPaymentModel.PaymentNotificationRequest{
			PaymentMethodType: constant.UnifiedPaymentMethodQris,
			ChargeStatus:      constant.MapProcessorToChargeStatus(req.Status, ""),
			Amount: unifiedPaymentModel.Amount{
				Currency: req.PaidAmount.Currency,
				Value:    amountInFloat,
			},
			TrxDatetime:              req.TransactionTime,
			Processor:                req.Processor,
			ProcessorID:              req.ProcessorID,
			ProcessorTransactionID:   req.ProcessorTransactionID,
			ProcessorReferenceNumber: req.ReferenceNo,
		}); err != nil {
			return err
		} else if processedUnifiedPaymentV2 {
			return nil
		}

		return c.paymentSvc.ProcessQrisPayment(ctx, req)

	case constant.ChannelEwallet:

		protoReq := ewallet.WalletPaymentNotificationPayload{}
		err := proto.Unmarshal(body, &protoReq)
		if err != nil {
			return pkgError.New(response.HttpErrRequest, err)
		}

		monitoringData = []string{
			fmt.Sprintf("status:%s", protoReq.Status),
			fmt.Sprintf("paymentMethod:%s", channel),
		}
		metricAttributes["status"] = protoReq.Status
		metricAttributes["paymentMethod"] = channel
		metricAttributes["acquirer"] = protoReq.Acquirer

		var trxTime time.Time
		if protoReq.TransactionTime != nil {
			trxTime = protoReq.TransactionTime.AsTime()
		}

		// Check and process unified payment v2
		amountInFloat := 0.0
		amountInFloat, _ = strconv.ParseFloat(protoReq.Amount.Value, 64)
		if processedUnifiedPaymentV2, err := c.checkAndProcessUnifiedPaymentV2(ctx, protoReq.AdditionalInfo, &unifiedPaymentModel.PaymentNotificationRequest{
			ChargeID:          protoReq.OriginalReferenceId,
			ChargeStatus:      constant.MapProcessorToChargeStatus(protoReq.Status, ""),
			PaymentMethodType: constant.UnifiedPaymentMethodEWallet,
			Amount: unifiedPaymentModel.Amount{
				Currency: protoReq.Amount.Currency,
				Value:    amountInFloat,
			},
			TrxDatetime:            trxTime,
			Processor:              protoReq.Processor,
			ProcessorID:            protoReq.ProcessorReferenceID,
			ProcessorTransactionID: protoReq.ProcessorTransactionID,
		}); err != nil {
			c.logger.Error(ctx, "error when process unified payment v2", logger.Error(err), logger.Any("request", protoReq))
			return err
		} else if processedUnifiedPaymentV2 {
			return nil
		}
		return nil
	default:
		return pkgError.New(response.HttpErrRequest, errors.New("invalid channel"))
	}
}

func (c *paymentConsumer) checkAndProcessUnifiedPaymentV2(ctx context.Context, additionalInfo string, request *unifiedPaymentModel.PaymentNotificationRequest) (isProcessed bool, err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/payment/ProcessPaymentNotification")
	defer segment.End()

	var (
		chargeID         = ""
		paymentSessionID = ""
	)

	// Check if exist ledger, then check the reference id
	json.Unmarshal([]byte(additionalInfo), &struct {
		ProcessorExternalId *string `json:"externalId"`
	}{
		ProcessorExternalId: &chargeID,
	})
	if chargeID == "" && request.ChargeID != "" { // ewallet payment additional data doesn't contained ProcessorReferenceID
		chargeID = request.ChargeID
	}

	// Check the payment is unified payment v2 or not
	isUnifiedPaymentV2 := false
	if chargeID != "" {
		paymentSessionID, err = c.orchestratorSvc.GetReferenceIdByTransactionIdAndType(ctx, chargeID, constant.TypePayment)
		if err != nil {
			return isProcessed, err
		}

		payment, errFind := c.paymentSvc.GetDetailByID(ctx, paymentSessionID)
		if errFind != nil {
			return isProcessed, err
		}

		paymentMetadata := map[string]interface{}{}
		if payment.Metadata != nil {
			paymentMetadata = *payment.Metadata
		}

		if isUnifiedPayment, exists := paymentMetadata["isUnifiedPaymentV2"]; exists {
			isUnifiedPaymentV2 = isUnifiedPayment.(bool)
		}

	} else if request.ProcessorReferenceNumber != "" {
		payment, errFind := c.paymentSvc.GetActivePaymentByProcessorReferenceNumber(ctx, &paymentModel.GetActivePaymentByProcessorReferenceNumberRequest{
			ProcessorReferenceNumber: request.ProcessorReferenceNumber,
			Amount:                   decimal.NewFromFloat(request.Amount.Value),
		})
		if errFind != nil {
			return isProcessed, errFind
		}

		paymentSessionID = payment.UUID
		paymentMetadata := map[string]interface{}{}
		if payment.Metadata != nil {
			paymentMetadata = *payment.Metadata
		}

		if isUnifiedPayment, exists := paymentMetadata["isUnifiedPaymentV2"]; exists {
			isUnifiedPaymentV2 = isUnifiedPayment.(bool)
		}
	}

	if isUnifiedPaymentV2 {
		// Pass paymentSessionId and chargeId
		request.PaymentSessionID = paymentSessionID
		request.ChargeID = chargeID

		// Process unified payment v2
		if err = c.unifiedPaymentSvc.ProcessNotification(ctx, request); err != nil {
			return isProcessed, err
		}

		isProcessed = true
	}

	return
}
