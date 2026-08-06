package adjustment

import (
	"encoding/json"
	"net/http"

	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) CreateBalanceAdjustment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/balance/CreateBalanceAdjustment")
	defer segment.End()

	var request adjustModel.MerchantBalanceAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	data, err := h.service.CreateMerchantBalanceAdjustment(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, data.ToMerchantBalanceAdjustmentResponse())
}
