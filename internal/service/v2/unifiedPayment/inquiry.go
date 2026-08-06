package unifiedPaymentService

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreQRModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) performPaymentInquiry(ctx context.Context, payment *paymentModel.Payment, inquiryReq unifiedPaymentModel.PerformInquiryRequest) (*unifiedPaymentModel.InquiryResult, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/performPaymentInquiry")
	defer segment.End()

	if payment == nil {
		return nil, nil
	}

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceID,
		OriginId:    payment.UUID,
		ReferenceId: payment.MerchantID,
	})

	if !s.isInquiryEligible(ctx, payment) {
		return nil, nil
	}

	if s.isWithinCooldown(ctx, payment.UUID) {
		s.logger.Info(ctx, "Payment inquiry within cooldown, skipping",
			logger.String("paymentUUID", payment.UUID),
		)
		return nil, nil
	}

	config := s.getPaymentAutoInquiryConfig(ctx)
	inquiryStartTime := time.Now()

	s.logger.Info(ctx, "Starting payment inquiry",
		logger.String("paymentUUID", payment.UUID),
		logger.String("paymentMethod", payment.PaymentMethod.Type),
		logger.String("currentStatus", payment.Status),
		logger.Time("inquiryAttemptTime", inquiryStartTime),
	)

	var result *unifiedPaymentModel.InquiryResult
	var err error

	switch payment.PaymentMethod.Type {
	case paymentConst.PAYMENT_METHOD_QRIS:
		result, err = s.performQrisInquiry(ctx, payment)
	case paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		result, err = s.performVAInquiry(ctx, payment, inquiryReq)
	default:
		return nil, nil
	}

	if err != nil {
		s.logger.Error(ctx, "Payment inquiry failed, returning stored status",
			logger.String("paymentUUID", payment.UUID),
			logger.String("paymentMethod", payment.PaymentMethod.Type),
			logger.String("currentStatus", payment.Status),
			logger.Duration("inquiryDuration", time.Since(inquiryStartTime)),
			logger.Error(err),
		)
		s.setCooldown(ctx, payment.UUID, config.CooldownSeconds)
		return nil, nil
	}

	now := time.Now()
	s.setCooldown(ctx, payment.UUID, config.CooldownSeconds)

	if result != nil {
		result.LastInquiryAt = &now
		s.logger.Info(ctx, "Payment inquiry completed",
			logger.String("paymentUUID", payment.UUID),
			logger.String("paymentMethod", payment.PaymentMethod.Type),
			logger.String("inquiryStatus", result.Status),
			logger.String("responseCode", result.ResponseCode),
			logger.String("responseMessage", result.ResponseMessage),
			logger.Bool("statusUpdated", result.UpdatedStatus),
			logger.Duration("inquiryDuration", time.Since(inquiryStartTime)),
		)
	}

	return result, nil
}

func (s *UnifiedPaymentService) performQrisInquiry(ctx context.Context, payment *paymentModel.Payment) (*unifiedPaymentModel.InquiryResult, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/performQrisInquiry")
	defer segment.End()

	snapCoreId := s.getSnapCoreIdForQris(ctx, payment)
	if snapCoreId == "" {
		s.logger.Warn(ctx, "Payment has no snap core ID for QRIS inquiry",
			logger.String("paymentUUID", payment.UUID),
		)
		return nil, nil
	}

	s.logger.Info(ctx, "Calling QRIS inquiry partner API",
		logger.String("paymentUUID", payment.UUID),
		logger.String("snapCoreId", snapCoreId),
	)

	resp, err := s.snapCoreRepo.InquiryStatusQris(ctx, &snapCoreQRModel.InquiryStatusQrMpmRequest{
		QrisUUID:    snapCoreId,
		SkipPublish: true,
	})
	if err != nil {
		s.logger.Error(ctx, "QRIS inquiry partner API error",
			logger.String("paymentUUID", payment.UUID),
			logger.String("snapCoreId", snapCoreId),
			logger.Error(err),
		)
		return nil, err
	}

	if resp == nil || resp.Data == nil {
		s.logger.Warn(ctx, "QRIS inquiry returned empty response",
			logger.String("paymentUUID", payment.UUID),
			logger.String("snapCoreId", snapCoreId),
		)
		return nil, nil
	}

	s.logger.Info(ctx, "QRIS inquiry partner response received",
		logger.String("paymentUUID", payment.UUID),
		logger.String("snapCoreId", snapCoreId),
		logger.String("responseCode", resp.Data.ResponseCode),
		logger.String("status", resp.Data.Status),
	)

	return s.mapQrisInquiryResponse(resp), nil
}

