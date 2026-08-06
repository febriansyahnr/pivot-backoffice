package disbursementController

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// FindByID		godoc
// @Summary		Disbursement by ID endpoint
// @Description	Disbursement by ID endpoint
// @ID			api-disbursement-by-id
// @Tags		API - Disbursement
// @Accept		json
// @Produce		json
// @Param 		id				path		string  true  "Disbursement ID"
// @Param 		subMerchantId	query		string  false "Sub Merchant ID (for platform admin viewing submerchant data)"
// @Success		200  	{object}	response.ApiResponse{data=disbursementModel.DisbursementWithTransactionResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/disbursements/{id} [get]
// @Security	Bearer
func (c *Controller) FindByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/disbursement/FindByID")
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

	// Get data from service
	disbursement, err := c.disbursementSvc.FindByID(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check disbursement with merchant
	if disbursement.MerchantID != merchantID {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrDisbursementNotFound))
		return
	}

	response.SendApiResponseOK(w, disbursement.DisbursementWithTransactionToResponse())
}
