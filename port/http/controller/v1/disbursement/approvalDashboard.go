package disbursementController

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetDisbursementApprovalDashboard		godoc
// @Summary		Disbursement approval dashboard endpoint
// @Description	Disbursement approval dashboard endpoint
// @ID			api-disbursement-approval-dashboard
// @Tags		API - Disbursement
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=disbursementDashboardModel.DisbursementApprovalDashboardResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/disbursements/approval-dashboard [get]
// @Security	Bearer
func (c *Controller) GetDisbursementApprovalDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/disbursement/GetDisbursementApprovalDashboard")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	userInfoFromCtx := r.Context().Value(constant.CtxUserInfoKey)
	user, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err = constant.ErrUserNotFound
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrUnauthorized, err))
		return
	}

	// Convert merchantID to uuid
	merchantID, err := uuid.Parse(user.MerchantId)
	if err != nil {
		err = constant.ErrMerchantIDNotValid
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrUnauthorized, err))
		return
	}

	dashboard, err := c.disbursementDashboardSvc.GetApprovalDashboard(r.Context(), merchantID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, dashboard)
}
