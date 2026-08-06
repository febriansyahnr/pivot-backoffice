package v1CrmPaymentController

import (
	"net/http"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Publish request re publish payment to snap-core
func (h *handler) RetryNotification(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/publish")
	defer segment.End()

	var (
		err     error
		payload paymentModel.CRMRetryNotificationRequest
	)

	paymentID := r.PathValue("id")
	if err = uuid.Validate(paymentID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrPaymentIDNotValid))
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	payload.ID = paymentID
	err = h.paymentSvc.CRMRetryNotification(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendApiResponseOK(w, map[string]interface{}{
		"id":      paymentID,
		"message": "Payment published to CRM successfully",
	})
}

// RetryStaticVANotification godoc
// @Summary				Retry static VA payment notification
// @Description			Retry static VA payment notification for CRM
// @ID					crm-retry-static-va-payment-notification
// @Tags				API - CRM
// @Accept				json
// @Produce				json
// @Param				request	body	paymentModel.CRMStaticVARetryNotificationRequest	true	"retry notification request"
// @Success				200		{object}	response.Response{data=map[string]interface{}}
// @Failure				400		{object}	response.Response
// @Failure				500		{object}	response.Response
// @Router				/crm/v1/payments/static-va/retry-payment-notif [post]
// @Header				all		{string}	X-CRM-Key	"{"key": "value"}"
func (h *handler) RetryStaticVANotification(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/RetryStaticVANotification")
	defer segment.End()

	var (
		err     error
		payload paymentModel.CRMStaticVARetryNotificationRequest
	)

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err = h.validator.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	err = h.paymentSvc.CRMStaticVARetryNotification(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, map[string]interface{}{
		"vaNumber": payload.VANumber,
		"message":  "Static VA payment notification published successfully",
	})
}
