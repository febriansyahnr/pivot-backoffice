package settlementHold

import (
	"encoding/json"
	"net/http"

	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMSettlementHoldController) CreateSettlementHold(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/settlementHold/CreateSettlementHold")
	defer segment.End()

	var payload settlementHold.CreateUpdateSettlementHoldRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.settlementHoldSvc.CreateUpdate(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)

}
