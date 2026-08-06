package paymentService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) InquiryPayment(ctx context.Context, request *paymentModel.InquiryPaymentRequest) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/InquiryPayment")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if payment == nil {
		s.logger.Info(ctx, constant.ErrPaymentNotFound.Error(), logger.String("paymentId", request.PaymentID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	if payment.Status == constant.UnifiedPaymentSessionStatusProcessing || payment.Status == constant.UnifiedPaymentSessionStatusRequireAction {
		if payment.PaymentMethod.Type == constant.UnifiedPaymentMethodEWallet {
			payment, err = s.unifiedPaymentSvc.InquiryEWalletPayment(ctx, payment)
			if err != nil {
				return nil, err
			}
		}
	}

	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_CREDIT_CARD && payment.Status == constant.UnifiedPaymentSessionStatusProcessing {
		payment, err = s.unifiedPaymentSvc.InquiryCardPayment(ctx, payment)
		if err != nil {
			return nil, err
		}
	}

	return payment, nil
}
