package disbursementController

import (
	"context"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetBulkDisbursementDetail	godoc
// @Summary		Bulk disbursement detail endpoint
// @Description	Get a single bulk disbursement by ID, including aggregate counts (totalAmount, totalTrx, totalApproved, totalRejected, totalSuccess, totalFailed, totalCancelled, totalPending). totalPending is the number of disbursement items in the bulk that are still in PENDING status.
// @ID			api-bulk-disbursement-detail
// @Tags		API - Disbursement
// @Accept		json
// @Produce		json
// @Param 		id				path		string  true  "Bulk Disbursement ID"
// @Success		200  			{object}	response.ApiResponse{data=disbursementModel.BulkDisbursementDetail}
// @Failure		400  			{object}	response.ApiResponse
// @Failure		401  			{object}	response.ApiResponse
// @Failure		404  			{object}	response.ApiResponse
// @Failure		500  			{object}	response.ApiResponse
// @Router		/api/v1/disbursements/bulk/{id} [get]
// @Security	Bearer
func (c *Controller) GetBulkDisbursementDetail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetBulkDisbursementDetail")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	ctx = context.WithValue(ctx, constant.CtxTimeZone, r.Header.Get(constant.HeaderTimeZoneKey))
	detail, err := c.disbursementSvc.GetBulkDisbursementDetail(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check bulk disbursement belongs to the requesting merchant
	if detail.MerchantID != user.MerchantId {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrBulkDisbursementNotFound))
		return
	}

	response.SendApiResponseOK(w, detail)
}
