package activityController

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetList		godoc
// @Summary		List activities endpoint
// @Description	List activities endpoint
// @ID			api-activity-list
// @Tags		API - Activity
// @Accept		json
// @Produce		json
// @Param       page    		query     	string  false  "pagination current page"
// @Param       startCreatedAt	query     	string  false  "filter startCreatedAt"
// @Param       endCreatedAt	query     	string  false  "filter endCreatedAt"
// @Success		200  			{object}	response.ApiResponse{data=[]activityModel.Activity,meta=commonModel.Meta}
// @Failure		500  			{object}	response.ApiResponse
// @Router		/api/v1/activities [get]
// @Security	Bearer
func (c *activity) GetList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/activity/GetList")
	defer segment.End()

	var startCreatedAt *time.Time // default nil
	var endCreatedAt *time.Time   // default nil
	var page int64 = 1            // default 1
	var err error

	// Get query params
	startCreatedAtStr := r.URL.Query().Get("startCreatedAt")
	endCreatedAtStr := r.URL.Query().Get("endCreatedAt")
	pageStr := r.URL.Query().Get("page")

	// Validation and parsing
	if startCreatedAtStr != "" {
		parsedStartCreatedAt, err := time.Parse(util.UTCLayout, startCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("Invalid startCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format.")))
			return
		}

		startCreatedAt = &parsedStartCreatedAt
	}
	if endCreatedAtStr != "" {
		parsedEndCreatedAt, err := time.Parse(util.UTCLayout, endCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("Invalid endCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format.")))
			return
		}

		endCreatedAt = &parsedEndCreatedAt
	}
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("Invalid page format. Use number format instead.")))
			return
		}
	}

	// Get Merchant ID from jwt token
	var merchantID *string
	merchantIDFromContext := r.Context().Value(constant.CtxMerchantIDKey)
	if merchantIDStr, ok := merchantIDFromContext.(string); ok {
		merchantID = &merchantIDStr
	}

	filter := activityModel.ActivityFilterRequest{
		MerchantID:     merchantID,
		StartCreatedAt: startCreatedAt,
		EndCreatedAt:   endCreatedAt,
	}
	var perPage int64 = constant.DefaultPaginationPageSize
	if c.config != nil {
		perPage = c.config.AppConfig.PaginationPerPage
	}

	list, err := c.activitySvc.GetList(r.Context(), filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}
