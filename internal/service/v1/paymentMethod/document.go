package paymentMethodService

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PaymentMethodService) GetRequiredMerchantDocuments(ctx context.Context, request *paymentMethodModel.GetRequiredMerchantDocumentsRequest) (*[]paymentMethodModel.MerchantRequiredDocumentsResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/GetRequiredMerchantDocuments")
	defer segment.End()

	merchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return nil, err

	} else if merchant == nil {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantShouldKYC)
	}

	paymentMethod, err := s.FindPaymentMethodByIdAndMerchant(ctx, request.PaymentMethodID, request.MerchantID)
	if err != nil {
		return nil, err

	}

	if paymentMethod.RequiredDocumentObjects == nil || len(*paymentMethod.RequiredDocumentObjects) == 0 {
		return nil, nil
	}

	resp := make([]paymentMethodModel.MerchantRequiredDocumentsResponse, 0)
	for _, document := range *paymentMethod.RequiredDocumentObjects {
		merchantDocumentResp := paymentMethodModel.MerchantRequiredDocumentsResponse{
			Name:                   document.Name,
			Format:                 document.Format,
			MerchantDocumentID:     "",
			MerchantDocumentStatus: constant.MerchantDocumentStatusNotSubmitted,
		}

		if merchantDocument, errDoc := s.merchantSvc.FindDocumentByType(ctx, request.MerchantID, document.Name); errDoc == nil && merchantDocument != nil {
			merchantDocumentResp.MerchantDocumentID = merchantDocument.Id
			merchantDocumentResp.MerchantDocumentStatus = merchantDocument.Status
		}

		if merchantDocumentResp.MerchantDocumentStatus == constant.StatusApproved {
			merchantDocumentResp.MerchantDocumentStatus = constant.MerchantDocumentStatusSubmitted
		}

		resp = append(resp, merchantDocumentResp)
	}

	return &resp, nil
}
