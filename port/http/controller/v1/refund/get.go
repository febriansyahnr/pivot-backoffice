package refund

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetByID godoc
// @Summary		Refund Detail Endpoint
// @Description	This endpoint used to get refund detail for merchant dashboard
// @ID			api-refund-detail-dashboard
// @Tags		API - Refund
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/refunds/{uuid} [get]
// @Security	Bearer
func (c *RefundController) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/refund/GetByID")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var request refundModel.FilterRefundRequest
	request.MerchantID = user.MerchantId
	request.UUID = chi.URLParam(r, "uuid")

	if request.UUID == "" {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("id is required")))
		return
	}

	refundDetail, err := c.refundService.GetRefundDetailWithStatusHistories(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, refundDetail)
}
