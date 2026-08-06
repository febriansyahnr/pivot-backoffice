package merchantTopUp

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Topup			godoc
// @Summary			Merchant Top-up
// @Description		Merchant Top-up
// @ID				api-merchant-top-up
// @Tags			API - Merchant Top Up
// @Accept			json
// @Produce			json
// @Param			Request	body		merchantTopUp.MerchantTopUpRequest true "JSON Body for merchant top up"
// @Success			200  	{object}	response.ApiResponse{data=merchantTopUp.MerchantTopUpResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/{usecase}/top-up [post]
// @Security		Bearer
func (c *handler) Topup(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchantTopUp/Topup")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	payload := model.MerchantTopUpRequest{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	payload.AccountName, _ = ctx.Value(constant.CtxAccountName).(string) // Filled in when mapping a route

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if reference, err := c.service.FindOrCreate(ctx, user.MerchantId, payload.AccountName, payload.PaymentMethodID); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, reference.ToResponse())
	}
}
