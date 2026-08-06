package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// ResendPaymentCallback resends the callback for a unified payment v2 transaction
func (s *UnifiedPaymentService) ResendPaymentCallback(ctx context.Context, request *callbackModel.ResendCallbackRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/ResendPaymentCallback")
	defer segment.End()

	var (
		payment *paymentModel.Payment
		err     error
	)

	if request.ReferenceID != "" {
		// Load payment by ID
		payment, err = s.paymentRepo.GetPaymentById(ctx, request.ReferenceID)
		if err != nil {
			s.logger.Error(ctx, "[ResendPaymentCallback] Failed to get payment by ID", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
	} else {
		// Load payment by client reference id
		payment, err = s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.MerchantID, request.ClientReferenceID)
		if err != nil {
			s.logger.Error(ctx, "[ResendPaymentCallback] Failed to get payment by client reference id", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
	}

	if payment == nil {
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrPaymentNotFound)
	}

	if payment.MerchantID != request.MerchantID {
		s.logger.Warn(ctx, "[ResendPaymentCallback] Merchant ID does not match payment ID", logger.String("MerchantID", payment.MerchantID))
		return pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch)
	}

	// Validate that this payment uses unified payment v2
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	if payment.Metadata != nil {
		metadataB, _ := json.Marshal(payment.Metadata)
		_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)
	}

	if !unifiedPaymentMetadata.IsUnifiedPaymentV2 {
		return pkgErrs.New(response.HttpErrRequest, fmt.Errorf("payment is not using unified payment v2"))
	}

	// Call existing callback function
	s.SendCallback(ctx, payment)

	s.logger.Info(ctx, "[ResendPaymentCallback] Successfully resent callback for payment", logger.Any("request", request))

	return nil
}
