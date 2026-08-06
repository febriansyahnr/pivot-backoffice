package submerchant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Create		godoc
// @Summary		Endpoint for resending invitation email
// @Description	Endpoint for resending invitation email
// @ID			open-api-sub-merchant-resend-invitation
// @Tags		API - Sub-Merchants
// @Accept		json
// @Produce		json
// @Param		Request	body		merchant.ResendInvitationRequest true "JSON body to resend invitation"
// @Success		200  	{object}	response.OpenApiErrorNonSnap{data=map[string]string}
// @Failure		500  	{object}	response.OpenApiErrorNonSnap
// @Router		/open-api/v1/sub-merchants/users/resend-invitation [post]
// @Security 	Bearer
// @Header      all     {string}  INTERNAL-API-KEY "{"key": "value"}"
// @Header      all     {string}  X-SubMerchant-Id "{"key": "value"}"
func (c *SubMerchantInternalController) ResendInvitation(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internal/subMerchant/ResendInvitation")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType2(nil))

	// Merchant info from JWT
	merchantClaims, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, errors.New(constant.ErrMsgUnauthorized)))
		return
	}

	subMerchantId := ""
	httputil.BindSubmerchantID(r, &subMerchantId)
	if subMerchantId == "" {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrRequiredField(constant.HeaderXSubMerchantID))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrMissingSubMerchantId))
		return
	}

	request := &merchant.ResendInvitationRequest{
		MerchantId:       subMerchantId,
		ParentMerchantId: merchantClaims.MerchantId,
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidPayload(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("cannot unmarshal: %w", err)))
		return
	}

	if err := c.validate.StructCtx(ctx, request); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrFieldValidation(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.merchantSvc.SubMerchantResendInvitation(ctx, request); err != nil {
		switch {
		case errors.Is(err, constant.ErrMerchantNotAllowedPerformAction):
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, pkgErrs.New(response.HttpErrForbidden, errors.New(constant.ErrMessageV2RequestForbidden)))
		case errors.Is(err, constant.ErrUserNotFound):
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("user", request.Email))
		case errors.Is(err, constant.ErrMerchantNotFound):
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("sub-account", subMerchantId))
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, map[string]string{"message": "OK"})
}
