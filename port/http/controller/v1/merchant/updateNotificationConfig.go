package merchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UpdateNotificationConfig godoc
// @Summary				Update merchant notification configuration
// @Description			Update merchant notification configuration
// @ID					merchant-update-notification-config
// @Tags				API - Merchant
// @Accept				json
// @Produce				json
// @Param 				id	path		string true "Merchant ID"
// @Param 				request	body		merchant.MerchantNotificationConfig true "Request Body"
// @Success				200  		{object}	response.ApiResponse
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/merchants/:id/notification-config [put]
// @Security			Bearer
func (c *MerchantController) UpdateNotificationConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/UpdateNotificationConfig")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var req merchantModel.MerchantNotificationConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(req); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.merchantSvc.UpdateNotificationConfig(ctx, user.MerchantId, &req)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
