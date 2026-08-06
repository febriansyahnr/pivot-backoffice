package internalPaymentController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/paper-indonesia/pdk/go/snap"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *InternalPaymentController) QueryQrMpmDynamic(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/payment/QueryQrMpmDynamic")
	defer segment.End()

	var (
		err        error
		payment    *paymentModel.PaymentResponse
		merchantID string
	)

	merchantInfoFromCtx := ctx.Value(constant.CtxMerchantInfo)
	merchantCtx, ok := merchantInfoFromCtx.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, fmt.Errorf("merchant not found")))
		return
	}
	merchantID = merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	// Get uuid or referenceId from URL query
	uuid := r.URL.Query().Get("uuid")
	referenceId := r.URL.Query().Get("referenceId")

	if uuid == "" && referenceId == "" {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, fmt.Errorf("uuid or referenceId is required")))
		return
	}

	if payment, err = c.paymentSvc.GetQrMpmDynamic(r.Context(), uuid, referenceId, merchantID); err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, payment)
}

func (h *InternalPaymentController) SNAPQueryQrMpmDynamic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	var err error

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payment/SnapQueryQrMpmDynamic")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiSnapResponseError(ctx, w, errors.New(response.SnapErrInvalidTokenB2B, fmt.Errorf("%s", constant.InvalidB2BTokenSnapErrMsg)))
		return
	}

	request := paymentModel.SnapQueryQrMpmDynamicRequest{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		for field, message := range invalidFldQueryQrMpmDynamic {
			if strings.Contains(err.Error(), message) {
				err = errors.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, field))
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
			ctx, "api-v1.0-snap-query-qr-mpm-dynamic", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", merchantAuth.MerchantId),
					fmt.Sprintf("response_code:%s", respBody.ResponseCode),
					fmt.Sprintf("response_message:%s", respBody.ResponseMessage),
				}
			},
		)
	}()

	payment, err := h.paymentSvc.GetQrMpmDynamic(ctx, request.OriginalReferenceNo, request.OriginalPartnerReferenceNo, merchantAuth.MerchantId)
	if err != nil {
		if strings.Contains(err.Error(), response.HttpErrNotFound) {
			errMsg := fmt.Sprintf("%s | %s", response.SnapErrTransactionNotFound, constant.TransactionNotFoundSnapErrMsg)
			response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
		} else {
			response.SendOpenApiSnapResponseError(ctx, w, err)
		}
		return
	}

	// Generate response code and message
	code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.QueryPaymentDynamicQrisMPMSnapApiCode)

	// Status mapping
	var latestTransactionStatus string
	var transactionStatusDesc string
	switch payment.Status {
	case "SUCCESS":
		latestTransactionStatus = "00"
		transactionStatusDesc = "SUCCESS"
	case "VOID":
		latestTransactionStatus = "06"
		transactionStatusDesc = "FAILED"
	default:
		latestTransactionStatus = "03"
		transactionStatusDesc = "PENDING"
	}

	resp := paymentModel.SnapQueryQrMpmDynamicResponse{
		ResponseCode:               code,
		ResponseMessage:            msg,
		OriginalReferenceNo:        payment.UUID,
		OriginalPartnerReferenceNo: payment.ReferenceID,
		ServiceCode:                constant.GenerateQrisMPMSnapApiCode,
		LatestTransactionStatus:    latestTransactionStatus,
		TransactionStatusDesc:      transactionStatusDesc,
		Amount: &commonModel.Amount{
			Value:    payment.Qris.Amount.Value.StringFixed(2),
			Currency: payment.Qris.Amount.Currency,
		},
		AdditionalInfo: &paymentModel.SnapQrDynamicAdditionalInfo{
			RRN:             payment.Qris.ReferenceNo,
			QrType:          payment.Qris.QrType,
			QrStatus:        payment.Qris.QrStatus,
			QrExpiredDate:   payment.Qris.QrExpiredDate,
			QrContent:       payment.Qris.QrContent,
			QrUrl:           payment.Qris.QrUrl,
			QrImage:         payment.Qris.QrImage,
			MerchantName:    payment.Qris.MerchantName,
			PaymentStatus:   transactionStatusDesc,
			TransactionDate: payment.Qris.TransactionDate,
		},
	}
	response.SendOpenApiSnapResponseOK(ctx, w, resp)
}

var invalidFldQueryQrMpmDynamic = map[string]string{
	"originalReferenceNo":        "originalReferenceNo of type string",
	"originalPartnerReferenceNo": "originalPartnerReferenceNo of type string",
	"serviceCode":                "serviceCode of type string",
}
