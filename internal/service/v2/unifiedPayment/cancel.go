package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *UnifiedPaymentService) CancelSession(ctx context.Context, request *unifiedPaymentModel.CancelUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/CancelSession")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentSessionID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if payment == nil {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	if payment.MerchantID != request.MerchantID {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)
	}

	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(payment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

	if err := s.checkCancellationEligibility(payment, unifiedPaymentMetadata.Mode, request.Source); err != nil {
		return nil, err
	}

	now := time.Now()

	metadata := map[string]interface{}{}
	if payment.Metadata != nil {
		metadata = *payment.Metadata
	}

	metadata["canceledAt"] = now
	metadata["cancellationReason"] = request.CancellationReason

	payment.Metadata = &metadata
	payment.Status = constant.UnifiedPaymentSessionStatusCancelled
	payment.UpdatedAt = now
	payment.ExpiredAt = &now

	metadataStr, err := json.Marshal(metadata)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}

	metadataString := string(metadataStr)
	paymentDTO := &paymentModel.PaymentDTO{
		UUID:       payment.UUID,
		MerchantID: payment.MerchantID,
		Status:     constant.UnifiedPaymentSessionStatusCancelled,
		UpdatedAt:  now,
		ExpiredAt:  &now,
		Metadata:   &metadataString,
	}

	err = s.paymentRepo.UpdatePaymentData(ctx, paymentDTO)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	return payment.ToUnifiedPaymentResponse(), nil
}

func (s *UnifiedPaymentService) checkCancellationEligibility(payment *paymentModel.Payment, paymentMode, source string) error {
	cancelAllowed := false
	cancelReason := fmt.Sprintf("Payment session cannot be cancelled because it is in %s status", payment.Status)

	switch payment.Status {
	case constant.UnifiedPaymentSessionStatusRequirePaymentMethod, constant.UnifiedPaymentSessionStatusRequireConfirmation:
		cancelAllowed = true
	case constant.UnifiedPaymentSessionStatusRequireAction:
		if paymentMode == constant.UnifiedPaymentModeRedirect && source == "CUSTOMER" {
			cancelAllowed = true
		} else {
			cancelReason = "Payment session in REQUIRE_ACTION status can only be cancelled by customer on payment page when using redirection method"
		}
	case constant.UnifiedPaymentSessionStatusProcessing, constant.UnifiedPaymentSessionStatusPaid,
		constant.UnifiedPaymentSessionStatusExpired, constant.UnifiedPaymentSessionStatusCancelled:
		cancelReason = fmt.Sprintf("Payment session in %s status cannot be cancelled", payment.Status)
		return pkgErrors.New(response.HttpErrUnprocessableContent, errors.New(cancelReason))
	}

	if !cancelAllowed {
		return pkgErrors.New(response.HttpErrRequest, errors.New(cancelReason))
	}

	return nil
}
