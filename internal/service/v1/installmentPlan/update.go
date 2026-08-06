package installmentplan

import (
	"context"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *InstallmentPlanService) Update(ctx context.Context, request *installmentPlanModel.UpdateInstallmentPlanRequest) (*installmentPlanModel.InstallmentPlan, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/installmentPlan/Update")
	defer span.End()

	installmentPlan, err := s.repo.GetById(ctx, request.UUID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrGetInstallmentPlan)
	}
	if installmentPlan == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrInstallmentPlanNotFound)
	}
	if (request.Acquirer != "" && request.Acquirer != installmentPlan.Acquirer) ||
		(request.Status != "" && request.Status == constant.InstallmentPlanStatusInactive) {
		// Validate if there were existing merchant payment config use this installment plan
		paymentMethods, err := s.paymentMethodSvc.GetPaymentMethodByMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
			Type:     paymentConstant.PAYMENT_METHOD_INSTALLMENT,
			Subtype:  constant.InstallmentPlanPaymentMethodCard,
			Acquirer: strings.ToLower(installmentPlan.Acquirer),
			Category: paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
			InstallmentPlan: paymentModel.InstallmentPlanFilterRequest{
				InstallmentPlanID: installmentPlan.UUID,
			},
		})
		if err != nil {
			s.logger.Error(ctx, "error when get merchant payment method which has installment plan id", logger.Error(err), logger.Any("request", request))
			return nil, err
		}
		existingPaymentMethods := []string{}
		for _, paymentMethod := range paymentMethods {
			existingPaymentMethods = append(existingPaymentMethods, paymentMethod.UUID)
		}
		if len(existingPaymentMethods) > 0 {
			return nil, pkgErrors.New(response.HttpErrUnprocessableContent, fmt.Errorf("merchant with payment method %v is using this installment plan", strings.Join(existingPaymentMethods, ", ")))
		}
	}

	err = installmentPlan.Update(request)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}
	if installmentPlan.Status == constant.InstallmentPlanStatusActive {
		mid, err := s.validateCardInstallment(ctx, &installmentPlanModel.ValidateCardInstallmentPlanRequest{
			MidId:          installmentPlan.PlanMetadata.Card.MidID,
			Tenor:          installmentPlan.Tenor,
			SettlementType: installmentPlan.SettlementType,
			AllowedBins:    installmentPlan.PlanMetadata.Card.AllowedBins,
		})
		if err != nil {
			s.logger.Error(ctx, "error when validate card installment", logger.Error(err), logger.Any("installmentPlan", installmentPlan))
			return nil, err
		}
		installmentPlan.UpdateMIDInfo(mid)
	}

	// TODO:
	// IF installment set to inactive, make sure there is no merchant payment config use installment plan
	// (for merchant assigned installment plan, could specific search to merchant)
	// filter by merchant (optional), settlement, acquirer, card payment method, tenor

	err = s.repo.Update(ctx, installmentPlan)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateInstallmentPlan)
	}

	return installmentPlan, nil
}
