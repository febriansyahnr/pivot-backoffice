package paymentService

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapModelVA "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) UpdateUnifiedPayment(ctx context.Context, request *paymentModel.UpdateUnifiedPaymentRequest) (resp *paymentModel.UpdateUnifiedPaymentResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/UpdateUnifiedPayment")
	defer segment.End()

	existingPayment, err := s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.MerchantID, request.ClientReferenceID)
	if err != nil {
		return nil, err
	} else if existingPayment == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	if !s.IsUpdateEligible(existingPayment) {
		return nil, fmt.Errorf("payment session is either expired or in a non-modifiable state")
	} else if !slices.Contains(allowedPaymentMethods, request.PaymentMethod) {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("invalid payment method: %s", request.PaymentMethod))
	}

	// Change Payment Method
	if s.AllowedToUpdateMethodOrPaymentOption(ctx, request, existingPayment) {
		supported := false
		switch existingPayment.PaymentMethod.Type {
		case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
			supported = s.config.UnifiedPaymentConfig.ChangePaymentMethodFromVA

		case paymentConstant.PAYMENT_METHOD_QRIS:
			supported = s.config.UnifiedPaymentConfig.ChangePaymentMethodFromQris

		case paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
			supported = s.config.UnifiedPaymentConfig.ChangePaymentMethodFromCreditCard
		}
		if !supported {
			return nil, pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("change of payment method is not supported"))
		}
		if request.PaymentMethod == paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
			invalidMethodOpts := request.PaymentMethodOptions == nil ||
				request.PaymentMethodOptions.Card == nil ||
				request.PaymentMethodOptions.Card.AuthenticationMethod == ""
			if invalidMethodOpts {
				return nil, pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("payment method options is required"))
			}
		}

		if existingPayment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_QRIS {
			_, err := s.snapCoreRepo.CancelQrMpm(ctx, *existingPayment.SnapCoreId)
			if err != nil {
				s.logger.Error(ctx, "error cancelling QRIS payment: %v", logger.Error(err))
				return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrCannotCancelQRISPayment)
			}
		}

		ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
			OriginId:    existingPayment.UUID,
			ReferenceId: existingPayment.MerchantID,
			From:        serviceName,
		})
		ctx = context.WithValue(ctx, constant.CtxChangePaymentMethod, true)
		ctx = context.WithValue(ctx, constant.CtxCurrentPaymentMethod, existingPayment.PaymentMethod.Type)

		// Record payment method change
		s.RecordPaymentStatusHistory(ctx, existingPayment.UUID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryRequirePaymentMethod)

		url, _ := url.Parse(existingPayment.PaymentURL)
		hashBeforeUpdate := util.HashString(url.Query().Get("token"))

		defer func() {
			if err != nil || resp == nil {
				return
			}

			switch existingPayment.PaymentMethod.Type {
			case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
				request := &snapModelVA.DeleteVirtualAccountRequest{
					UUID:   *existingPayment.SnapCoreId,
					Number: *existingPayment.ProcessorReferenceNumber,
				}
				_, _ = s.snapCoreRepo.DeleteVirtualAccount(ctx, request)
			default:
				// Currently only support changing payment methods from VA
			}

			url2, _ := url.Parse(resp.PaymentLink)
			if hash := util.HashString(url2.Query().Get("token")); hash != hashBeforeUpdate {
				_ = s.redis.Del(ctx, fmt.Sprintf(constant.PaymentTokenCacheKey, hashBeforeUpdate))
			}
		}()
		return s.changePaymentMethod(ctx, request, existingPayment)
	}

	paymentUpdateRequest := &paymentModel.PaymentUpdateRequest{
		MerchantId:   existingPayment.MerchantID,
		PaymentId:    existingPayment.UUID,
		TotalAmount:  request.Amount,
		ExpiredAt:    request.ExpiryAt,
		AccountEmail: request.Customer.Email,
		AccountPhone: request.Customer.Phone,
		AccountName:  request.Customer.Name,
	}

	// Save the updated payment session
	// TODO: discuss and rename the func due its for VA Only
	result, err := s.UpdatePayment(ctx, paymentUpdateRequest)
	if err != nil {
		return nil, fmt.Errorf("could not save payment updates: %w", err)
	}

	resp = &paymentModel.UpdateUnifiedPaymentResponse{
		ID:                result.UUID,
		ClientReferenceID: result.ReferenceID,
		PaymentMethod:     result.PaymentMethod,
		Amount:            *result.TotalAmount,
		Customer:          result.Customer,
	}

	switch request.PaymentMethod {
	case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		resp.ExpiryAt = *result.VirtualAccount.ExpiredDate

	case paymentConstant.PAYMENT_METHOD_QRIS:
		resp.ExpiryAt = *request.ExpiryAt

	case paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
		resp.ExpiryAt = *request.ExpiryAt
	}
	return resp, nil
}

// Helper function to check if the payment session is eligible for updates
// The payment session is eligible for updates if:
// - Payment method is credit card and status is pending
// - Payment method is not credit card and status is waiting for payment
// - Payment session is not expired
func (s *PaymentService) IsUpdateEligible(payment *paymentModel.Payment) bool {
	if time.Now().After(*payment.ExpiredAt) {
		return false
	}

	// should check it if cc change the payment status
	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_CREDIT_CARD {
		return payment.Status == paymentConstant.PAYMENT_STATUS_PENDING
	}

	return payment.Status == paymentConstant.UnifiedPaymentStatusWaitingForPayment
}

// AllowedToUpdateMethodOrPaymentOption determines whether a payment's method or options are allowed to be updated.
// The function checks several conditions:
// 1. If the requested payment method is different from the existing one, update is allowed.
// 2. For Virtual Account payment method, update is allowed only if the requested issuer is different from the existing acquirer.
func (s *PaymentService) AllowedToUpdateMethodOrPaymentOption(ctx context.Context, req *paymentModel.UpdateUnifiedPaymentRequest, existingPayment *paymentModel.Payment) bool {
	if req.PaymentMethod != existingPayment.PaymentMethod.Type {
		return true
	}

	if (req.PaymentMethod == paymentConstant.PAYMENT_METHOD_QRIS || req.PaymentMethod == paymentConstant.PAYMENT_METHOD_CREDIT_CARD) &&
		(req.PaymentMethod == existingPayment.PaymentMethod.Type) {
		return true
	}

	if req.PaymentMethod == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
		metadata, err := util.ConvertToStruct[orchestrator_model.MetadataPaymentMethodVA]((*existingPayment.Metadata)["snapCore"])
		if err != nil {
			s.logger.Warn(ctx, "error converting metadata to VA: %v", logger.Error(err))
			return false
		}

		return !strings.EqualFold(req.PaymentMethodOptions.VirtualAccount.Issuer, metadata.Acquirer)
	}

	return false
}
