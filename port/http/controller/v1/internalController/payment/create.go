package internalPaymentController

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/paper-indonesia/pdk/go/snap"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/shopspring/decimal"
)

func (c *InternalPaymentController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/payment/Create")
	defer segment.End()

	var (
		err error
		now = time.Now()
	)

	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, fmt.Errorf("merchant not found")))
		return
	}

	merchantID := merchantCtx.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantCtx.MerchantId)
	}

	var requestPayload paymentModel.PaymentRequest
	if err = json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(requestPayload); requestPayload.PaymentMethod == constantPayment.PAYMENT_METHOD_VIRTUAL_ACCOUNT && err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if requestPayload.PaymentItems != nil && len(*requestPayload.PaymentItems) > 0 {
		for i, item := range *requestPayload.PaymentItems {
			if item.Qty == 0 {
				(*requestPayload.PaymentItems)[i].Qty = 1
			}

			if err = c.validate.Struct(item); err != nil {
				response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
				return
			}
		}
	}

	switch requestPayload.PaymentMethod {
	case constantPayment.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		err = c.validateVirtualAccountPayload(requestPayload)
	case constantPayment.PAYMENT_METHOD_QRIS:
		err = c.validateQrisPayload(requestPayload)
	default:
		err = pkgErrors.New(response.HttpErrRequest, errors.New("payment method is not allowed"))
	}
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	ww := monitor.WrapResponse(w, r)
	defer monitor.WriteAndSend(
		ctx, "internal-controller-payment-create", now, ww, err, func() []string {
			return []string{
				fmt.Sprintf("merchant_id:%s", merchantID),
				fmt.Sprintf("amount:%s", requestPayload.TotalAmount.Value.String()),
				fmt.Sprintf("payment_method:%s", requestPayload.PaymentMethod),
			}
		},
	)

	// Call Payment Service
	resp, err := c.paymentSvc.CreatePayment(ctx, merchantID, requestPayload)
	if err != nil {
		response.SendOpenApiResponseError(ww, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&merchantCtx.MerchantId,
		nil,
		constant.TagPayment,
		constant.ActivityMerchantCreatePayment,
		requestPayload,
	)
	response.SendOpenApiResponseOK(ww, resp)
}

