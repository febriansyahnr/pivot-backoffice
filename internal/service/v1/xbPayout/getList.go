package xbPayoutService

import (
	"context"
	errs "errors"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *xbPayoutService) GetList(ctx context.Context, filter *xbModel.GetPayoutFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetList")
	defer segment.End()

	disbursementFilter := &disbursementModel.GetDisbursementFilterRequest{
		MerchantID:     filter.MerchantID,
		UUID:           filter.UUID,
		StartCreatedAt: &filter.StartAt,
		EndCreatedAt:   &filter.EndAt,
		ReasonType:     filter.Status,
		SortBy:         filter.SortBy,
		Sort:           filter.Sort,
		IsXbPayout:     true,
	}

	list, err := s.disbursementRepo.GetList(ctx, disbursementFilter, page, perPage)
	if err != nil {
		s.logger.Error(ctx, "GetList - Failed to get disbursement list", logger.Error(err))
		return nil, err
	}

	resp, err := buildDataResponse(list.Data)
	if err != nil {
		s.logger.Error(ctx, "GetList - Failed to build data response", logger.Error(err))
		return nil, err
	}

	return &commonModel.PaginationResponse{
		Data: resp,
		Meta: list.Meta,
	}, nil
}

func buildDataResponse(listData interface{}) (resp []*xbModel.GetPayoutResponse, err error) {
	disbursementDataList, ok := listData.([]*disbursementModel.DisbursementWithTransactionResponse)
	if !ok {
		// Tidak ada s.logger disini karena buildDataResponse function biasa (bukan receiver struct),
		// jadi cukup return error seperti ini:
		return nil, pkgErrors.New(response.HttpErrInternal, errs.New("failed to map response"))
	}

	tempResp := make([]*xbModel.GetPayoutResponse, len(disbursementDataList))
	for idx, disbursement := range disbursementDataList {
		status := ""
		if disbursement.ReasonType != nil {
			status = *disbursement.ReasonType
		}

		statusDesc := ""
		if disbursement.ReasonDescription != nil {
			statusDesc = *disbursement.ReasonDescription
		}

		remark := ""
		if disbursement.Remark != nil {
			remark = *disbursement.Remark
		}

		tempResp[idx] = &xbModel.GetPayoutResponse{
			Uuid:                disbursement.UUID,
			MerchantId:          disbursement.MerchantID,
			ReferenceId:         disbursement.ReferenceID,
			SourceCurrency:      disbursement.MetadataObj.XbDetail.SourceCurrency,
			DestinationCurrency: disbursement.Currency,
			DestinationAmount:   disbursement.Amount,
			FxRate:              disbursement.MetadataObj.XbDetail.FxRate,
			DestinationFxRate:   disbursement.MetadataObj.XbDetail.DestinationFxRate,
			Fee:                 decimal.NewFromFloat(disbursement.MetadataObj.FeeDetail.FinalAmount),
			TotalAmount:         disbursement.TotalAmount,
			Remark:              remark,
			CreatedAt:           disbursement.CreatedAt,
			BeneficiaryData: xbModel.BeneficiaryDataResponse{
				Name:               disbursement.MetadataObj.XbDetail.BeneficiaryData.Name,
				Address:            disbursement.MetadataObj.XbDetail.BeneficiaryData.Address,
				City:               disbursement.MetadataObj.XbDetail.BeneficiaryData.City,
				Postcode:           disbursement.MetadataObj.XbDetail.BeneficiaryData.Postcode,
				State:              disbursement.MetadataObj.XbDetail.BeneficiaryData.State,
				CountryCode:        disbursement.MetadataObj.XbDetail.BeneficiaryData.CountryCode,
				CountryName:        disbursement.MetadataObj.XbDetail.BeneficiaryData.CountryName,
				AccountType:        disbursement.MetadataObj.XbDetail.BeneficiaryData.AccountType,
				AccountNumber:      disbursement.MetadataObj.XbDetail.BeneficiaryData.AccountNumber,
				BankName:           disbursement.MetadataObj.XbDetail.BeneficiaryData.BankName,
				BankCode:           disbursement.MetadataObj.XbDetail.BeneficiaryData.BankCode,
				ContactCountryCode: disbursement.MetadataObj.XbDetail.BeneficiaryData.ContactCountryCode,
				ContactNumber:      disbursement.MetadataObj.XbDetail.BeneficiaryData.ContactNumber,
				Email:              disbursement.MetadataObj.XbDetail.BeneficiaryData.Email,
			},
			SenderData: xbModel.SenderDataResponse{
				Name:                 disbursement.MetadataObj.XbDetail.SenderData.Name,
				Address:              disbursement.MetadataObj.XbDetail.SenderData.Address,
				City:                 disbursement.MetadataObj.XbDetail.SenderData.City,
				Postcode:             disbursement.MetadataObj.XbDetail.SenderData.Postcode,
				State:                disbursement.MetadataObj.XbDetail.SenderData.State,
				CountryCode:          disbursement.MetadataObj.XbDetail.SenderData.CountryCode,
				CountryName:          disbursement.MetadataObj.XbDetail.SenderData.CountryName,
				AccountType:          disbursement.MetadataObj.XbDetail.SenderData.AccountType,
				IdentificationType:   disbursement.MetadataObj.XbDetail.SenderData.IdentificationType,
				IdentificationNumber: disbursement.MetadataObj.XbDetail.SenderData.IdentificationNumber,
				BankAccountNumber:    disbursement.MetadataObj.XbDetail.SenderData.BankAccountNumber,
				ContactCountryCode:   disbursement.MetadataObj.XbDetail.SenderData.ContactCountryCode,
				ContactNumber:        disbursement.MetadataObj.XbDetail.SenderData.ContactNumber,
				Dob:                  disbursement.MetadataObj.XbDetail.SenderData.Dob,
				SourceOfIncome:       disbursement.MetadataObj.XbDetail.SenderData.SourceOfIncome,
			},
			Status:            status,
			StatusDescription: statusDesc,
			RoutingCode:       disbursement.MetadataObj.XbDetail.RoutingCode,
			RoutingValue:      disbursement.MetadataObj.XbDetail.RoutingValue,
		}
	}

	return tempResp, nil
}
