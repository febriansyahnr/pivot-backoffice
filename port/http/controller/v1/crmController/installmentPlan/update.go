package installmentplan

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1CRMInstallmentPlanController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crm/installmentPlan/Update")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	var payload installmentPlanModel.UpdateInstallmentPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	payload.UUID = id

	plan, err := c.installmentPlanSvc.Update(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, plan.ToResponseModel())
}