func (h *InternalPaymentController) SNAPGenerateQRMpm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	var err error

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payment/SNAPGenerateQRMpm")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiSnapResponseError(ctx, w, pkgErrors.New(response.SnapErrInvalidTokenB2B, errors.New(constant.InvalidB2BTokenSnapErrMsg)))
		return
	}

	request := paymentModel.SnapGenerateQrMpmRequest{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		for field, message := range invalidFldQueryGenerateQRMpm {
			if strings.Contains(err.Error(), message) {
				errMsg := fmt.Sprintf(constant.InvalidFieldFormatSnapFmt, field)
				err = pkgErrors.New(response.SnapErrFieldFormat, fmt.Errorf("%s", errMsg))
				break
			}
		}
		response.SendOpenApiSnapResponseError(ctx, w, err)
		h.logger.Warn(ctx, "Open Api Snap | Unmarshal request", logger.String("information", err.Error()))
		return
	}
	if err = request.ValidateAndSetValue(); err != nil {
		response.SendOpenApiSnapResponseError(ctx, w, err)
		h.logger.Warn(ctx, "Open Api Snap | Validate request body", logger.String("information", err.Error()))
		return
	}

	ww := monitor.WrapResponse(w, r)
	defer func() {
		var respBody response.OpenApiSnapResp

		wr, ok := w.(*middleware.ResponseWriter)
		if ok {
			_ = json.Unmarshal(wr.BodyBytes(), &respBody)
		}
		monitor.WriteAndSend(
			ctx, "api/snap/v1.0/qr/qr-mpm-generate", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", merchantAuth.MerchantId),
					fmt.Sprintf("response_code:%s", respBody.ResponseCode),
					fmt.Sprintf("response_message:%s", respBody.ResponseMessage),
				}
			},
		)
	}()

	validityPeriod, _ := strconv.Atoi(request.ValidityPeriod)

	// mapping to qr mpm generate request
	paymentRequestPayload := &paymentModel.PaymentRequest{
		ReferenceID:   request.PartnerReferenceNo,
		PaymentMethod: "QRIS",
		Qris: &paymentModel.PaymentMetadataQris{
			QrType:         request.AdditionalInfo.QrType,
			QrMethodType:   "MPM",
			SubMerchantId:  request.SubMerchantId,
			Amount:         request.Amount,
			ValidityPeriod: validityPeriod,
		},
		IsSnap:    true,
		CreatedBy: merchantAuth.MerchantId,
	}
	err = h.validateQrisPayload(*paymentRequestPayload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	// Check feature flag for MULTIPLE payment type delegation for static QR
	if request.AdditionalInfo.QrType == constant.QrTypeStatic && h.isSnapMultiplePaymentDelegationEnabled(merchantAuth) {
		// Set payment type to MULTIPLE for delegation to Unified V2
		paymentRequestPayload.PaymentType = constant.UnifiedPaymentTypeMultiple

		resp, err := h.createPaymentViaUnifiedV2(ctx, merchantAuth.MerchantId, *paymentRequestPayload)
		if err != nil {
			h.logger.Error(ctx, "Failed to create payment via Unified V2", logger.Error(err))
			h.handleSnapQRError(ctx, w, err)
			return
		}

		code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.GenerateQrisMPMSnapApiCode)
		finalResp := &paymentModel.SnapGenerateQrMpmResponse{
			ResponseCode:       code,
			ResponseMessage:    msg,
			ReferenceNo:        resp.UUID,
			PartnerReferenceNo: resp.ReferenceID,
		}

		if resp.Qris != nil {
			finalResp.MerchantName = resp.Qris.MerchantName
			finalResp.QrContent = resp.Qris.QrContent
			finalResp.QrUrl = resp.Qris.QrUrl
			finalResp.QrImage = resp.Qris.QrImage
			finalResp.AdditionalInfo = &paymentModel.SnapQrAdditionalInfo{
				QrType:        constant.QrTypeStatic,
				QrStatus:      constant.QrStatusActive,
				PaymentStatus: resp.Status,
				Metadata:      resp.Qris.Metadata,
			}
		}

		// publish activity, do nothing on error
		_ = h.rabbitMqExt.PublishActivity(
			ctx,
			&merchantAuth.MerchantId,
			nil,
			constant.TagPayment,
			constant.ActivityMerchantCreatePayment,
			paymentRequestPayload,
		)
		response.SendOpenApiSnapResponseOK(ctx, w, finalResp)
		return
	}

	payment, err := h.paymentSvc.CreatePayment(ctx, merchantAuth.MerchantId, *paymentRequestPayload)
	if err != nil {
		h.handleSnapQRError(ctx, w, err)
		return
	}

	code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.GenerateQrisMPMSnapApiCode)
	resp := &paymentModel.SnapGenerateQrMpmResponse{

		ResponseCode:    code,
		ResponseMessage: msg,

		ReferenceNo:        payment.UUID,
		PartnerReferenceNo: payment.Qris.PartnerReferenceNo,
		MerchantName:       payment.Qris.MerchantName,
		QrContent:          payment.Qris.QrContent,
		QrUrl:              payment.Qris.QrUrl,
		QrImage:            payment.Qris.QrImage,
		AdditionalInfo: &paymentModel.SnapQrAdditionalInfo{
			QrType:        payment.Qris.QrType,
			QrStatus:      payment.Qris.QrStatus,
			QrExpiredDate: payment.Qris.QrExpiredDate,
			PaymentStatus: payment.Qris.PaymentStatus,
			Metadata:      payment.Qris.Metadata,
		},
	}
	_ = h.rabbitMqExt.PublishActivity(
		ctx,
		&merchantAuth.MerchantId,
		nil,
		constant.TagPayment,
		constant.ActivityMerchantCreatePayment,
		paymentRequestPayload,
	)
	response.SendOpenApiSnapResponseOK(ctx, w, resp)
}

