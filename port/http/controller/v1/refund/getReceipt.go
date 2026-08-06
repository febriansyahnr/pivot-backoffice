package refund

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetReceipt godoc
// @Summary		Get Refund Receipt
// @Description	Generate and get receipt URL for a refund
// @ID			api-refund-receipt
// @Tags		API - Refund
// @Accept		json
// @Produce		json
// @Param		uuid	path		string	true	"Refund ID (UUID)"
// @Success		200		{object}	response.ApiResponse{data=refundModel.GetRefundReceiptResponse}
// @Failure		400		{object}	response.ApiResponse
// @Failure		401		{object}	response.ApiResponse
// @Failure		404		{object}	response.ApiResponse
// @Failure		500		{object}	response.ApiResponse
// @Router		/api/v1/payments/refunds/{uuid}/receipt [get]
// @Security	Bearer
func (c *RefundController) GetReceipt(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/refund/GetReceipt")
	defer span.End()

	// Get refund ID from URL
	refundID := chi.URLParam(r, "uuid")
	if err := uuid.Validate(refundID); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("uuid is required and must be a valid UUID")))
		return
	}

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Build request
	request := &refundModel.GetRefundReceiptRequest{
		RefundID:   refundID,
		MerchantID: user.MerchantId,
	}

	// Get receipt from service
	result, err := c.refundService.GetReceipt(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
