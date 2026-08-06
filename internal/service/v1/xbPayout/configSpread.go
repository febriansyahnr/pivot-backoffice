package xbPayoutService

import (
	"context"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *xbPayoutService) GetListConfigSpread(ctx context.Context, request *xbModel.GetListConfigSpreadRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListConfigSpread")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantID.String(),
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListConfigSpread(ctx, &xbCoreProcessorModel.GetListConfigSpreadRequest{
		Page:       request.Page,
		PerPage:    request.PerPage,
		MerchantID: request.MerchantID.String(),
	})
	if err != nil {
		s.logger.Error(ctx, "GetListConfigSpread - Failed to get list config spread from xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetConfigSpreadDetailByID(ctx context.Context, id string) (*xbModel.GetConfigSpreadDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetConfigSpreadDetailByID")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: "",
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetConfigSpreadDetailByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "GetConfigSpreadDetailByID - Failed to get config spread detail from xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.GetConfigSpreadDetailResponse{
		UUID:                xbCoreResp.UUID,
		MerchantID:          xbCoreResp.MerchantID,
		SourceCurrency:      xbCoreResp.SourceCurrency,
		DestinationCurrency: xbCoreResp.DestinationCurrency,
		SpreadType:          xbCoreResp.SpreadType,
		SpreadValue:         xbCoreResp.SpreadValue,
		CreatedAt:           xbCoreResp.CreatedAt,
		UpdatedAt:           xbCoreResp.UpdatedAt,
	}, nil
}
func (s *xbPayoutService) CreateConfigSpread(ctx context.Context, request *xbModel.CreateConfigSpreadRequest) (*xbModel.CreateConfigSpreadResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/CreateConfigSpread")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantID.String(),
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.CreateConfigSpread(ctx, &xbCoreProcessorModel.CreateConfigSpreadRequest{
		MerchantID:          request.MerchantID,
		SourceCurrency:      request.SourceCurrency,
		DestinationCurrency: request.DestinationCurrency,
		SpreadType:          request.SpreadType,
		SpreadValue:         request.SpreadValue,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateConfigSpread - Failed to create config spread in xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateConfigSpreadResponse{
		UUID:    xbCoreResp.UUID,
		Created: xbCoreResp.Created,
	}, nil
}

func (s *xbPayoutService) UpdateConfigSpread(ctx context.Context, request *xbModel.UpdateConfigSpreadRequest) (*xbModel.UpdateConfigSpreadResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/UpdateConfigSpread")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: "",
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.UpdateConfigSpread(ctx, &xbCoreProcessorModel.UpdateConfigSpreadRequest{
		UUID:                request.UUID,
		SourceCurrency:      request.SourceCurrency,
		DestinationCurrency: request.DestinationCurrency,
		SpreadType:          request.SpreadType,
		SpreadValue:         request.SpreadValue,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateConfigSpread - Failed to update config spread in xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.UpdateConfigSpreadResponse{
		UUID:    xbCoreResp.UUID,
		Updated: xbCoreResp.Updated,
	}, nil
}