var invalidFldQueryGenerateQRMpm = map[string]string{
	"partnerReferenceNo": "partnerReferenceNo of type string",
	"subMerchantId":      "subMerchantId of type string",
	"validityPeriod":     "validityPeriod of type string",
}

func (h *InternalPaymentController) SNAPCreateVA(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	var err error

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payment/SNAPCreateVA")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiSnapResponseError(ctx, w, pkgErrors.New(response.SnapErrInvalidTokenB2B, errors.New(constant.InvalidB2BTokenSnapErrMsg)))
		return
	}

	request := paymentModel.SnapVACreateData{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		for field, message := range invalidFldCreateVa {
			if strings.Contains(err.Error(), message) {
				errMsg := fmt.Sprintf(constant.InvalidFieldFormatSnapFmt, field)
				err = pkgErrors.New(response.SnapErrFieldFormat, fmt.Errorf("%s", errMsg))
				break
			}
		}
		response.SendOpenApiSnapResponseError(ctx, w, err)
		h.logger.Warn(ctx, "Open Api Snap | Unmarshal request", logger.String("information", err.Error()))
		return
	}
	minAmount, maxAmount, totalAmount, paymentItems, expiredDate, err := request.ValidateAndSetValue()
	if err != nil {
		response.SendOpenApiSnapResponseError(ctx, w, err)
		h.logger.Warn(ctx, "Open Api Snap | Validate request body", logger.String("information", err.Error()))
		return
	}

	ww := monitor.WrapResponse(w, r)
	defer func() {
		var respBody response.OpenApiSnapResp

		wr, ok := w.(*middleware.ResponseWriter)
		if ok {
			_ = json.Unmarshal(wr.BodyBytes(), &respBody)
		}
		monitor.WriteAndSend(
			ctx, "api/snap/v1.0/transfer-va/create-va", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", merchantAuth.MerchantId),
					fmt.Sprintf("response_code:%s", respBody.ResponseCode),
					fmt.Sprintf("response_message:%s", respBody.ResponseMessage),
				}
			},
		)
	}()

	paymentRequestPayload := &paymentModel.PaymentRequest{
		ReferenceID:   request.AdditionalInfo.ReferenceID,
		PaymentMethod: "VIRTUAL_ACCOUNT",
		TotalAmount:   totalAmount,
		VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
			Issuer:                request.AdditionalInfo.Issuer,
			VirtualAccountTrxType: request.VirtualAccountTrxType,
			VirtualAccountName:    request.VirtualAccountName,
			VirtualAccountNumber:  request.VirtualAccountNo,
			MinAmount:             minAmount,
			MaxAmount:             maxAmount,
			BillDetails:           request.BillDetails,
			ExpiredDate:           expiredDate,
		},
		Customer:     *request.AdditionalInfo.Customer,
		PaymentItems: paymentItems,
		IsSnap:       true,
		CreatedBy:    merchantAuth.MerchantId,
	}

	err = h.validateVirtualAccountPayload(*paymentRequestPayload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	// Check feature flag for MULTIPLE payment type delegation for static VAs
	if (request.VirtualAccountTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC || request.VirtualAccountTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC) && h.isSnapMultiplePaymentDelegationEnabled(merchantAuth) {
		// Set payment type to MULTIPLE for delegation to Unified V2
		paymentRequestPayload.PaymentType = constant.UnifiedPaymentTypeMultiple

		resp, err := h.createPaymentViaUnifiedV2(ctx, merchantAuth.MerchantId, *paymentRequestPayload)
		if err != nil {
			h.logger.Error(ctx, "Failed to create payment via Unified V2", logger.Error(err))
			h.handleSnapVAError(ctx, w, err)
			return
		}

		code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.CreateVirtualAccountSnapApiCode)
		snapResp := resp.ToOpenApiSnapVAPaymentResponse()
		snapResp.BillDetails = request.BillDetails // during VA creation no mapping of bill detail in the service layer
		finalResp := &paymentModel.SnapVaResponse{
			ResponseCode:       code,
			ResponseMessage:    msg,
			VirtualAccountData: snapResp,
		}

		// publish activity, do nothing on error
		_ = h.rabbitMqExt.PublishActivity(
			ctx,
			&merchantAuth.MerchantId,
			nil,
			constant.TagPayment,
			constant.ActivityMerchantCreatePayment,
			paymentRequestPayload,
		)
		response.SendOpenApiSnapResponseOK(ctx, w, finalResp)
		return
	}

	payment, err := h.paymentSvc.CreatePayment(ctx, merchantAuth.MerchantId, *paymentRequestPayload)
	if err != nil {
		h.handleSnapVAError(ctx, w, err)
		return
	}

	code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.CreateVirtualAccountSnapApiCode)
	snapResp := payment.ToOpenApiSnapVAPaymentResponse()
	snapResp.BillDetails = request.BillDetails // during VA creation no mapping of bill detail in the service layer
	finalResp := &paymentModel.SnapVaResponse{
		ResponseCode:       code,
		ResponseMessage:    msg,
		VirtualAccountData: snapResp,
	}
	_ = h.rabbitMqExt.PublishActivity(
		ctx,
		&merchantAuth.MerchantId,
		nil,
		constant.TagPayment,
		constant.ActivityMerchantCreatePayment,
		paymentRequestPayload,
	)
	response.SendOpenApiSnapResponseOK(ctx, w, finalResp)
}

