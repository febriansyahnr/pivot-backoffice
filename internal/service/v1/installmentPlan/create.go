package installmentplan

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *InstallmentPlanService) Create(ctx context.Context, request *installmentPlanModel.CreateInstallmentPlanRequest) (*installmentPlanModel.InstallmentPlan, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/installmentPlan/Create")
	defer span.End()

	if request.MerchantID != "" {
		merchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
		if err != nil {
			return nil, err
		}
		if merchant == nil {
			return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
		}
	}

	if request.PaymentMethod == constant.InstallmentPlanPaymentMethodCard && request.CardDetail != nil {
		midDetail, err := s.validateCardInstallment(ctx, &installmentPlanModel.ValidateCardInstallmentPlanRequest{
			MidId:          request.CardDetail.MidID,
			Tenor:          request.Tenor,
			SettlementType: request.SettlementType,
			AllowedBins:    request.CardDetail.AllowedBins,
		})
		if err != nil {
			return nil, err
		}
		request.CardDetail.Mid = midDetail.Mid
		request.CardDetail.MidInstallmentType = midDetail.InstallmentType
	}

	installmentPlan := installmentPlanModel.New(request)
	err := s.repo.Create(ctx, installmentPlan)
	if err != nil {
		s.logger.Error(ctx, "error create new installment plan", logger.Error(err), logger.Any("request", request))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrCreateInstallmentPlan)
	}
	return installmentPlan, nil
}

func (s *InstallmentPlanService) validateCardInstallment(ctx context.Context, request *installmentPlanModel.ValidateCardInstallmentPlanRequest) (*creditcardCoreProcessorModel.MIDResponseData, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/installmentPlan/validateCardInstallment")
	defer span.End()

	midDetail, err := s.creditCardSvc.GetMIDDetail(ctx, request.MidId)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetMIDDetail)
	}
	if request.SettlementType != midDetail.Type {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrInvalidMIDSettlementType)
	}
	if midDetail.TransactionType != "" && midDetail.TransactionType == constant.CreditCardMidTransactionTypeDirectPay {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrInvalidMIDTransactionType)
	}
	if midDetail.InstallmentTenor != request.Tenor {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMismatchInstallmentTenor)
	}

	err = s.creditCardSvc.ValidateMIDInstallmentBins(ctx, &card.ValidateMIDInstallmentBinsRequest{
		MidID: request.MidId,
		Bins:  request.AllowedBins,
	})
	if err != nil {
		return nil, err
	}

	return midDetail, nil
}
