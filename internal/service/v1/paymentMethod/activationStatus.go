package paymentMethodService

import (
	"context"
	"fmt"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PaymentMethodService) ChangeActivationStatus(ctx context.Context, request *paymentMethodModel.ChangeActivationStatusRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/Activate")
	defer segment.End()

	merchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return err

	} else if merchant == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantShouldKYC)
	}

	paymentMethod, err := s.FindPaymentMethodByIdAndMerchant(ctx, request.PaymentMethodID, request.MerchantID)
	if err != nil {
		return err

	}

	if paymentMethod.ActivationStatus == constant.PaymentMethodActivationStatusApproved {
		s.logger.Info(ctx, "Change activation status not allowed for status APPROVED")
		return pkgErrors.New(response.HttpErrUnprocessableContent,
			fmt.Errorf("change activation status not allowed from %s to %s", paymentMethod.ActivationStatus, request.Status))
	}

	paymentMethod.ActivationStatus = request.Status
	return s.paymentMethodRepo.UpsertPaymentMethodMerchantByIdAndMerchant(ctx, paymentMethod)
}