var invalidFldCreateVa = map[string]string{
	"virtualAccountEmail":   "virtualAccountEmail of type string",
	"virtualAccountPhone":   "virtualAccountPhone of type string",
	"virtualAccountTrxType": "virtualAccountTrxType of type string",
	"virtualAccountName":    "virtualAccountName of type string",
}

func (c *InternalPaymentController) validateVirtualAccountPayload(payload paymentModel.PaymentRequest) error {
	virtualAccount := payload.VirtualAccount

	if virtualAccount == nil {
		return pkgErrors.New(response.HttpErrRequest, errors.New("virtualAccount object is required"))
	}

	if err := c.validate.Struct(virtualAccount); err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	if virtualAccount.ExpiredDate != nil && virtualAccount.ExpiredDate.Before(time.Now()) {
		return pkgErrors.New(response.HttpErrRequest, errors.New("expiredDate is not allowed to be less than current time"))
	}

	switch virtualAccount.VirtualAccountTrxType {
	case constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC:
		if virtualAccount.VirtualAccountNumber != "" {
			return pkgErrors.New(response.HttpErrRequest, errors.New("virtualAccountNumber is not allowed for type CLOSED_DYNAMIC payment"))
		}

		if payload.PaymentItems == nil {
			return pkgErrors.New(response.HttpErrRequest, errors.New("paymentItems is required for type CLOSED_DYNAMIC payment"))
		}

		if payload.TotalAmount.Value.Cmp(decimal.NewFromInt(constantPayment.VIRTUAL_ACCOUNT_MINIMUM_AMOUNT)) < 0 {
			return pkgErrors.New(response.HttpErrRequest, errors.New("totalAmount is not allowed to be less than 10000 for type CLOSED_DYNAMIC payment"))
		}
	case constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC:
		if virtualAccount.VirtualAccountNumber == "" {
			return pkgErrors.New(response.HttpErrRequest, errors.New("virtualAccountNumber is required for type OPEN_STATIC payment"))
		}

	case constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC:
		if virtualAccount.VirtualAccountNumber == "" {
			return pkgErrors.New(response.HttpErrRequest, errors.New("virtualAccountNumber is required for type CLOSED_STATIC payment"))
		}
	default:
		return pkgErrors.New(response.HttpErrRequest, errors.New("virtualAccountTrxType is not allowed"))
	}

	return nil
}

