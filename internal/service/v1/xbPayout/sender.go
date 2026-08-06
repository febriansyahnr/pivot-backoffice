package xbPayoutService

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	"github.com/paper-indonesia/pdk/v2/logger"

	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func (s *xbPayoutService) CreateSender(ctx context.Context, request *xbModel.CreateSenderRequest) (*xbModel.CreateSenderResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/CreateSender")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbSender, err := s.xbCoreProcessorRepo.CreateSender(ctx, &xbCoreProcessorModel.CreateSenderRequest{
		MerchantId:           request.MerchantId,
		Name:                 request.Name,
		AccountType:          request.AccountType,
		Address:              request.Address,
		City:                 request.City,
		Postcode:             request.Postcode,
		State:                request.State,
		CountryCode:          request.CountryCode,
		IdentificationType:   request.IdentificationType,
		IdentificationNumber: request.IdentificationNumber,
		BankAccountNumber:    request.BankAccountNumber,
		ContactCountryCode:   request.ContactCountryCode,
		ContactNumber:        request.ContactNumber,
		Dob:                  request.Dob,
		SourceOfIncome:       request.SourceOfIncome,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateSender - Failed to create sender in xb-core-processor", logger.Error(err))
		return nil, err
	}

	var created bool
	if xbSender.UUID != uuid.Nil {
		created = true
	}

	return &xbModel.CreateSenderResponse{
		UUID:                 xbSender.UUID,
		Name:                 xbSender.Name,
		CountryCode:          xbSender.CountryCode,
		State:                xbSender.State,
		City:                 xbSender.City,
		Address:              xbSender.Address,
		Postcode:             xbSender.Postcode,
		AccountType:          xbSender.AccountType,
		IdentificationType:   xbSender.IdentificationType,
		IdentificationNumber: xbSender.IdentificationNumber,
		BankAccountNumber:    xbSender.BankAccountNumber,
		Dob:                  xbSender.Dob,
		ContactCountryCode:   xbSender.ContactCountryCode,
		ContactNumber:        xbSender.ContactNumber,
		SourceOfIncome:       xbSender.SourceOfIncome,
		CreatedAt:            xbSender.CreatedAt,
		Created:              created,
	}, nil
}

