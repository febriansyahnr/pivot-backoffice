package xbPayoutService

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *xbPayoutService) CreateBeneficiary(ctx context.Context, request *xbModel.CreateBeneficiaryRequest) (*xbModel.CreateBeneficiaryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/CreateBeneficiary")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbBeneficiary, err := s.xbCoreProcessorRepo.CreateBeneficiary(ctx, &xbCoreProcessorModel.CreateBeneficiaryRequest{
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
		AccountNumber:        request.AccountNumber,
		BankName:             request.BankName,
		BankCode:             request.BankCode,
		ContactNumber:        request.ContactNumber,
		ContactCountryCode:   request.ContactCountryCode,
		Email:                request.Email,
		PayoutMethod:         strings.ToUpper(request.PayoutMethod),
	})
	if err != nil {
		s.logger.Error(ctx, "CreateBeneficiary - Failed to create beneficiary in xb-core-processor", logger.Error(err))
		return nil, err
	}

	var created bool
	if xbBeneficiary.UUID != uuid.Nil {
		created = true
	}

	beneficiaryAccountReq := &beneficiaryAccountModel.BeneficiaryAccount{
		UUID:                   xbBeneficiary.UUID.String(),
		BeneficiaryAccountNo:   xbBeneficiary.AccountNumber,
		BeneficiaryAccountName: xbBeneficiary.Name,
		BeneficiaryBankCode:    xbBeneficiary.BankCode,
		BeneficiaryBankName:    xbBeneficiary.BankName,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
		MerchantID:             request.MerchantId,
	}
	beneficiaryMetadata := beneficiaryAccountModel.Metadata{
		IsXb: true,
		XbDetail: &xbModel.BeneficiaryDataResponse{
			Name:               xbBeneficiary.Name,
			Address:            xbBeneficiary.Address,
			City:               xbBeneficiary.City,
			Postcode:           xbBeneficiary.Postcode,
			State:              xbBeneficiary.State,
			CountryCode:        xbBeneficiary.CountryCode,
			AccountType:        xbBeneficiary.AccountType,
			AccountNumber:      xbBeneficiary.AccountNumber,
			BankName:           xbBeneficiary.BankName,
			BankCode:           xbBeneficiary.BankCode,
			ContactCountryCode: xbBeneficiary.ContactCountryCode,
			ContactNumber:      xbBeneficiary.ContactNumber,
			Email:              xbBeneficiary.Email,
			PayoutMethod:       request.PayoutMethod,
		},
	}

	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		beneficiaryMetadata.OnBehalf = &merchantModel.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}

	beneficiaryAccountReq.Metadata.Valid = true
	beneficiaryAccountReq.Metadata.JSONText, _ = json.Marshal(beneficiaryMetadata)
	if err = s.beneficiaryAccountRepo.Upsert(ctx, beneficiaryAccountReq); err != nil {
		s.logger.Error(ctx, "CreateBeneficiary - Failed to upsert beneficiary account", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateBeneficiaryResponse{
		UUID:                 xbBeneficiary.UUID,
		Name:                 xbBeneficiary.Name,
		AccountType:          xbBeneficiary.AccountType,
		Address:              xbBeneficiary.Address,
		City:                 xbBeneficiary.City,
		Postcode:             xbBeneficiary.Postcode,
		State:                xbBeneficiary.State,
		CountryCode:          xbBeneficiary.CountryCode,
		IdentificationType:   xbBeneficiary.IdentificationType,
		IdentificationNumber: xbBeneficiary.IdentificationNumber,
		AccountNumber:        xbBeneficiary.AccountNumber,
		BankName:             xbBeneficiary.BankName,
		BankCode:             xbBeneficiary.BankCode,
		ContactCountryCode:   xbBeneficiary.ContactCountryCode,
		ContactNumber:        xbBeneficiary.ContactNumber,
		Email:                xbBeneficiary.Email,
		CreatedAt:            xbBeneficiary.CreatedAt,
		Created:              created,
	}, nil
}
func (s *xbPayoutService) GetListBeneficiary(ctx context.Context, request *xbModel.GetListBeneficiaryRequest) (*xbModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetListBeneficiary")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetListBeneficiary(ctx, &xbCoreProcessorModel.GetListBeneficiaryRequest{
		MerchantId:      request.MerchantId,
		Page:            request.Page,
		PerPage:         request.PerPage,
		FetchAll:        request.FetchAll,
		ShowDeactivated: request.ShowDeactivated,
		Name:            request.Name,
		CountryCode:     request.CountryCode,
		AccountNumber:   request.AccountNumber,
		AccountType:     request.AccountType,
	})
	if err != nil {
		s.logger.Error(ctx, "GetListBeneficiary - Failed to get list from xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.PaginationResponse{
		Results:    util.MapSnakeToCamel(xbCoreResp.Results),
		Pagination: xbModel.Pagination(xbCoreResp.Pagination),
	}, nil
}
func (s *xbPayoutService) GetBeneficiaryById(ctx context.Context, request *xbModel.GetBeneficiaryByIdRequest) (*xbModel.CreateBeneficiaryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetBeneficiaryById")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetBeneficiaryById(ctx, &xbCoreProcessorModel.GetBeneficiaryByIdRequest{
		MerchantId:    request.MerchantId,
		BeneficiaryId: request.BeneficiaryId,
	})
	if err != nil {
		s.logger.Error(ctx, "GetBeneficiaryById - Failed to get beneficiary by ID from xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateBeneficiaryResponse{
		UUID:                 xbCoreResp.UUID,
		Name:                 xbCoreResp.Name,
		AccountType:          xbCoreResp.AccountType,
		Address:              xbCoreResp.Address,
		City:                 xbCoreResp.City,
		Postcode:             xbCoreResp.Postcode,
		State:                xbCoreResp.State,
		CountryCode:          xbCoreResp.CountryCode,
		IdentificationType:   xbCoreResp.IdentificationType,
		IdentificationNumber: xbCoreResp.IdentificationNumber,
		AccountNumber:        xbCoreResp.AccountNumber,
		BankName:             xbCoreResp.BankName,
		BankCode:             xbCoreResp.BankCode,
		ContactCountryCode:   xbCoreResp.ContactCountryCode,
		ContactNumber:        xbCoreResp.ContactNumber,
		Email:                xbCoreResp.Email,
		CreatedAt:            xbCoreResp.CreatedAt,
		DeactivatedAt:        xbCoreResp.DeactivatedAt,
		UpdatedAt:            xbCoreResp.UpdatedAt,
	}, nil
}
func (s *xbPayoutService) UpdateBeneficiary(ctx context.Context, request *xbModel.UpdateBeneficiaryRequest) (*xbModel.CreateBeneficiaryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/UpdateBeneficiary")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.UpdateBeneficiary(ctx, &xbCoreProcessorModel.UpdateBeneficiaryRequest{
		BeneficiaryId:        request.BeneficiaryId,
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
		AccountNumber:        request.AccountNumber,
		BankName:             request.BankName,
		BankCode:             request.BankCode,
		ContactCountryCode:   request.ContactCountryCode,
		ContactNumber:        request.ContactNumber,
		Email:                request.Email,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateBeneficiary - Failed to update beneficiary in xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateBeneficiaryResponse{
		UUID:                 xbCoreResp.UUID,
		Name:                 xbCoreResp.Name,
		AccountType:          xbCoreResp.AccountType,
		Address:              xbCoreResp.Address,
		City:                 xbCoreResp.City,
		Postcode:             xbCoreResp.Postcode,
		State:                xbCoreResp.State,
		CountryCode:          xbCoreResp.CountryCode,
		IdentificationType:   xbCoreResp.IdentificationType,
		IdentificationNumber: xbCoreResp.IdentificationNumber,
		AccountNumber:        xbCoreResp.AccountNumber,
		BankName:             xbCoreResp.BankName,
		BankCode:             xbCoreResp.BankCode,
		ContactCountryCode:   xbCoreResp.ContactCountryCode,
		ContactNumber:        xbCoreResp.ContactNumber,
		Email:                xbCoreResp.Email,
		CreatedAt:            xbCoreResp.CreatedAt,
		DeactivatedAt:        xbCoreResp.DeactivatedAt,
		UpdatedAt:            xbCoreResp.UpdatedAt,
	}, nil
}
func (s *xbPayoutService) DeactivateBeneficiary(ctx context.Context, request *xbModel.GetBeneficiaryByIdRequest) (*xbModel.CreateBeneficiaryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/DeactivateBeneficiary")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.DeactivateBeneficiary(ctx, &xbCoreProcessorModel.GetBeneficiaryByIdRequest{
		MerchantId:    request.MerchantId,
		BeneficiaryId: request.BeneficiaryId,
	})
	if err != nil {
		s.logger.Error(ctx, "DeactivateBeneficiary - Failed to deactivate beneficiary in xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.CreateBeneficiaryResponse{
		UUID:                 xbCoreResp.UUID,
		Name:                 xbCoreResp.Name,
		AccountType:          xbCoreResp.AccountType,
		Address:              xbCoreResp.Address,
		City:                 xbCoreResp.City,
		Postcode:             xbCoreResp.Postcode,
		State:                xbCoreResp.State,
		CountryCode:          xbCoreResp.CountryCode,
		IdentificationType:   xbCoreResp.IdentificationType,
		IdentificationNumber: xbCoreResp.IdentificationNumber,
		AccountNumber:        xbCoreResp.AccountNumber,
		BankName:             xbCoreResp.BankName,
		BankCode:             xbCoreResp.BankCode,
		ContactCountryCode:   xbCoreResp.ContactCountryCode,
		ContactNumber:        xbCoreResp.ContactNumber,
		Email:                xbCoreResp.Email,
		CreatedAt:            xbCoreResp.CreatedAt,
		DeactivatedAt:        xbCoreResp.DeactivatedAt,
		UpdatedAt:            xbCoreResp.UpdatedAt,
	}, nil
}
