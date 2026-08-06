package adjustment

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// CreateBalanceAdjustmentFromManualTopup	godoc
// @Summary				Balance adjustment from Top up merchant
// @Description			Balance adjustment from Top up merchant
// @ID					crm-balance-adjustment-from-topup-manual
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id	path		string true "ID to adjust balance"
// @Param 				Request	body		adjustment.BalanceAdjustmentRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=adjustment.BalanceAdjustmentResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/balances/topup/manual/{id}/adjustment [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) CreateAdjustmentFromManualTopup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/balance/CreateAdjustmentFromManualTopup")
	defer segment.End()

	adjustmentID := chi.URLParam(r, "id")
	if err := uuid.Validate(adjustmentID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	request := adjustModel.BalanceAdjustmentRequest{
		AdjustmentID: adjustmentID,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	id, err := h.service.CreateBalanceAdjustmentFromManualTopUp(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, adjustModel.BalanceAdjustmentResponse{ID: id})
}
