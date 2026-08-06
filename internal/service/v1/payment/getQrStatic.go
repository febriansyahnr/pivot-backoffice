package paymentService

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) GetQrMpmStatic(ctx context.Context, request *paymentModel.QueryQrMpmStaticRequest, merchantId string) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetQrMpmStatic")
	defer segment.End()

	var (
		payment           *paymentModel.Payment
		err               error
		qrisMetadata      *paymentModel.PaymentMetadataQris
		paymentResponse   paymentModel.PaymentResponse
		snapCoreResp      *snapCoreModel.GenerateQrMpmResponseData
		snapCoreQueryResp *snapCoreModel.QueryQrMpmStaticResponseData
		validateMerchant  = merchantId
	)

	// Get payment by UUID
	payment, err = s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, merchantId, request.ReferenceId)

	// Handle if error when get payment data
	if err != nil {
		s.logger.Error(ctx, "error when get payment data by id", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if payment == nil || payment.ReferenceID != nil && *payment.ReferenceID != request.ReferenceId || payment.Metadata == nil {
		err = constant.ErrPaymentNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	// Validate payment data
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
	if !s.validatePaymentData(payment, qrisMetadata, merchant, merchantId, true) {
		err = constant.ErrPaymentNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	// query the payment status to snapcore
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    payment.UUID,
		ReferenceId: merchantId,
		From:        serviceName,
	})

	// Request to snap core to query qr mpm static using request data
	// Get qr uuid from snap core metadata
	metadataSnapCoreByte, _ := json.Marshal(metadata["snapCore"])
	json.Unmarshal(metadataSnapCoreByte, &snapCoreResp)

	if snapCoreQueryResp, err = s.snapCoreRepo.QueryQrMpmStatic(ctx, snapCoreModel.QueryQrMpmStaticRequest{
		PartnerReferenceNo: snapCoreResp.PartnerReferenceNo,
		FromDateTime:       request.FromDateTime,
		ToDateTime:         request.ToDateTime,
		PageNumber:         request.PageNumber,
		PageSize:           request.PageSize,
	}); err != nil {
		s.logger.Error(ctx, "error when query qr mpm static to snapcore", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}

	// merchant descriptor using name from merchant short name
	snapCoreResp.MerchantName = merchant.ShortName
	if snapCoreResp.MerchantName == "" {
		snapCoreResp.MerchantName = merchant.Name
	}

	paymentResponse.ToQueryQrMpmStaticResponse(
		payment,
		snapCoreResp,
		snapCoreQueryResp,
	)

	return &paymentResponse, nil
}

func (s *PaymentService) validatePaymentData(payment *paymentModel.Payment, qrisMetadata *paymentModel.PaymentMetadataQris, merchant *merchantModel.Merchant, merchantId string, isStatic bool) bool {
	if payment.PaymentMethod.Type != constantPayment.PAYMENT_METHOD_QRIS {
		return false
	}

	expectedQrType := constant.QrTypeStatic
	if !isStatic {
		expectedQrType = constant.QrTypeDynamic
	}

	if qrisMetadata != nil && qrisMetadata.QrType != expectedQrType {
		return false
	}

	if qrisMetadata != nil && qrisMetadata.SubMerchantId != "" && merchant.ParentID.String != merchantId {
		return false
	}

	if qrisMetadata != nil && qrisMetadata.SubMerchantId == "" && merchant.UUID != merchantId {
		return false
	}

	return true
}
