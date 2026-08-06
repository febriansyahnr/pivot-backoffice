package internalPaymentController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/paper-indonesia/pdk/go/snap"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
)

// Update						godoc
// @Summary						Update a payment amount and expiredAt by id
// @Description					Update a payment amount and expiredAt by id
// @ID							api-payment-update-by-id
// @Tags						API - Payment
// @Accept						json
// @Produce						json
// @Param						category	query	string	true	"Id of payment"
// @Success						200  	{object}	response.Response{data=paymentModel.Payment}
// @Failure						500  	{object}	response.Response
// @Router						/api/v1/payments/update/:id [patch]
func (c *InternalPaymentController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/payment/UpdatePayment")
	defer segment.End()

	var (
		err        error
		paymentRes *paymentModel.PaymentResponse
		merchantID string
		parsedTime *time.Time
	)

	// get id from url path
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, fmt.Errorf("payment id is required")))
		return
	}

	var payload paymentModel.PaymentUpdateRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	merchantInfoFromCtx := ctx.Value(constant.CtxMerchantInfo)
	merchantCtx, ok := merchantInfoFromCtx.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		err = fmt.Errorf("merchant not found")
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
		return
	}
	merchantID = merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	merchantInfo, err := c.merchantSvc.FindMerchantByID(ctx, merchantID)
	if err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	if merchantInfo == nil {
		err = fmt.Errorf("merchant not found")
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	// format time
	if payload.ExpiredAt != nil {
		parsedTime = new(time.Time)
		*parsedTime, err = time.Parse(util.UTCLayout, payload.ExpiredAt.Format(util.UTCLayout))
		if err != nil {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
			return
		}
	}

	payload.PaymentId = id
	payload.MerchantId = merchantID
	payload.ExpiredAt = parsedTime
	if paymentRes, err = c.paymentSvc.UpdatePayment(ctx, &payload); err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&merchantCtx.MerchantId,
		nil,
		constant.TagPayment,
		constant.ActivityMerchantUpdatePayment,
		payload,
	)

	response.SendOpenApiResponseOK(w, paymentRes)
}

func (h *InternalPaymentController) SNAPUpdateVA(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	var err error
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payment/SNAPUpdateVA")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiSnapResponseError(ctx, w, errors.New(response.SnapErrInvalidTokenB2B, fmt.Errorf("%s", constant.InvalidB2BTokenSnapErrMsg)))
		return
	}

	request := paymentModel.SnapVaUpdateRequest{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		for field, message := range invalidFldUpdateVa {
			if strings.Contains(err.Error(), message) {
				err = errors.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, field))
				break
			}
		}
		response.SendOpenApiSnapResponseError(ctx, w, err)
		h.logger.Warn(ctx, "Open Api Snap | Unmarshal request", logger.String("information", err.Error()))
		return
	}
	currentPayment, err := h.paymentSvc.FindPaymentById(ctx, request.TrxID, merchantAuth.MerchantId)
	if err != nil {
		response.SendOpenApiSnapResponseError(ctx, w, err)
		h.logger.Warn(ctx, "Open Api Snap | Find payment", logger.String("information", err.Error()))
		return
	}

	request.VirtualAccountTrxType = currentPayment.VirtualAccount.VirtualAccountTrxType
	minAmount, maxAmount, totalAmount, expiredDate, err := request.ValidateAndSetValue()
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
			ctx, "api/snap/v1.0/transfer-va/update-va", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", merchantAuth.MerchantId),
					fmt.Sprintf("response_code:%s", respBody.ResponseCode),
					fmt.Sprintf("response_message:%s", respBody.ResponseMessage),
				}
			},
		)
	}()

	updateRequestPayload := &paymentModel.PaymentUpdateRequest{
		TotalAmount: &totalAmount,
		ExpiredAt:   expiredDate,
		AccountName: request.VirtualAccountName,
		MinAmount:   minAmount,
		MaxAmount:   maxAmount,
		PaymentId:   request.TrxID,
		MerchantId:  merchantAuth.MerchantId,
	}
	payment, err := h.paymentSvc.UpdatePayment(ctx, updateRequestPayload)
	if err != nil {
		if strings.Contains(err.Error(), response.HttpErrNotFound) {
			errMsg := fmt.Sprintf("%s | %s", response.SnapErrTransactionNotFound, constant.TransactionNotFoundSnapErrMsg)
			response.SendOpenApiSnapResponseError(ctx, w, fmt.Errorf("%s", errMsg))
		} else {
			response.SendOpenApiSnapResponseError(ctx, w, err)
		}
		return
	}

	code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, constant.UpdateVirtualAccountSnapApiCode)
	snapResp := payment.ToOpenApiSnapVAPaymentResponse()
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
		constant.ActivityMerchantUpdatePayment,
		updateRequestPayload,
	)
	response.SendOpenApiSnapResponseOK(ctx, w, finalResp)
}

var invalidFldUpdateVa = map[string]string{
	"virtualAccountEmail":   "virtualAccountEmail of type string",
	"virtualAccountPhone":   "virtualAccountPhone of type string",
	"virtualAccountTrxType": "virtualAccountTrxType of type string",
	"virtualAccountName":    "virtualAccountName of type string",
}
