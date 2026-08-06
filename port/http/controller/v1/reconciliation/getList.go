package reconciliation

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"fmt"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetList		godoc
// @Summary		List recon endpoint
// @Description	List recon endpoint
// @ID			api-recon-list
// @Tags		API - Recon
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=[]reconciliation.Reconciliation}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/recon/list [get]
// @Security	Crm-Service-Key
func (c *ReconciliationController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/reconciliation/GetList")
	defer segment.End()

	var (
		page int64 = 1
		err  error
	)
	// Get query params
	pageStr := r.URL.Query().Get("page")
	status := r.URL.Query().Get("status")

	// Validation and parsing
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}

	var perPage int64 = constant.DefaultPaginationPageSize
	filter := &reconciliation.ReconciliationFilterRequest{
		PerPage: perPage,
		Page:    page,
		Status:  status,
	}

	list, err := c.reconSvc.ListRecon(ctx, filter)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}
	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}
