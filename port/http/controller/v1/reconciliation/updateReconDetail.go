package reconciliation

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	reconModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UpdateReconDetail implements controller.V1ReconciliationController.
func (c *ReconciliationController) UpdateReconDetail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/reconciliation/UpdateReconDetail")
	defer segment.End()

	var (
		err error
	)

	var request reconModel.ReconDetailRequest
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := uuid.Validate(request.ID); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	payload := reconModel.ReconDetail{
		Status: request.Status,
		Reason: request.Reason,
	}

	if err = payload.Validate(); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.reconSvc.UpdateReconDetail(ctx, request.ID, &payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	response.SendGeneralResponseOK(w, "Update Recon Detail Success")
}
