package installmentplan

import (
	"encoding/json"
	"net/http"

	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1CRMInstallmentPlanController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crm/installmentPlan/Create")
	defer segment.End()

	var payload installmentPlanModel.CreateInstallmentPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	plan, err := c.installmentPlanSvc.Create(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, plan.ToResponseModel())
}
