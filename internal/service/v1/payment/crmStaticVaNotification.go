package paymentService

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) CRMStaticVARetryNotification(ctx context.Context, payload *paymentModel.CRMStaticVARetryNotificationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/CRMStaticVARetryNotification")
	defer segment.End()

	s.logger.Info(ctx, "CRM - start retry static VA payment notification", logger.String("vaNumber", payload.VANumber), logger.String("amount", payload.Amount.Value))

	if payload.VANumber == "" {
		err := pkgErrors.New(response.HttpErrRequest, errors.New("va number is required"))
		s.logger.Error(ctx, "va number is required")
		return err
	}

	if payload.Amount.Value == "" {
		err := pkgErrors.New(response.HttpErrRequest, errors.New("amount value is required"))
		s.logger.Error(ctx, "amount value is required")
		return err
	}

	if payload.Amount.Currency == "" {
		err := pkgErrors.New(response.HttpErrRequest, errors.New("amount currency is required"))
		s.logger.Error(ctx, "amount currency is required")
		return err
	}

	if err := s.snapCoreRepo.PublishPayment(ctx, snapPaymentModel.PublishRequest{
		InternalReference: payload.VANumber,
		PaymentMethod:     constant.UnifiedPaymentMethodVA,
		ForceSuccess:      true,
		Amount: commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    payload.Amount.Value,
		},
	}); err != nil {
		s.logger.Error(ctx, "error when publish payment", logger.Error(err))
		return err
	}

	s.logger.Info(ctx, "CRM - successfully published static VA payment notification", logger.String("vaNumber", payload.VANumber))

	return nil
}