func (s *xbPayoutService) GetListSender(ctx context.Context, request *xbModel.GetListSenderRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListSender")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListSender(ctx, &xbCoreProcessorModel.GetListSenderRequest{
		MerchantId:      request.MerchantId,
		Page:            request.Page,
		PerPage:         request.PerPage,
		FetchAll:        request.FetchAll,
		ShowDeactivated: request.ShowDeactivated,
		Name:            request.Name,
		AccountType:     request.AccountType,
	})
	if err != nil {
		s.logger.Error(ctx, "GetListSender - Failed to get list sender from xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}

func (s *xbPayoutService) GetSenderById(ctx context.Context, request *xbModel.GetSenderByIdRequest) (*xbModel.CreateSenderResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetSenderById")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetSenderById(ctx, &xbCoreProcessorModel.GetSenderByIdRequest{
		MerchantId: request.MerchantId,
		SenderId:   request.SenderId,
	})
	if err != nil {
		s.logger.Error(ctx, "GetSenderById - Failed to get sender detail from xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateSenderResponse{
		UUID:                 xbCoreResp.UUID,
		Name:                 xbCoreResp.Name,
		CountryCode:          xbCoreResp.CountryCode,
		State:                xbCoreResp.State,
		City:                 xbCoreResp.City,
		Address:              xbCoreResp.Address,
		Postcode:             xbCoreResp.Postcode,
		AccountType:          xbCoreResp.AccountType,
		IdentificationType:   xbCoreResp.IdentificationType,
		IdentificationNumber: xbCoreResp.IdentificationNumber,
		BankAccountNumber:    xbCoreResp.BankAccountNumber,
		Dob:                  xbCoreResp.Dob,
		ContactCountryCode:   xbCoreResp.ContactCountryCode,
		ContactNumber:        xbCoreResp.ContactNumber,
		SourceOfIncome:       xbCoreResp.SourceOfIncome,
		CreatedAt:            xbCoreResp.CreatedAt,
		DeactivatedAt:        xbCoreResp.DeactivatedAt,
		UpdatedAt:            xbCoreResp.UpdatedAt,
	}, nil
}

func (s *xbPayoutService) UpdateSender(ctx context.Context, request *xbModel.UpdateSenderRequest) (*xbModel.CreateSenderResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/UpdateSender")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.UpdateSender(ctx, &xbCoreProcessorModel.UpdateSenderRequest{
		SenderId:             request.SenderId,
		MerchantId:           request.MerchantId,
		Name:                 request.Name,
		CountryCode:          request.CountryCode,
		State:                request.State,
		City:                 request.City,
		Address:              request.Address,
		Postcode:             request.Postcode,
		AccountType:          request.AccountType,
		IdentificationType:   request.IdentificationType,
		IdentificationNumber: request.IdentificationNumber,
		BankAccountNumber:    request.BankAccountNumber,
		Dob:                  request.Dob,
		ContactCountryCode:   request.ContactCountryCode,
		ContactNumber:        request.ContactNumber,
		SourceOfIncome:       request.SourceOfIncome,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateSender - Failed to update sender in xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateSenderResponse{
		UUID:                 xbCoreResp.UUID,
		Name:                 xbCoreResp.Name,
		CountryCode:          xbCoreResp.CountryCode,
		State:                xbCoreResp.State,
		City:                 xbCoreResp.City,
		Address:              xbCoreResp.Address,
		Postcode:             xbCoreResp.Postcode,
		AccountType:          xbCoreResp.AccountType,
		IdentificationType:   xbCoreResp.IdentificationType,
		IdentificationNumber: xbCoreResp.IdentificationNumber,
		BankAccountNumber:    xbCoreResp.BankAccountNumber,
		Dob:                  xbCoreResp.Dob,
		ContactCountryCode:   xbCoreResp.ContactCountryCode,
		ContactNumber:        xbCoreResp.ContactNumber,
		SourceOfIncome:       xbCoreResp.SourceOfIncome,
		CreatedAt:            xbCoreResp.CreatedAt,
		DeactivatedAt:        xbCoreResp.DeactivatedAt,
		UpdatedAt:            xbCoreResp.UpdatedAt,
	}, nil
}

func (s *xbPayoutService) DeactivateSender(ctx context.Context, request *xbModel.GetSenderByIdRequest) (*xbModel.CreateSenderResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/DeactivateSender")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.DeactivateSender(ctx, &xbCoreProcessorModel.GetSenderByIdRequest{
		MerchantId: request.MerchantId,
		SenderId:   request.SenderId,
	})
	if err != nil {
		s.logger.Error(ctx, "DeactivateSender - Failed to deactivate sender in xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateSenderResponse{
		UUID:                 xbCoreResp.UUID,
		Name:                 xbCoreResp.Name,
		CountryCode:          xbCoreResp.CountryCode,
		State:                xbCoreResp.State,
		City:                 xbCoreResp.City,
		Address:              xbCoreResp.Address,
		Postcode:             xbCoreResp.Postcode,
		AccountType:          xbCoreResp.AccountType,
		IdentificationType:   xbCoreResp.IdentificationType,
		IdentificationNumber: xbCoreResp.IdentificationNumber,
		BankAccountNumber:    xbCoreResp.BankAccountNumber,
		Dob:                  xbCoreResp.Dob,
		ContactCountryCode:   xbCoreResp.ContactCountryCode,
		ContactNumber:        xbCoreResp.ContactNumber,
		SourceOfIncome:       xbCoreResp.SourceOfIncome,
		CreatedAt:            xbCoreResp.CreatedAt,
		DeactivatedAt:        xbCoreResp.DeactivatedAt,
		UpdatedAt:            xbCoreResp.UpdatedAt,
	}, nil
}
