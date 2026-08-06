package disbursementController

import (
	"context"
	errs "errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetList		godoc
// @Summary		Disbursement list endpoint
// @Description	Disbursement list endpoint
// @ID			api-disbursement-list
// @Tags		API - Disbursement
// @Accept		json
// @Produce		json
// @Param       page    				query     	string  false  "pagination current page"
// @Param       bulkId				    query     	string  false  "filter bulkId"
// @Param       status					query	    string  false  "filter status"
// @Param       startCreatedAt			query     	string  false  "filter startCreatedAt"
// @Param       endCreatedAt			query     	string  false  "filter endCreatedAt"
// @Param       keyword					query     	string  false  "search by beneficiary account name or reference ID"
// @Success		200  					{object}	response.ApiResponse{data=[]disbursementModel.Disbursement,meta=commonModel.Meta}
// @Failure		500  					{object}	response.ApiResponse
// @Router		/api/v1/disbursements [get]
// @Security	Bearer
func (c *Controller) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetList")
	defer segment.End()

	var (
		startCreatedAt    *time.Time // default nil
		endCreatedAt      *time.Time // default nil
		bulkID            string
		status            string // approval status
		disbursementType  string
		transactionStatus string
		keyword           string
		sortBy            string
		sort              string

		page    int64 = 1
		perPage int64 = constant.DefaultPaginationPageSize
		err     error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Get query params
	bulkID = r.URL.Query().Get("bulkId")
	status = r.URL.Query().Get("status")
	disbursementType = r.URL.Query().Get("type")
	transactionStatus = r.URL.Query().Get("transactionStatus")
	keyword = r.URL.Query().Get("keyword")
	sortBy = r.URL.Query().Get("sortBy")
	sort = r.URL.Query().Get("sort")

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

	// Don't filter created_at if bulkId exist
	if bulkID != "" && disbursementType == "" {
		startCreatedAt = nil
		endCreatedAt = nil
	}

	filter := &disbursementModel.GetDisbursementFilterRequest{
		MerchantID:             user.MerchantId,
		StartCreatedAt:         startCreatedAt,
		EndCreatedAt:           endCreatedAt,
		BulkID:                 bulkID,
		Status:                 status,
		Type:                   disbursementType,
		TransactionStatus:      transactionStatus,
		Keyword:                keyword,
		SortBy:                 sortBy,
		Sort:                   sort,
	}

	ctx = context.WithValue(ctx, constant.CtxTimeZone, r.Header.Get(constant.HeaderTimeZoneKey))
	list, err := c.disbursementSvc.GetList(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}

func (c *Controller) bindOptionalDateQuery(key string, r *http.Request, d *time.Time) (err error) {
	dateStr := r.URL.Query().Get(key)

	if dateStr == "" {
		return

	} else if d == nil {
		return errs.New("dst date can't be nil")
	}

	if *d, err = time.Parse(time.RFC3339, dateStr); err != nil {
		return errors.New(response.HttpErrRequest, fmt.Errorf("invalid %s format. Use RFC3339 format", key))
	}

	if r.Header.Get(constant.HeaderTimeZoneKey) == "" {
		return errors.New(response.HttpErrRequest, fmt.Errorf("missing %s header", constant.HeaderTimeZoneKey))
	}

	// Parse timezone
	*d, err = util.TimeToUTC(*d, r.Header.Get(constant.HeaderTimeZoneKey))
	if err != nil {
		return errors.New(response.HttpErrRequest, fmt.Errorf("invalid %s header format. Use valid timezone", constant.HeaderTimeZoneKey))
	}

	return
}

func (c *Controller) bindOptionalInt64Query(key string, r *http.Request, dst *int64) (err error) {
	val := r.URL.Query().Get(key)

	if val == "" {
		return

	} else if dst == nil {
		return errs.New("dst value can't be nil")
	}

	if *dst, err = strconv.ParseInt(val, 10, 64); err != nil {
		return errors.New(response.HttpErrRequest, fmt.Errorf("invalid %s format. Use number format instead", key))
	}
	return nil
}
