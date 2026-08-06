package purpose

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/purpose"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// List				godoc
// @Summary			List of all master purposes
// @Description		List of all master purposes
// @ID				api-purpose-list
// @Tags			API - Master Purpose
// @Accept			json
// @Produce			json
// @Success			200  	{object}	response.ApiResponse{data=[]purpose.PurposeListResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/purposes/ [get]
// @Security		Bearer
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/purpose/List")
	defer segment.End()

	// response payload
	_ = purpose.PurposeListResponse{}

	response.SendApiResponseOK(w, "Unimplemented")
}
