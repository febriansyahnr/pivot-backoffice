package xbPayoutService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *xbPayoutService) UploadUnderlyingDocument(ctx context.Context, request *xbModel.UploadUnderlyingDocumentRequest) (*xbModel.UploadUnderlyingDocumentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/UploadUnderlyingDocument")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    request.PayoutId,
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	// Find payout by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.PayoutId)
	if err != nil {
		s.logger.Error(ctx, "UploadUnderlyingDocument - Failed to find disbursement by ID", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)

	} else if disbursement == nil || disbursement.MerchantID != request.MerchantId || disbursement.MetadataObj.XbDetail == nil {
		s.logger.Info(ctx, "UploadUnderlyingDocument - Payout not found or merchant ID does not match or missing XB metadata", logger.Any("request", map[string]string{
			"merchantId": request.MerchantId,
			"payoutId":   request.PayoutId,
		}))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	// Return if expired
	if time.Now().UTC().After(disbursement.MetadataObj.XbDetail.ExpiredAt) {
		s.logger.Error(ctx, "UploadUnderlyingDocument - Payout is already expired", logger.Any("payoutId", request.PayoutId))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPayoutAlreadyExpired)
	}

	// Upload document to xb-core-processor
	xbResp, err := s.xbCoreProcessorRepo.UploadUnderlyingDocument(ctx, &xbCoreProcessorModel.UploadUnderlyingDocumentRequest{
		XbPayoutId: disbursement.MetadataObj.XbDetail.Uuid,
		MerchantId: request.MerchantId,
		Document:   request.Document,
	})
	if err != nil {
		s.logger.Error(ctx, "UploadUnderlyingDocument - Failed to upload underlying document to xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.UploadUnderlyingDocumentResponse{
		DocumentReference: xbResp.DocumentReference,
	}, nil
}
