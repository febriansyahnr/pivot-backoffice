package crmPaymentMethodController

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) SetupConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/paymentMethod/SetupConfig")
	defer segment.End()

	var (
		payload *paymentMethodModel.SetupPaymentMethodConfigRequest
		err     error
	)

	merchantID := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	paymentMethodId := chi.URLParam(r, "paymentMethodId")
	if err := uuid.Validate(paymentMethodId); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, errors.New("paymentMethodId is required")))
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	if err = h.validator.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	payload.MerchantID = merchantID
	payload.PaymentMethodID = paymentMethodId
	if err = h.paymentMethodSvc.SetupConfig(ctx, payload); err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{
		"updated": true,
	})
}
