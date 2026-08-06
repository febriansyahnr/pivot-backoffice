package adjustment

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustmentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// HoldMerchantBalance	godoc
// @Summary				Hold merchant balance
// @Description			Hold merchant balance
// @ID					crm-balance-hold
// @Tags				API - CRM
// @Accept				json
// @Produce				json
// @Param 				Request	body		adjustment.HoldMerchantBalanceRequest true "Request body"
// @Success				200  	{object}	response.Response{data=adjustment.HoldMerchantBalanceResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/balances/hold [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) HoldMerchantBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/balance/HoldMerchantBalance")
	defer segment.End()

	var request adjustmentModel.HoldMerchantBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	var adjustmentResult *adjustmentModel.HoldMerchantBalanceResponse
	var err error

	switch request.Type {
	case string(constant.HoldedBalanceTypeHold):
		adjustmentResult, err = h.service.HoldMerchantBalance(ctx, &request)
	case string(constant.HoldedBalanceTypeRelease):
		adjustmentResult, err = h.service.ReleaseHoldedMerchantBalance(ctx, &request)
	}
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, adjustmentResult)
}