func (c *InternalPaymentController) validateQrisPayload(payload paymentModel.PaymentRequest) error {
	qris := payload.Qris

	if qris == nil {
		return pkgErrors.New(response.HttpErrRequest, errors.New("qris object is required"))
	}

	if err := c.validate.Struct(qris); err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	switch strings.ToUpper(qris.QrType) {
	case constant.QrTypeStatic:
		if qris.Amount != nil {
			return pkgErrors.New(response.HttpErrRequest, errors.New("amount is not allowed for type STATIC payment"))
		}

		qris.Amount = &paymentModel.Amount{
			Value:    decimal.NewFromInt(0),
			Currency: "IDR",
		}

		if qris.ValidityPeriod != 0 {
			return pkgErrors.New(response.HttpErrRequest, errors.New("validityPeriod is not allowed for type STATIC payment"))
		}
		// Todo: Check if qris static already exist, can only generate once
	case constant.QrTypeDynamic:
		if qris.Amount == nil || qris.Amount.Value.Cmp(decimal.NewFromInt(0)) < 0 {
			return pkgErrors.New(response.HttpErrRequest, errors.New("amount is required for type DYNAMIC payment"))
		}
		if qris.ValidityPeriod == 0 {
			qris.ValidityPeriod = 3600
		}
		if qris.ValidityPeriod > constant.QrisDynamicValidityPeriodMax {
			return pkgErrors.New(response.HttpErrRequest, errors.New("validityPeriod is not allowed to be more than 10800 for type DYNAMIC payment"))
		}
	default:
		return pkgErrors.New(response.HttpErrRequest, errors.New("qrType is not allowed"))
	}

	return nil
}

func (h *InternalPaymentController) handleSnapVAError(ctx context.Context, w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), response.HttpErrNotFound) {
		errMsg := fmt.Sprintf("%s | %s", response.SnapErrTransactionNotFound, constant.TransactionNotFoundSnapErrMsg)
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else if strings.Contains(err.Error(), response.HttpErrDupCheck) {
		errMsg := fmt.Sprintf("%s | %s", response.SnapErrDuplicatePartnerRef, constant.DuplicatePartnerReferenceNoErrMsg)
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else if strings.Contains(err.Error(), "payment method") {
		errMsg := fmt.Sprintf("%s | %s", response.SnapErrInvalidPartner, constant.PartnerNotFoundErrMsg)
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else if strings.Contains(err.Error(), constant.ErrInvalidAmount.Error()) {
		errMsg := fmt.Sprintf("%s | %s", response.SnapErrInvalidAmount, constant.InvalidAmountSnapErrMsg)
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else if strings.Contains(err.Error(), constant.ErrExceedMaxExpiryDateCheck) {
		errRaw := err.Error()
		errSplits := strings.Split(errRaw, "|")
		errMsg := ""
		if len(errSplits) > 1 {
			errMsg = fmt.Sprintf("%s | %s", response.SnapErrFieldFormat, strings.TrimSpace(errSplits[1]))
		} else {
			errMsg = fmt.Sprintf("%s | %s", response.SnapErrFieldFormat, fmt.Sprintf(constant.InvalidFieldFormatSnapFmt, "expiredDate"))
		}
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else {
		response.SendOpenApiSnapResponseError(ctx, w, err)
	}
}

func (h *InternalPaymentController) handleSnapQRError(ctx context.Context, w http.ResponseWriter, err error) {
	if strings.Contains(strings.ToLower(err.Error()), constant.ErrMerchantNotFound.Error()) {
		errMsg := fmt.Sprintf("%s | %s", response.SnapErrInvalidMerchant, constant.InvalidMerchant)
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else if strings.Contains(err.Error(), response.HttpErrRequest) && strings.Contains(strings.ToLower(err.Error()), strings.ToLower(constant.DuplicatePartnerReferenceNoErrMsg)) {
		errMsg := fmt.Sprintf("%s | %s", response.SnapErrDuplicatePartnerRef, constant.DuplicatePartnerReferenceNoErrMsg)
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else if strings.Contains(err.Error(), response.HttpErrNotFound) {
		errMsg := fmt.Sprintf("%s | %s", response.SnapErrTransactionNotFound, constant.TransactionNotFoundSnapErrMsg)
		response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
	} else {
		response.SendOpenApiSnapResponseError(ctx, w, err)
	}
}
