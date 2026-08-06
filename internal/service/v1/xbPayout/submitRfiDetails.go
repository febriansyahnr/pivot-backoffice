package xbPayoutService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *xbPayoutService) SubmitRfiDetails(ctx context.Context, request *xbModel.SubmitRfiDetailsRequest) (*xbModel.SubmitRfiDetailsResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/SubmitRfiDetails")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    request.PayoutId,
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	// Find payout by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.PayoutId)
	if err != nil {
		s.logger.Error(ctx, "SubmitRfiDetails - Failed to find disbursement by ID", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if disbursement == nil {
		s.logger.Error(ctx, "SubmitRfiDetails - Disbursement not found", logger.Any("payoutId", request.PayoutId))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	// Get current payout status
	xbRespPayout, err := s.xbCoreProcessorRepo.GetPayoutById(ctx, &xbCoreProcessorModel.GetPayoutRequest{
		Id:         disbursement.MetadataObj.XbDetail.Uuid,
		MerchantId: request.MerchantId,
	})
	if err != nil {
		s.logger.Error(ctx, "SubmitRfiDetails - Failed to get payout detail from xb-core-processor", logger.Error(err))
		return nil, err
	}
	if xbRespPayout == nil {
		s.logger.Error(ctx, "SubmitRfiDetails - Payout detail response is nil from xb-core-processor", logger.Any("payoutId", request.PayoutId))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrPayoutIsNotFound)
	}

	// Validate payout status must be INFO_REQUESTED
	if xbRespPayout.Status != constant.XbStatusInfoRequested {
		s.logger.Error(ctx, "SubmitRfiDetails - Payout status is not INFO_REQUESTED", logger.Any("status", xbRespPayout.Status))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPayoutStatusNotRFI)
	}

	// Submit RFI details
	xbResp, err := s.xbCoreProcessorRepo.SubmitRfiDetails(ctx, &xbCoreProcessorModel.SubmitRfiDetailsRequest{
		PayoutId:   disbursement.MetadataObj.XbDetail.Uuid,
		MerchantId: request.MerchantId,
		DocumentId: request.DocumentId,
		Comment:    request.Comment,
		Value:      request.Value,
		Document:   request.Document,
	})
	if err != nil {
		s.logger.Error(ctx, "SubmitRfiDetails - Failed to submit RFI details to xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.SubmitRfiDetailsResponse{
		Uuid:        disbursement.UUID,
		MerchantId:  disbursement.MerchantID,
		ReferenceId: disbursement.ReferenceID,
		DocumentID:  xbResp.UUID.String(),
		Actor:       xbResp.Actor,
		Entity:      xbResp.Entity,
		Type:        xbResp.DocumentType,
		URL:         xbResp.DocumentURL,
		Filename:    xbResp.Filename,
		Value:       xbResp.Value,
		Comment:     xbResp.Comment,
		Status:      xbResp.Status,
		RequestedAt: xbResp.RequestedAt,
	}, nil
}
