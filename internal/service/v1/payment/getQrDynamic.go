package paymentService

import (
	"context"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) GetQrMpmDynamic(ctx context.Context, uuid string, referenceId string, merchantId string) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetQrMpmDynamic")
	defer segment.End()

	var (
		payment *paymentModel.Payment
		err     error

		paymentResponse  paymentModel.PaymentResponse
		snapCoreResp     *snapCoreModel.QueryQrMpmDynamicResponseData
		qrisMetadata     *paymentModel.PaymentMetadataQris
		validateMerchant = merchantId
	)

	// Get payment by UUID or reference ID with matching merchant ID
	if uuid != "" {
		payment, err = s.paymentRepo.GetPaymentById(ctx, uuid)
	} else if referenceId != "" {
		payment, err = s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, merchantId, referenceId)
	}

	// Handle if error when get payment data
	if err != nil {
		s.logger.Error(ctx, "error when get payment data by id", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if payment == nil || (uuid != "" && payment.UUID != uuid) || (referenceId != "" && payment.ReferenceID != nil && *payment.ReferenceID != referenceId) {
		err = constant.ErrPaymentNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	// Get prior snapcore response from payment metadata
	metadata := *payment.Metadata
	metadataByte, _ := json.Marshal(metadata)
	json.Unmarshal(metadataByte, &qrisMetadata)

	// Validate merchant here whether the merchant is parent or sub merchant
	if qrisMetadata != nil && qrisMetadata.SubMerchantId != "" {
		validateMerchant = qrisMetadata.SubMerchantId
	}

	// Validate merchantId
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, validateMerchant)
	if err != nil {
		s.logger.Error(ctx, "error when find merchant by id", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if merchant == nil {
		// Note that merchantId is parentId because process happened on behalf of sub merchant
		err = constant.ErrMerchantNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	// Validate payment data
	if !s.validatePaymentData(payment, qrisMetadata, merchant, merchantId, false) {
		err = constant.ErrPaymentNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	// If payment status not PENDING (SUCCESS or VOID), return the payment data
	if payment.Status != constantPayment.PAYMENT_STATUS_PENDING {
		paymentResponse = s.processExistingPaymentQr(payment)
		return &paymentResponse, nil
	}

	// Otherwise, if the payment is still PENDING, query the payment status to snapcore
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    payment.UUID,
		ReferenceId: merchantId,
		From:        serviceName,
	})

	// Request to snap core to query payment status by UUID
	// Get qr uuid from snap core metadata
	metadataSnapCoreByte, _ := json.Marshal(metadata["snapCore"])
	json.Unmarshal(metadataSnapCoreByte, &snapCoreResp)

	if snapCoreResp, err = s.snapCoreRepo.QueryQrMpmDynamic(ctx, snapCoreResp.UUID); err != nil {
		s.logger.Error(ctx, "error when query qr mpm dynamic to snapcore", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}

	// If the payment status is persist PENDING, return the payment data as is
	if snapCoreResp.Status == constantPayment.PAYMENT_STATUS_PENDING {
		paymentResponse = s.processExistingPaymentQr(payment)
		return &paymentResponse, nil
	}

	// Otherwise, update the payment status according snapcore response status
	switch snapCoreResp.Status {
	case constantPayment.PAYMENT_STATUS_SUCCESS:
		payment.Status = constantPayment.PAYMENT_STATUS_SUCCESS
	default: // Payment status from snap core is not SUCCESS nor PENDING
		payment.Status = constantPayment.PAYMENT_STATUS_VOID
	}
	payment.UpdatedAt = time.Now().UTC()
	metadata["snapCoreQuery"] = snapCoreResp
	payment.Metadata = &metadata

	// Update in database
	paymentDTO := payment.ToDTO()
	if err = s.paymentRepo.UpdatePaymentData(ctx, paymentDTO); err != nil {
		s.logger.Error(ctx, "error when updating payment data", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Return the payment data
	paymentResponse = s.processExistingPaymentQr(payment)

	return &paymentResponse, nil
}
