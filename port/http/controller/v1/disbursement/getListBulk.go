package disbursementController

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetListBulkDisbursement		godoc
// @Summary		Bulk disbursement list endpoint
// @Description	Bulk disbursement list endpoint
// @ID			api-bulk-disbursement-list
// @Tags		API - Disbursement
// @Accept		json
// @Produce		json
// @Param       page    				query     	string  false  "pagination current page"
// @Param       status					query	    string  false  "filter status"
// @Param       startCreatedAt			query     	string  false  "filter startCreatedAt"
// @Param       endCreatedAt			query     	string  false  "filter endCreatedAt"
// @Param       referenceId			    query     	string  false  "search by batch UUID (reference ID)"
// @Success		200  					{object}	response.ApiResponse{data=[]disbursementModel.BulkDisbursementWithAggregate,meta=commonModel.Meta}
// @Failure		500  					{object}	response.ApiResponse
// @Router		/api/v1/disbursements/bulk/list [get]
// @Security	Bearer
func (c *Controller) GetListBulkDisbursement(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetListBulkDisbursement")
	defer segment.End()

	var (
		startCreatedAt *time.Time // default nil
		endCreatedAt   *time.Time // default nil
		status         string
		sortBy         string
		sort           string
		referenceID    string

		page    int64 = 1
		perPage int64 = constant.DefaultPaginationPageSize
		err     error
	)

	// Get User Info from jwt token
	userInfoFromCtx := ctx.Value(constant.CtxUserInfoKey)
	user, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err = fmt.Errorf("user not found")
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	// Get query params
	status = r.URL.Query().Get("status")
	sortBy = r.URL.Query().Get("sortBy")
	sort = r.URL.Query().Get("sort")
	referenceID = r.URL.Query().Get("referenceId")

	// Validation and parsing
	startCreated := time.Time{}
	if err = c.bindOptionalDateQuery("startCreatedAt", r, &startCreated); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	if !startCreated.IsZero() {
		startCreatedAt = &startCreated
	}

	endCreated := time.Time{}
	if err = c.bindOptionalDateQuery("endCreatedAt", r, &endCreated); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	if !endCreated.IsZero() {
		endCreatedAt = &endCreated
	}
	if err = c.bindOptionalInt64Query("page", r, &page); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if err = c.bindOptionalInt64Query("perPage", r, &perPage); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	filter := &disbursementModel.GetBulkDisbursementFilterRequest{
		MerchantID:     user.MerchantId,
		StartCreatedAt: startCreatedAt,
		EndCreatedAt:   endCreatedAt,
		Status:         status,
		Sort:           sort,
		SortBy:         sortBy,
		ReferenceID:    referenceID,
	}

	ctx = context.WithValue(ctx, constant.CtxTimeZone, r.Header.Get(constant.HeaderTimeZoneKey))
	list, err := c.disbursementSvc.GetListBulk(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}
