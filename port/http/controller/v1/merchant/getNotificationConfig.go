package merchant

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetNotificationConfig godoc
// @Summary				Get merchant notification configuration
// @Description			Get merchant notification configuration
// @ID					merchant-get-notification-config
// @Tags				API - Merchant
// @Accept				json
// @Produce				json
// @Param 				id	path		string true "Merchant ID"
// @Success				200  		{object}	response.ApiResponse{data=merchant.MerchantNotificationConfig}
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/merchants/:id/notification-config [get]
// @Security			Bearer
func (c *MerchantController) GetNotificationConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/GetNotificationConfig")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	res, err := c.merchantSvc.GetNotificationConfig(ctx, user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, res)
}
