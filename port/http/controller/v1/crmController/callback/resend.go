package crmCallbackController

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callback_model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) ResendCallback(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/callback/ResendCallback")
	defer segment.End()

	// Parse request body
	var req callback_model.ResendCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	// Route to appropriate service based on type
	switch req.Type {
	case constant.TypePayment:
		if err := h.unifiedPaymentSvc.ResendPaymentCallback(ctx, &req); err != nil {
			h.logger.Error(ctx, "[ResendCallback] Failed to resend payment callback", logger.Error(err))
			response.SendGeneralResponseError(w, err)
			return
		}

	case constant.TypeDisbursement:
		if err := h.disbursementSvc.ResendDisbursementCallback(ctx, &req); err != nil {
			h.logger.Error(ctx, "[ResendCallback] Failed to resend disbursement callback", logger.Error(err))
			response.SendGeneralResponseError(w, err)
			return
		}

	default:
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("payment type is invalid")))
		return
	}

	// Send success response
	res := callback_model.ResendCallbackResponse{
		Message:           "Callback resent successfully",
		Type:              req.Type,
		ReferenceID:       req.ReferenceID,
		ClientReferenceID: req.ClientReferenceID,
	}

	response.SendApiResponseOK(w, res)
}
