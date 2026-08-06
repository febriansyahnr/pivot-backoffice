package internalPayoutController

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (c *InternalPayoutController) FindByBulkId(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/payout/FindByPayoutId")
	defer segment.End()

	var (
		err        error
		page       int64 = 1
		perPage    int64 = constant.DefaultPaginationPageSize
		merchantID string
	)

	id := chi.URLParam(r, "id")

	// Check if type query param is 'single' before UUID validation
	typeQuery := r.URL.Query().Get("type")
	if typeQuery != "single" {
		// Only validate UUID if not type=single
		if err := uuid.Validate(id); err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("uuid is required")))
			return
		}
	}

	// Merchant info from JWT
	merchantInfoFromCtx := r.Context().Value(constant.CtxMerchantInfo)
	merchantCtx, ok := merchantInfoFromCtx.(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantID = merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	// Process type=single requests
	if typeQuery == "single" {
		// Get payout by ID
		resp, err := c.disbursementSvc.GetDisbursementByReferenceID(ctx, id, merchantID)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, err)
			return
		}

		// Make response array of single payout object
		response.SendOpenApiResponseOK(w, []disbursementModel.PayoutObject{*resp})
		return
	}

	queueKey := fmt.Sprintf(constant.BulkDisbursementInProgressQueueLockFmt, merchantID, id)
	val, _ := c.redis.Get(ctx, queueKey).Result()
	if val != "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrDupCheck, constant.ErrPayoutsInProcess))
		return
	}

	// Get payouts by referenceId
	referenceId := r.URL.Query().Get("referenceId")
	if referenceId != "" {
		resp, err := c.disbursementSvc.GetBulkDisbursementForOpenApiByReferenceID(ctx, id, referenceId, merchantID)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, err)
			return
		}

		response.SendOpenApiResponseOK(w, resp)
		return
	}

	// Get payout by ID
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}
	if perPageStr != "" {
		perPage, err = strconv.ParseInt(perPageStr, 10, 64)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead")))
			return
		}

		if !(perPage > 0 && perPage <= 100) {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(
				response.HttpErrRequest, fmt.Errorf("perPage value must be 1 to 100")))
			return
		}
	}

	filter := &disbursementModel.GetDisbursementFilterRequest{
		MerchantID: merchantID,
		BulkID:     id,
	}
	resp, err := c.disbursementSvc.GetBulkDisbursementForOpenApiByID(ctx, filter, page, perPage)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponsePaginationOK(w, resp.Data, resp.Meta)
}
