package unifiedPaymentService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/fds"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) UploadProofOfPayment(
	ctx context.Context,
	request *unifiedPaymentModel.UploadProofOfPaymentRequest,
) (_ *unifiedPaymentModel.UploadProofOfPaymentResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/UploadProofOfPayment")
	defer segment.End()

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentID)
	if err != nil {
		s.logger.Error(ctx, "Failed to find payment", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}
	if payment == nil {
		return nil, pkgErr.New(response.HttpErrNotFound, constant.ErrPaymentNotFound)
	}

	if payment.MerchantID != request.MerchantID {
		return nil, pkgErr.New(response.HttpErrNotFound, constant.ErrPaymentNotFound)
	}

	if !s.isValidStatusForInvestigation(payment.Status) {
		return nil, pkgErr.New(response.HttpErrRequest, constant.ErrPaymentAlreadyInFinalStatus)
	}

	if !s.isValidPaymentMethodForInvestigation(payment.PaymentMethod.Type) {
		return nil, pkgErr.New(response.HttpErrRequest, constant.ErrPaymentMethodNotAllowed)
	}

	if enabled, err := s.merchantRepo.IsInvestigationFlowEnabled(ctx, payment.MerchantID); err != nil {
		s.logger.Error(ctx, "Failed to check enabled payment investigation", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if !enabled {
		return nil, pkgErr.New(response.HttpErrRequest, constant.ErrInvestigationNotEnabled)
	}

	merchantConfig, err := s.merchantSvc.GetFDSConfig(ctx, payment.MerchantID)
	if err != nil {
		return nil, err // Error has been wrapped
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceID,
		OriginId:    payment.UUID,
		ReferenceId: payment.MerchantID,
	})

	inquiryResult, inquiryErr := s.performBankInquiryForInvestigation(ctx, payment)
	if inquiryErr != nil {
		if isClientError(inquiryErr) {
			return nil, pkgErr.New(response.HttpErrRequest, constant.ErrBankInquiryFailed)
		}
	}

	charge, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.TypePayment)
	chargeID := ""
	if err != nil {
		s.logger.Warn(ctx, "Failed to find existing charge for payment, will create new",
			logger.String("paymentUUID", payment.UUID),
			logger.Error(err),
		)
	}
	if charge != nil {
		chargeID = charge.UUID.String()
	} else {
		chargeID = uuid.NewString()
	}

	if inquiryResult != nil {
		switch inquiryResult.Status {
		case constant.InquiryStatusSuccess:
			_ = s.processNotificationFromInquiry(ctx, payment, inquiryResult, chargeID)
			return nil, pkgErr.New(response.HttpErrRequest, constant.ErrBankConfirmedSuccess)
		case constant.InquiryStatusFailed:
			_ = s.processNotificationFromInquiry(ctx, payment, inquiryResult, chargeID)
			return nil, pkgErr.New(response.HttpErrRequest, constant.ErrBankConfirmedFailed)
		}
	}

	// FDS (Fraud Detection System) Checker
	velocityRule := fds.VelocityRule{
		Member: payment.UUID,
		Period: merchantConfig.FDSConfig.ProofOfPayment.Velocity.Window.Period(),
		Rate:   merchantConfig.FDSConfig.ProofOfPayment.Velocity.Threshold.Count,
	}
	velocityKey := fmt.Sprintf(constant.FDSVelocityMerchantUploadPoPKeyFmt, payment.MerchantID)

	if result, err := s.fdsVelocityCheck.Allow(ctx, velocityKey, velocityRule); err != nil {
		s.logger.Error(ctx, "FDS velocity rule check failed", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if !result.Allowed {
		return nil, pkgErr.New(response.HttpErrTooManyRequest, constant.ErrProofOfPaymentRateLimitExceeded)
	}
	defer func() {
		if err == nil {
			return
		}
		if rollbackErr := s.fdsVelocityCheck.Rollback(ctx, velocityKey, payment.UUID); rollbackErr != nil {
			s.logger.Error(ctx, "Velocity check rollback failed; unable to restore total remaining", logger.Error(rollbackErr))
		}
	}()

	objectName := fmt.Sprintf("%s/%s/%s.%s",
		s.config.GCSConfig.InvestigationPoPFolderName,
		payment.MerchantID,
		payment.UUID,
		request.FileExtension,
	)

	uploadResult, err := s.storage.UploadFileFromMultipartToBucket(
		ctx,
		s.config.GCSConfig.ServiceBucketName,
		objectName,
		request.ProofOfPayment,
		true,
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to upload proof of payment to GCS", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	err = s.paymentRepo.UpdatePaymentForInvestigation(ctx, paymentModel.UpdatePaymentForInvestigationRequest{
		PaymentID:  payment.UUID,
		MerchantID: payment.MerchantID,
		ReasonType: paymentConst.InvestigationStatusInProcess,
		InvestigationMetadata: paymentModel.InvestigationPoPMetadata{
			Bucket:        uploadResult.Bucket,
			Path:          uploadResult.ObjectName,
			MerchantNotes: request.Reason,
		},
		StartedAt: time.Now().UTC(),
	})
	if err != nil {
		s.logger.Error(ctx, "Failed to update payment investigation metadata", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	s.paymentSvc.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryInvestigationInProcess)

	notificationRequest := &unifiedPaymentModel.PaymentNotificationRequest{
		PaymentSessionID:  payment.UUID,
		ChargeID:          chargeID,
		ChargeStatus:      constant.ChargeStatusSuccess,
		PaymentMethodType: constant.MapToUnifiedPaymentMethod(payment.PaymentMethod.Type),
		Amount: unifiedPaymentModel.Amount{
			Value:    payment.Amount.InexactFloat64(),
			Currency: payment.Currency,
		},
		Processor:   constant.SnapCoreProcessor, // currently only for QR and VA
		TrxDatetime: time.Now().UTC(),
	}

	if err := s.ProcessNotification(ctx, notificationRequest); err != nil {
		s.logger.Error(ctx, "Failed to process notification for investigation",
			logger.String("paymentId", payment.UUID),
			logger.Error(err),
		)
		return nil, err
	}

	return &unifiedPaymentModel.UploadProofOfPaymentResponse{
		PaymentID:           payment.UUID,
		Status:              constant.ChargeStatusSuccess,
		InvestigationStatus: paymentConst.InvestigationStatusInProcess,
		CreatedAt:           payment.CreatedAt,
		UpdatedAt:           time.Now().UTC(),
	}, nil
}

func (s *UnifiedPaymentService) isValidStatusForInvestigation(status string) bool {
	switch status {
	case constant.ChargeStatusProcessing,
		constant.ChargeStatusWaitingForUserAction,
		constant.ChargeStatusExpired,
		constant.UnifiedPaymentSessionStatusRequireAction:
		return true
	default:
		return false
	}
}

func (s *UnifiedPaymentService) isValidPaymentMethodForInvestigation(paymentMethodType string) bool {
	return paymentMethodType == paymentConst.PAYMENT_METHOD_QRIS ||
		paymentMethodType == paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT
}

func (s *UnifiedPaymentService) performBankInquiryForInvestigation(ctx context.Context, payment *paymentModel.Payment) (*unifiedPaymentModel.InquiryResult, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/performBankInquiryForInvestigation")
	defer segment.End()

	switch payment.PaymentMethod.Type {
	case paymentConst.PAYMENT_METHOD_QRIS:
		return s.performQrisInquiry(ctx, payment)
	case paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		return s.performVAInquiry(ctx, payment, unifiedPaymentModel.PerformInquiryRequest{})
	default:
		return nil, nil
	}
}

func isClientError(err error) bool {
	if err == nil {
		return false
	}

	errType, _ := pkgErr.ExtractError(err)
	return errType == response.HttpErrRequest
}
