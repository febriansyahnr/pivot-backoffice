package internalPaymentController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

// FindPaymentById				godoc
// @Summary						Find a payment by id
// @Description					Find a payment by id
// @ID							api-payment-get-by-id
// @Tags						API - Payment
// @Accept						json
// @Produce						json
// @Param						category	query	string	true	"Id of payment"
// @Success						200  	{object}	response.Response{data=paymentModel.PaymentMethodResponse}
// @Failure						500  	{object}	response.Response
// @Router						/api/v1/payments/:id [get]
func (c *InternalPaymentController) FindPaymentById(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/payment/FindPaymentById")
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

	// get id from url path
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, fmt.Errorf("payment id is required")))
		return
	}

	if payment, err = c.paymentSvc.FindPaymentById(r.Context(), id, merchantID); err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, payment)
}

func (h *InternalPaymentController) SNAPGetVA(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	var err error

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payment/SNAPGetVA")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiSnapResponseError(ctx, w, errors.New(response.SnapErrInvalidTokenB2B, fmt.Errorf("%s", constant.InvalidB2BTokenSnapErrMsg)))
		return
	}

	request := paymentModel.SnapVaGetRequest{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		for field, message := range invalidFldGetVA {
			if strings.Contains(err.Error(), message) {
				errMsg := fmt.Sprintf(constant.InvalidFieldFormatSnapFmt, field)
				err = errors.New(response.SnapErrFieldFormat, fmt.Errorf("%s", errMsg))
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
			ctx, "api/snap/v1.0/transfer-va/get-va", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", merchantAuth.MerchantId),
					fmt.Sprintf("response_code:%s", respBody.ResponseCode),
					fmt.Sprintf("response_message:%s", respBody.ResponseMessage),
				}
			},
		)
	}()

	payment, err := h.paymentSvc.FindPaymentById(ctx, request.TrxID, merchantAuth.MerchantId)
	if err != nil {
		if strings.Contains(err.Error(), response.HttpErrNotFound) {
			errMsg := fmt.Sprintf("%s | %s", response.SnapErrInvalidVA, constant.InvalidBillVirtualAccountErrMsg)
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

	code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.GetVirtualAccountSnapApiCode)
	snapResp := payment.ToOpenApiSnapVAPaymentResponse()
	finalResp := &paymentModel.SnapVaResponse{
		ResponseCode:       code,
		ResponseMessage:    msg,
		VirtualAccountData: snapResp,
	}
	response.SendOpenApiSnapResponseOK(ctx, w, finalResp)
}

var invalidFldGetVA = map[string]string{
	"trxId": "trxId of type string",
}