func (s *UnifiedPaymentService) performVAInquiry(ctx context.Context, payment *paymentModel.Payment, inquiryReq unifiedPaymentModel.PerformInquiryRequest) (*unifiedPaymentModel.InquiryResult, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/performVAInquiry")
	defer segment.End()

	vaNumber := s.getVANumberFromPayment(payment)
	if vaNumber == "" {
		s.logger.Warn(ctx, "Payment has no VA number for VA inquiry",
			logger.String("paymentUUID", payment.UUID),
		)
		return nil, nil
	}

	s.logger.Info(ctx, "Calling VA inquiry partner API",
		logger.String("paymentUUID", payment.UUID),
		logger.String("vaNumber", vaNumber),
	)

	req := &snapCoreVAModel.InquiryStatusVARequest{
		VirtualAccount: vaNumber,
		SkipPublish:    true,
		ExternalID:     inquiryReq.LedgerID,
	}

	resp, err := s.snapCoreRepo.InquiryStatusVirtualAccount(ctx, req)
	if err != nil {
		s.logger.Error(ctx, "VA inquiry partner API error",
			logger.String("paymentUUID", payment.UUID),
			logger.String("vaNumber", vaNumber),
			logger.Error(err),
		)
		return nil, err
	}

	if resp == nil {
		s.logger.Warn(ctx, "VA inquiry returned empty response",
			logger.String("paymentUUID", payment.UUID),
			logger.String("vaNumber", vaNumber),
		)
		return nil, nil
	}

	s.logger.Info(ctx, "VA inquiry partner response received",
		logger.String("paymentUUID", payment.UUID),
		logger.String("vaNumber", vaNumber),
		logger.String("responseCode", resp.Data.ResponseCode),
	)

	return s.mapVAInquiryResponse(resp), nil
}

func (s *UnifiedPaymentService) getVANumberFromPayment(payment *paymentModel.Payment) string {
	if payment.ProcessorReferenceNumber != nil && *payment.ProcessorReferenceNumber != "" {
		return *payment.ProcessorReferenceNumber
	}

	if payment.Metadata != nil {
		if snapCore, ok := (*payment.Metadata)["snapCore"].(map[string]any); ok {
			if number, ok := snapCore["number"].(string); ok {
				return number
			}
		}
	}
	return ""
}

func (s *UnifiedPaymentService) isInquiryEligible(ctx context.Context, payment *paymentModel.Payment) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/isInquiryEligible")
	defer segment.End()

	pendingStatuses := []string{
		constant.UnifiedPaymentSessionStatusProcessing,
		constant.UnifiedPaymentSessionStatusRequireAction,
		constant.ChargeStatusProcessing,
		constant.ChargeStatusWaitingForUserAction,
	}
	if !slices.Contains(pendingStatuses, payment.Status) {
		return false
	}

	config := s.getPaymentAutoInquiryConfig(ctx)
	return slices.Contains(config.EnabledMethods, payment.PaymentMethod.Type)
}

