package paymentService

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) CRMRetryNotification(ctx context.Context, payload *paymentModel.CRMRetryNotificationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/CRMRetryNotification")
	defer segment.End()

	s.logger.Info(ctx, "CRM - start retry payment", logger.String("id", payload.ID), logger.String("bankReference", payload.BankReference))

	payment, err := s.paymentRepo.GetPaymentById(ctx, payload.ID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment data by id", logger.Error(err))
		return err
	}

	if payment == nil {
		err := pkgErrors.New(response.HttpErrNotFound, errors.New("payment not found"))
		s.logger.Error(ctx, "payment not found", logger.String("id", payload.ID))
		return err
	}

	paymentMethod, err := s.paymentMethodSvc.FindPaymentMethodByIdAndMerchant(ctx, payment.PaymentMethodID, payment.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method by id and merchant", logger.Error(err))
		return err
	}

	if paymentMethod == nil {
		err := pkgErrors.New(response.HttpErrNotFound, errors.New("payment method not found"))
		s.logger.Error(ctx, "payment method not found", logger.String("id", payload.ID))
		return err
	}

	paymentMethodType := paymentMethod.Type
	internalReference := util.ValueOfPtr(payment.ProcessorReferenceNumber)

	if paymentMethod.Type == constant.UnifiedPaymentMethodQris ||
		paymentMethod.Type == paymentConstant.PAYMENT_METHOD_QRIS {

		paymentMethodType = paymentConstant.PAYMENT_METHOD_QRIS
		internalReference = util.ValueOfPtr(payment.ReferenceID)
	}

	if err = s.snapCoreRepo.PublishPayment(ctx, snapPaymentModel.PublishRequest{
		InternalReference: internalReference,
		PaymentMethod:     paymentMethodType,
		ForceSuccess:      payload.ForceSuccess,
		Amount: commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    payment.Amount.String(),
		},
	}); err != nil {
		s.logger.Error(ctx, "error when publish payment", logger.Error(err))
		return err
	}

	return nil
}
