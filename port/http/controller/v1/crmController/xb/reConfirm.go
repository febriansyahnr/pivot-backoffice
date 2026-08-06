package crmXbController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) ReConfirm(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/xb/ReConfirm")
	defer segment.End()

	payoutId := chi.URLParam(r, "id")
	if err := uuid.Validate(payoutId); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	request := &xbModel.ConfirmPayoutRequest{
		PayoutId: payoutId,
	}

	event, err := h.XbPayoutSvc.ReConfirm(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	if event.NeedAutoConfirm {
		resp, err := h.XbPayoutSvc.Confirm(ctx, &xbModel.ConfirmPayoutRequest{PayoutId: payoutId, MerchantId: event.MerchantID})
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, err)
			return
		}
		h.logger.Info(ctx, "confirm payout", pdkLog.Any("resp", resp))
	}

	response.SendGeneralResponseOK(w, event)
}