func (s *UnifiedPaymentService) isWithinCooldown(ctx context.Context, paymentUUID string) bool {
	if s.redis == nil {
		return false
	}

	key := fmt.Sprintf(constant.RedisKeyInquiryCooldownFmt, paymentUUID)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		s.logger.Warn(ctx, "Failed to check cooldown",
			logger.String("paymentUUID", paymentUUID),
			logger.Error(err),
		)
		return false
	}

	return exists > 0
}

func (s *UnifiedPaymentService) setCooldown(ctx context.Context, paymentUUID string, cooldownSeconds int) {
	if s.redis == nil {
		return
	}

	key := fmt.Sprintf(constant.RedisKeyInquiryCooldownFmt, paymentUUID)
	ttl := time.Duration(cooldownSeconds) * time.Second

	if err := s.redis.Set(ctx, key, time.Now().Format(time.RFC3339), ttl).Err(); err != nil {
		s.logger.Warn(ctx, "Failed to set cooldown",
			logger.String("paymentUUID", paymentUUID),
			logger.Error(err),
		)
	}
}

func (s *UnifiedPaymentService) mapQrisInquiryResponse(resp *snapCoreQRModel.QrisInquiryStatusResponse) *unifiedPaymentModel.InquiryResult {
	if resp == nil || resp.Data == nil {
		return nil
	}

	result := &unifiedPaymentModel.InquiryResult{
		Status:                 util.MapQRLatestStatusToPaymentStatus(resp.Data.Status),
		ResponseCode:           resp.Data.ResponseCode,
		ResponseMessage:        resp.Data.ResponseMessage,
		ProcessorID:            resp.Data.UUID,
		ProcessorTransactionID: resp.Data.TransactionID,
		ProcessorReferenceNo:   resp.Data.AcquirerReferenceNo,
	}

	if resp.Data.Amount != nil {
		if amountVal, err := strconv.ParseFloat(resp.Data.Amount.Value, 64); err == nil {
			result.Amount = &unifiedPaymentModel.Amount{
				Value:    amountVal,
				Currency: resp.Data.Amount.Currency,
			}
		}
	}

	result.UpdatedStatus = s.isFinalStatus(result.Status)

	return result
}

func (s *UnifiedPaymentService) mapVAInquiryResponse(resp *snapCoreVAModel.InquiryStatusVAResponse) *unifiedPaymentModel.InquiryResult {
	if resp == nil {
		return nil
	}

	// Notes:
	// - Failed/Not found still considered as pending, we need to double check with snapcore & bank partner first before decide final status
	status := constant.InquiryStatusPending
	switch {
	case resp.IsPaid():
		status = constant.InquiryStatusSuccess
	case resp.IsConflict():
		status = constant.InquiryStatusSuccess
	}

	result := &unifiedPaymentModel.InquiryResult{
		Status:          status,
		ResponseCode:    resp.Data.ResponseCode,
		ResponseMessage: resp.Data.ResponseMessage,
	}

	if resp.Data.VirtualAccountData != nil {
		vaData := resp.Data.VirtualAccountData
		if vaData.PaidAmount != nil {
			if amountVal, err := strconv.ParseFloat(vaData.PaidAmount.Value, 64); err == nil {
				result.Amount = &unifiedPaymentModel.Amount{
					Value:    amountVal,
					Currency: vaData.PaidAmount.Currency,
				}
			}
		}
		result.ProcessorTransactionID = vaData.PaymentRequestId
		result.ProcessorReferenceNo = vaData.ReferenceNo
		if vaData.TrxDateTime != "" {
			if t, err := time.Parse(time.RFC3339, vaData.TrxDateTime); err == nil {
				result.TrxDatetime = &t
			}
		}
	}

	result.UpdatedStatus = s.isFinalStatus(result.Status)

	return result
}

func (s *UnifiedPaymentService) isFinalStatus(status string) bool {
	finalStatuses := []string{
		constant.InquiryStatusSuccess,
		constant.InquiryStatusFailed,
		constant.InquiryStatusExpired,
		constant.ChargeStatusSuccess,
		constant.ChargeStatusFailed,
		constant.ChargeStatusExpired,
	}
	return slices.Contains(finalStatuses, status)
}

