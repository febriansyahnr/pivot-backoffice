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
	"github.com/shopspring/decimal"
)

func (s *xbPayoutService) GetPayoutById(ctx context.Context, request *xbModel.GetPayoutRequest) (*xbModel.GetPayoutResponse, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetPayoutById")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    request.PayoutId,
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	// Find payout by ID
	disbursement, err := s.disbursementRepo.FindByID(ctx, request.PayoutId)
	if err != nil {
		s.logger.Error(ctx, "GetPayoutById - Failed to find disbursement by ID", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)

	} else if disbursement == nil || disbursement.MerchantID != request.MerchantId || disbursement.MetadataObj.XbDetail == nil {
		s.logger.Info(ctx, "GetPayoutById - Payout not found or merchant ID does not match or missing XB metadata", logger.Any("request", map[string]string{
			"merchantId": request.MerchantId,
			"payoutId":   request.PayoutId,
		}))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	// Get new status from XB core processor
	xbResp, err := s.xbCoreProcessorRepo.GetPayoutById(ctx, &xbCoreProcessorModel.GetPayoutRequest{
		Id:         disbursement.MetadataObj.XbDetail.Uuid,
		MerchantId: disbursement.MerchantID,
	})
	if err != nil {
		s.logger.Error(ctx, "GetPayoutById - Failed to get payout detail from xb-core-processor", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, err)
	}

	// Use stored final fee amount (already calculated with FX conversion during creation)
	finalFeeAmount := disbursement.MetadataObj.FeeDetail.FinalAmount

	response := &xbModel.GetPayoutResponse{
		Uuid:                disbursement.UUID,
		MerchantId:          disbursement.MerchantID,
		ReferenceId:         disbursement.ReferenceID,
		SourceCurrency:      xbResp.SourceCurrency,
		DestinationCurrency: xbResp.DestinationCurrency,
		DestinationAmount:   xbResp.DestinationAmount,
		FxRate:              xbResp.FxRate,
		DestinationFxRate:   xbResp.DestinationFXRate,
		Fee:                 decimal.NewFromFloat(finalFeeAmount),
		TotalAmount:         xbResp.TotalAmount,
		Remark:              xbResp.StatementNarrative,
		CreatedAt:           xbResp.CreatedAt,
		BeneficiaryData: xbModel.BeneficiaryDataResponse{
			Name:               xbResp.BeneficiaryData.Name,
			Address:            xbResp.BeneficiaryData.Address,
			City:               xbResp.BeneficiaryData.City,
			Postcode:           xbResp.BeneficiaryData.Postcode,
			State:              xbResp.BeneficiaryData.State,
			CountryCode:        xbResp.BeneficiaryData.CountryCode,
			CountryName:        xbResp.BeneficiaryData.CountryName,
			AccountType:        xbResp.BeneficiaryData.AccountType,
			AccountNumber:      xbResp.BeneficiaryData.AccountNumber,
			BankName:           xbResp.BeneficiaryData.BankName,
			BankCode:           xbResp.BeneficiaryData.BankCode,
			ContactCountryCode: xbResp.BeneficiaryData.ContactCountryCode,
			ContactNumber:      xbResp.BeneficiaryData.ContactNumber,
			Email:              xbResp.BeneficiaryData.Email,
		},
		SenderData: xbModel.SenderDataResponse{
			Name:                 xbResp.SenderData.Name,
			Address:              xbResp.SenderData.Address,
			City:                 xbResp.SenderData.City,
			Postcode:             xbResp.SenderData.Postcode,
			State:                xbResp.SenderData.State,
			CountryCode:          xbResp.SenderData.CountryCode,
			CountryName:          xbResp.SenderData.CountryName,
			AccountType:          xbResp.SenderData.AccountType,
			IdentificationType:   xbResp.SenderData.IdentificationType,
			IdentificationNumber: xbResp.SenderData.IdentificationNumber,
			BankAccountNumber:    xbResp.SenderData.BankAccountNumber,
			ContactCountryCode:   xbResp.SenderData.ContactCountryCode,
			ContactNumber:        xbResp.SenderData.ContactNumber,
			Dob:                  xbResp.SenderData.Dob,
			SourceOfIncome:       xbResp.SenderData.SourceOfIncome,
		},
		Status:            s.mapStatus(xbResp.Status),
		StatusDescription: xbResp.StatusDescription,
		RoutingCode:       xbResp.RoutingCode,
		RoutingValue:      xbResp.RoutingValue,
	}

	var RfiDetails []*xbModel.RfiDetails
	if xbResp.RfiDetails != nil {
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
	}
	response.RfiDetails = RfiDetails

	return response, nil
}
