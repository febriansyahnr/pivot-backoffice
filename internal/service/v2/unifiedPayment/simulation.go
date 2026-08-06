package unifiedPaymentService

import (
	"context"
	"fmt"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *UnifiedPaymentService) SimulatePayment(ctx context.Context, request *unifiedPaymentModel.SimulatePaymentRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/SimulatePayment")
	defer span.End()

	if s.config.Environment == constant.EnvironmentProduction {
		s.logger.Warn(ctx, "can't simulate payment on production environment")
		return pkgErr.New(httpResponse.HttpErrForbidden, constant.ErrForbiddenAccess)
	}

	// Find payment by paymentID
	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentSessionID)
	if err != nil {
		return pkgErr.New(httpResponse.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Warn(ctx, "[SimulatePayment] Payment not found")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	} else if payment.MerchantID != request.MerchantID {
		s.logger.Warn(ctx, "[SimulatePayment] Merchant ID mismatch")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)
	} else if !slices.Contains([]string{constant.UnifiedPaymentSessionStatusRequireAction, constant.UnifiedStaticPaymentStatusActive}, payment.Status) {
		s.logger.Warn(ctx, "[SimulatePayment] Payment session status not require action or active")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrStatusNotAllowed)
	}

	// Find charge by paymentID and chargeID
	var paymentCharge *orchestratorModel.AccountTransactionWithUseCase
	if request.ChargeID != "" {
		paymentCharge, err = s.accountTransactionRepo.FindByID(ctx, request.ChargeID)
	} else {
		paymentCharge, err = s.accountTransactionRepo.FindByReference(ctx, request.PaymentSessionID, constant.TypePayment)
	}

	if err != nil {
		return pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	currency := payment.Currency
	amount := request.Amount.Value
	if payment.Status == constant.UnifiedPaymentSessionStatusRequireAction {
		if paymentCharge == nil {
			s.logger.Warn(ctx, "[SimulatePayment] Payment charge not found")
			return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
		} else if paymentCharge.ReferenceID != request.PaymentSessionID {
			s.logger.Warn(ctx, "[SimulatePayment] Payment charge is not match with payment session id")
			return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
		}

		currency = paymentCharge.Currency
		amount = paymentCharge.Credit
	}

	if !slices.Contains([]string{
		paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		paymentConstant.PAYMENT_METHOD_QRIS,
		paymentConstant.PAYMENT_METHOD_EWALLET,
	}, payment.PaymentMethod.Type) {
		s.logger.Warn(ctx, "[SimulatePayment] Payment method type is not allowed to simulate")
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentMethodNotAllowed)
	}

	paidAmount := commonModel.Amount{
		Currency: currency,
		Value:    fmt.Sprintf("%.2f", amount),
	}

	switch request.ChargeStatus {
	case constant.ChargeStatusProcessing:
		s.logger.Info(ctx, "[SimulatePayment] Payment charge is PROCESSING")
		return s.paymentSvc.ProcessPaymentForSimulationByID(ctx, payment.UUID, paidAmount, constant.ChargeStatusProcessing)
	case constant.ChargeStatusExpired:
		s.logger.Info(ctx, "[SimulatePayment] Payment charge is EXPIRED")
		return s.paymentSvc.ProcessPaymentForSimulationByID(ctx, payment.UUID, paidAmount, constant.ChargeStatusExpired)
	case constant.ChargeStatusSuccess:
		s.logger.Info(ctx, "[SimulatePayment] Payment charge is SUCCESS")
		return s.paymentSvc.ProcessPaymentForSimulationByID(ctx, payment.UUID, paidAmount, constant.ChargeStatusSuccess)
	case constant.ChargeStatusFailed:
		if payment.PaymentMethod.Type != paymentConstant.PAYMENT_METHOD_EWALLET {
			s.logger.Warn(ctx, "[SimulatePayment] Payment method type is not allowed to simulate for FAILED status")
			return pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentMethodNotAllowed)
		}

		s.logger.Info(ctx, "[SimulatePayment] Payment charge is FAILED")
		return s.paymentSvc.ProcessPaymentForSimulationByID(ctx, payment.UUID, paidAmount, constant.ChargeStatusFailed)
	}

	return nil
}
