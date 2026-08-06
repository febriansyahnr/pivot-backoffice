package disbursementController

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetReceiptByID 	godoc
// @Summary		Disbursement Receipt by ID endpoint
// @Description	Disbursement Receipt by ID endpoint
// @ID			api-disbursement-receipt-by-id
// @Tags		API - Disbursement
// @Accept		json
// @Produce		json
// @Param 		id				path		string  true  "Disbursement ID"
// @Param 		subMerchantId	query		string  false "Sub Merchant ID (for platform admin viewing submerchant data)"
// @Success		200  	{object}	response.ApiResponse{data=disbursementModel.Disbursement}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/disbursements/{id}/receipt [get]
// @Security	Bearer
func (c *Controller) GetReceiptByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/disbursement/GetReceiptByID")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	//Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchantID := user.MerchantId
	subMerchantId := r.URL.Query().Get("subMerchantId")
	if subMerchantId != "" {
		err := c.merchant.ValidateSubMerchantParent(ctx, user.MerchantId, subMerchantId)
		if err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}
		merchantID = subMerchantId
	}

	// Get disbursement receipt
	resp, err := c.disbursementSvc.GetReceiptByID(ctx, &disbursementModel.GetDisbursementReceiptRequest{
		DisbursementID: id,
		MerchantID:     merchantID,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
