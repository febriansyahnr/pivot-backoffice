package disbursementController

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ApprovalActions	godoc
// @Summary			Approval actions disbursement
// @Description		Approval actions disbursement
// @ID				api-approval-actions-disbursements
// @Tags			API - Disbursement
// @Accept			json
// @Produce			json
// @Param			Request	body		disbursementModel.ApprovalActionsRequest true "JSON Body for Approve/Reject disbursement"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/disbursements/approval-actions [post]
// @Security		Bearer
func (c *Controller) ApprovalActions(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/ApprovalActions")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var requestPayload disbursementModel.ApprovalActionsRequest
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	requestPayload.UserID = user.UUID
	requestPayload.MerchantID = user.MerchantId

	payoutCutOffTime, err := c.disbursementSvc.GetCutOffTimeStatus(ctx, time.Now().UTC(), requestPayload.MerchantID, nil)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	cleanedRequestPayload, err := c.disbursementSvc.ValidateBatchPayoutItems(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	resp, err := c.disbursementSvc.ApprovalAction(ctx, cleanedRequestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return

	} else if payoutCutOffTime.Status == constant.DisbursementCutOffTimeStatusOngoing {
		response.SendApiResponseSuccess(
			w, http.StatusAccepted, c.config.DisbursementConfig.CutOffTimeWindow.TransactionInfo, resp,
		)
		return
	}
	response.SendApiResponseOK(w, resp)
}
