package xbPayoutService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *xbPayoutService) GetRfiDetails(ctx context.Context, request *xbModel.GetRfiDetailsRequest) (*xbModel.GetRfiDetailsResponse, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetRfiDetails")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    request.PayoutId,
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	// Find payout by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.PayoutId)
	if err != nil {
		s.logger.Error(ctx, "GetRfiDetails - Failed to find disbursement by ID", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if disbursement == nil || disbursement.MetadataObj.XbDetail == nil || disbursement.MerchantID != request.MerchantId {
		s.logger.Error(ctx, "GetRfiDetails - Disbursement not found or merchant mismatch", logger.Any("payoutId", request.PayoutId))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	// Get payout from xb-core-processor
	xbResp, err := s.xbCoreProcessorRepo.GetPayoutById(ctx, &xbCoreProcessorModel.GetPayoutRequest{
		Id:         disbursement.MetadataObj.XbDetail.Uuid,
		MerchantId: request.MerchantId,
	})
	if err != nil {
		s.logger.Error(ctx, "GetRfiDetails - Failed to get payout detail from xb-core-processor", logger.Error(err))
		return nil, err
	}
	if xbResp == nil {
		s.logger.Error(ctx, "GetRfiDetails - Payout response from xb-core-processor is nil", logger.Any("payoutId", request.PayoutId))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrPayoutIsNotFound)
	}

	// Validate status
	if xbResp.Status != constant.XbStatusInfoRequested {
		s.logger.Error(ctx, "GetRfiDetails - Payout status is not RFI requested", logger.Any("status", xbResp.Status))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPayoutStatusNotRFI)
	}

	response := &xbModel.GetRfiDetailsResponse{
		Uuid:        disbursement.UUID,
		MerchantId:  disbursement.MerchantID,
		ReferenceId: disbursement.ReferenceID,
	}

	var RfiDetails []*xbModel.RfiDetails
	for _, rfi := range xbResp.RfiDetails {
		RfiDetails = append(RfiDetails, &xbModel.RfiDetails{
			PartnerDocumentID: rfi.UUID.String(),
			Actor:             rfi.Actor,
			Entity:            rfi.Entity,
			DocumentType:      rfi.DocumentType,
			DocumentURL:       rfi.DocumentURL,
			Filename:          rfi.Filename,
			Value:             rfi.Value,
			Comment:           rfi.Comment,
			Status:            rfi.Status,
			RequestedAt:       rfi.RequestedAt,
		})
	}
	response.RfiDetails = RfiDetails

	return response, nil
}