func (s *UnifiedPaymentService) mapInquiryStatusToChargeStatus(inquiryStatus string) string {
	switch inquiryStatus {
	case constant.InquiryStatusSuccess:
		return constant.ChargeStatusSuccess
	case constant.InquiryStatusFailed:
		return constant.ChargeStatusFailed
	case constant.InquiryStatusExpired:
		return constant.ChargeStatusExpired
	case constant.InquiryStatusPending:
		return constant.ChargeStatusProcessing
	default:
		return inquiryStatus
	}
}

func (s *UnifiedPaymentService) getSnapCoreIdForQris(ctx context.Context, payment *paymentModel.Payment) string {
	if payment.SnapCoreId != nil && *payment.SnapCoreId != "" {
		return *payment.SnapCoreId
	}

	accountTrx, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		s.logger.Error(ctx, "Failed to get account transaction for QRIS inquiry fallback",
			logger.String("paymentUUID", payment.UUID),
			logger.Error(err),
		)
		return ""
	}

	if accountTrx != nil && accountTrx.ProcessorReferenceId != "" && accountTrx.ProcessorReferenceId != "-" {
		return accountTrx.ProcessorReferenceId
	}

	return ""
}

func (s *UnifiedPaymentService) buildNotificationRequestFromInquiry(
	payment *paymentModel.Payment,
	inquiryResult *unifiedPaymentModel.InquiryResult,
	chargeID string,
) *unifiedPaymentModel.PaymentNotificationRequest {
	chargeStatus := s.mapInquiryStatusToChargeStatus(inquiryResult.Status)

	req := &unifiedPaymentModel.PaymentNotificationRequest{
		PaymentSessionID:         payment.UUID,
		PaymentMethodType:        payment.PaymentMethod.Type,
		ChargeID:                 chargeID,
		ChargeStatus:             chargeStatus,
		Processor:                constant.SnapCoreProcessor,
		ProcessorID:              inquiryResult.ProcessorID,
		ProcessorTransactionID:   inquiryResult.ProcessorTransactionID,
		ProcessorReferenceNumber: inquiryResult.ProcessorReferenceNo,
	}

	if inquiryResult.Amount != nil {
		req.Amount = unifiedPaymentModel.Amount{
			Value:    inquiryResult.Amount.Value,
			Currency: inquiryResult.Amount.Currency,
		}
	} else {
		req.Amount = unifiedPaymentModel.Amount{
			Value:    payment.Amount.InexactFloat64(),
			Currency: constant.CurrencyIDR,
		}
	}

	if inquiryResult.TrxDatetime != nil {
		req.TrxDatetime = *inquiryResult.TrxDatetime
	} else {
		req.TrxDatetime = time.Now().UTC()
	}

	return req
}

func (s *UnifiedPaymentService) processNotificationFromInquiry(
	ctx context.Context,
	payment *paymentModel.Payment,
	inquiryResult *unifiedPaymentModel.InquiryResult,
	chargeID string,
) error {
	if inquiryResult == nil || payment == nil {
		return nil
	}

	if s.mapInquiryStatusToChargeStatus(inquiryResult.Status) == constant.ChargeStatusExpired {
		return nil
	}

	notificationReq := s.buildNotificationRequestFromInquiry(payment, inquiryResult, chargeID)

	s.logger.Info(ctx, "Processing inquiry result via notification flow",
		logger.String("paymentUUID", payment.UUID),
		logger.String("chargeStatus", notificationReq.ChargeStatus),
	)

	if err := s.ProcessNotification(ctx, notificationReq); err != nil {
		s.logger.Error(ctx, "Failed to process inquiry notification",
			logger.String("paymentUUID", payment.UUID),
			logger.Error(err),
		)
		return err
	}

	return nil
}
