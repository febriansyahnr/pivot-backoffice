package merchantRcn

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"

	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (m *MerchantRcnController) FindByIDAndMerchantID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/internalController/v1/merchantRCN/FindByIDAndMerchantID")
	defer segment.End()

	id := chi.URLParam(r, "rcn_id")
	if id == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("rcn id is required")))
		return
	}

	merchantId := r.Header.Get(constant.HeaderXMerchantId)
	if merchantId == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("merchant id is required")))
		return
	}

	merchant, err := m.merchantSvc.FindByIDAndMerchantID(r.Context(), id, merchantId)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	if merchant == nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrNotFound, err))
		return
	}

	response.SendApiResponseOK(w, merchant)
}
