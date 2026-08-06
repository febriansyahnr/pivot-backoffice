package xbPayoutService

import (
	"context"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func (s *xbPayoutService) GetListMasterCountry(ctx context.Context, request *xbModel.GetListMasterCountryRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterCountry")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterCountry(ctx, &xbCoreProcessorModel.GetListMasterCountryRequest{
		ActiveOnly:   request.ActiveOnly,
		CountryCode:  request.CountryCode,
		CurrencyCode: request.CurrencyCode,
		FetchAll:     request.FetchAll,
		Page:         request.Page,
		PerPage:      request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterState(ctx context.Context, request *xbModel.GetListMasterStateRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterState")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterState(ctx, &xbCoreProcessorModel.GetListMasterStateRequest{
		CountryCode: request.CountryCode,
		Name:        request.Name,
		FetchAll:    request.FetchAll,
		Page:        request.Page,
		PerPage:     request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterCity(ctx context.Context, request *xbModel.GetListMasterCityRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterCity")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterCity(ctx, &xbCoreProcessorModel.GetListMasterCityRequest{
		StateUUID: request.StateUUID,
		Name:      request.Name,
		FetchAll:  request.FetchAll,
		Page:      request.Page,
		PerPage:   request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterCurrency(ctx context.Context, request *xbModel.GetListMasterCurrencyRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterCurrency")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterCurrency(ctx, &xbCoreProcessorModel.GetListMasterCurrencyRequest{
		Code:     request.Code,
		FetchAll: request.FetchAll,
		Page:     request.Page,
		PerPage:  request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterCurrencyMapping(ctx context.Context, request *xbModel.GetListMasterCurrencyMappingRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterCurrencyMapping")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterCurrencyMapping(ctx, &xbCoreProcessorModel.GetListMasterCurrencyMappingRequest{
		CountryCode:    request.CountryCode,
		TransferMethod: request.TransferMethod,
		FetchAll:       request.FetchAll,
		Page:           request.Page,
		PerPage:        request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterIdentificationType(ctx context.Context, request *xbModel.GetListMasterIdentificationTypeRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterIdentificationType")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterIdentificationType(ctx, &xbCoreProcessorModel.GetListMasterIdentificationTypeRequest{
		AccountType: request.AccountType,
		FetchAll:    request.FetchAll,
		Page:        request.Page,
		PerPage:     request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterAccountType(ctx context.Context, request *xbModel.GetListMasterAccountTypeRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterAccountType")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterAccountType(ctx, &xbCoreProcessorModel.GetListMasterAccountTypeRequest{
		Code:     request.Code,
		FetchAll: request.FetchAll,
		Page:     request.Page,
		PerPage:  request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterPurpose(ctx context.Context, request *xbModel.GetListMasterPurposeRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterPurpose")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterPurpose(ctx, &xbCoreProcessorModel.GetListMasterPurposeRequest{
		Code:        request.Code,
		FetchAll:    request.FetchAll,
		Page:        request.Page,
		PerPage:     request.PerPage,
		CountryCode: request.CountryCode,
		RoutingCode: request.RoutingCode,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterTransferMethod(ctx context.Context, request *xbModel.GetListMasterTransferMethodRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterTransferMethod")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterTransferMethod(ctx, &xbCoreProcessorModel.GetListMasterTransferMethodRequest{
		Code:     request.Code,
		FetchAll: request.FetchAll,
		Page:     request.Page,
		PerPage:  request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetListMasterSourceOfIncome(ctx context.Context, request *xbModel.GetListMasterSourceOfIncomeRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListMasterSourceOfIncome")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListMasterSourceOfIncome(ctx, &xbCoreProcessorModel.GetListMasterSourceOfIncomeRequest{
		Name:     request.Name,
		FetchAll: request.FetchAll,
		Page:     request.Page,
		PerPage:  request.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}
