package simulationController

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *Handler) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/simulation/GetPaymentByID")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	payment, err := h.paymentSvc.FindPaymentForSimulationByID(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, payment)
}

func (h *Handler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/simulation/ProcessPayment")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	var payload paymentModel.ProcessPaymentSimulation
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	chargeStatus := constant.ChargeStatusSuccess
	if payload.PhoneNumber != "" {
		chargeStatus = constant.GetChargeStatusByPhoneNumber(payload.PhoneNumber)
		if chargeStatus == constant.PhoneNumberUnknown {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, errors.New("invalid phone number")))
			return
		}
	}

	if err := h.paymentSvc.ProcessPaymentForSimulationByID(ctx, id, payload.PaidAmount, chargeStatus); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	redirectURL, err := h.GetRedirectionURL(ctx, id, chargeStatus)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, map[string]any{
		"id":          id,
		"redirectUrl": redirectURL,
	})
}
