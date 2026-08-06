package internalPaymentController

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	"github.com/paper-indonesia/pdk/go/snap"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *InternalPaymentController) QueryQrMpmStatic(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/payment/QueryQrMpmStatic")
	defer segment.End()

	var (
		err        error
		payment    *paymentModel.PaymentResponse
		merchantID string
	)

	merchantInfoFromCtx := ctx.Value(constant.CtxMerchantInfo)
	merchantCtx, ok := merchantInfoFromCtx.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf("merchant not found")))
		return
	}
	merchantID = merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	var requestPayload paymentModel.QueryQrMpmStaticRequest
	if err = json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(requestPayload); err != nil {
		response.SendOpenApiResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	requestPayload.Validate()

	if payment, err = c.paymentSvc.GetQrMpmStatic(r.Context(), &requestPayload, merchantID); err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, payment)

}

func (h *InternalPaymentController) SNAPQueryQrMpmStatic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	var err error

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payment/GetSNAPQueryQrMpmStatic")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.SnapErrInvalidTokenB2B, errors.New(constant.InvalidB2BTokenSnapErrMsg)))
		return
	}

	request := paymentModel.SNAPQueryQrMpmStaticReq{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		for field, message := range invalidFldQueryQrMpmStatic {
			if strings.Contains(err.Error(), message) {
				err = pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, field))
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
			ctx, "api-v1.0-snap-query-qr-mpm-static", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", merchantAuth.MerchantId),
					fmt.Sprintf("response_code:%s", respBody.ResponseCode),
					fmt.Sprintf("response_message:%s", respBody.ResponseMessage),
				}
			},
		)
	}()

	paymentRequest := &paymentModel.QueryQrMpmStaticRequest{
		ReferenceId:  request.PartnerReferenceNo,
		FromDateTime: request.FromDateTime,
		ToDateTime:   request.ToDateTime,
		PageNumber:   request.PageNumber,
		PageSize:     request.PageSize,
	}

	payment, err := h.paymentSvc.GetQrMpmStatic(ctx, paymentRequest, merchantAuth.MerchantId)
	if err != nil {
		if strings.Contains(err.Error(), response.HttpErrNotFound) {
			errMsg := fmt.Sprintf("%s | %s", response.SnapErrTransactionNotFound, constant.TransactionNotFoundSnapErrMsg)
			response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
		} else {
			response.SendOpenApiSnapResponseError(ctx, w, err)
		}
		return
	}

	if payment.PaymentType == constant.UnifiedPaymentTypeMultiple && h.isSnapMultiplePaymentDelegationEnabled(merchantAuth) && h.unifiedPaymentSvc != nil {

		if unifiedSession, err := h.getUnifiedPaymentSession(ctx, merchantAuth.MerchantId, payment.UUID); err == nil && unifiedSession != nil {

			payment.Status = h.mapUnifiedStatusToSnapStatus(unifiedSession.Status)

			if len(unifiedSession.ChargeDetails) > 0 {
				totalPaidAmount := 0.0
				for _, charge := range unifiedSession.ChargeDetails {
					if charge.Status == "SUCCESS" && charge.CapturedAmount != nil {
						totalPaidAmount += charge.CapturedAmount.Value
					}
				}
				if totalPaidAmount > 0 {
					payment.PaidAmount = &commonModel.Amount{
						Value:    fmt.Sprintf("%.2f", totalPaidAmount),
						Currency: "IDR",
					}
				}
			}
		} else if err != nil {
			h.logger.Warn(ctx, "Failed to get unified payment session for MULTIPLE payment",
				logger.String("paymentUUID", payment.UUID),
				logger.Error(err))
		}
	}

	code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.QueryPaymentStaticQrisMPMSnapApiCode)
	resp := &paymentModel.SnapQueryQrMpmStaticResponse{
		ResponseCode:       code,
		ResponseMessage:    msg,
		ReferenceNo:        payment.UUID,
		PartnerReferenceNo: payment.ReferenceID,
		DetailData:         payment.Qris.DetailData,
		AdditionalInfo: &paymentModel.SnapQrAdditionalInfo{
			QrType:        payment.Qris.QrType,
			QrStatus:      payment.Qris.QrStatus,
			QrExpiredDate: payment.Qris.QrExpiredDate,
			QrUrl:         payment.Qris.QrUrl,
			QrContent:     payment.Qris.QrContent,
			QrImage:       payment.Qris.QrImage,
			MerchantName:  payment.Qris.MerchantName,
		},
	}
	response.SendOpenApiSnapResponseOK(ctx, w, resp)
}

var invalidFldQueryQrMpmStatic = map[string]string{
	"partnerReferenceNo": "partnerReferenceNo of type string",
	"fromDateTime":       "fromDateTime of type string",
	"toDateTime":         "toDateTime of type string",
	"pageSize":           "pageSize of type int",
	"pageNumber":         "pageNumber of type int",
}
