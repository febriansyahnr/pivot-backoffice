package xbPayoutService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *xbPayoutService) ReConfirm(ctx context.Context, request *xbModel.ConfirmPayoutRequest) (*xbModel.ReConfirmEvent, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/ReConfirm")
	defer segment.End()

	var (
		event = &xbModel.ReConfirmEvent{
			PayoutId: request.PayoutId,
		}
	)

	// Find payout by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.PayoutId)
	if err != nil {
		s.logger.Error(ctx, "ReConfirm - Find disbursement by uuid", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if disbursement == nil {
		s.logger.Error(ctx, "ReConfirm - Find disbursement by uuid is not found", logger.Any("uuid", request.PayoutId), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	event.MerchantID = disbursement.MerchantID
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    disbursement.UUID,
		ReferenceId: disbursement.MerchantID,
		From:        serviceName,
	})

	currentReasonType := util.ValueOfPtr(disbursement.ReasonType)
	notAllowToAutoConfirmReasonType := []string{
		constant.XbDisbursementReasonTypeInsufficientBalance,
		constant.XbDisbursementReasonTypeExpired,
	}
	if !util.Contains(notAllowToAutoConfirmReasonType, currentReasonType) {
		event.NeedAutoConfirm = true
	}

	xbAcquirerTransactionId := util.ValueOfPtr(disbursement.ProcessorReferenceID)

	_, err = s.xbCoreProcessorRepo.ReConfirmPayout(ctx, &xbCoreProcessorModel.ConfirmPayoutRequest{
		XbPayoutId:            disbursement.MetadataObj.XbDetail.Uuid,
		MerchantId:            disbursement.MerchantID,
		AcquirerTransactionId: xbAcquirerTransactionId,
	})
	if err != nil {
		s.logger.Error(ctx, "ReConfirm - ReConfirmPayout", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	extendedTime := time.Now().Add(s.config.XbCoreProcessorConfig.GetExtendedExpireAt())
	if err := s.disbursementRepo.ReconfirmXB(ctx, &disbursementModel.ReconfirmXBRequest{
		PayoutId:     request.PayoutId,
		ExtendedTime: extendedTime,
		XBStatus:     constant.XbStatusWaiting,
	}); err != nil {
		return nil, err
	}

	return event, nil
}
