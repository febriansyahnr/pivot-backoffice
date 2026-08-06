package credential

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Get			godoc
// @Summary		To display client ID and client secrets summary
// @Description	To display client ID and client secrets summary
// @ID			settings-credentials
// @Tags		Settings - Credentials
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse{data=credential.CredentialDashboardResp}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/settings/credentials [get]
// @Security	Bearer
func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/setting/credential/Get")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &credential.CredentialDashboardReq{
		Info:       r,
		UserID:     user.UUID,
		MerchantID: user.MerchantId,
	}
	if resp, err := h.service.Get(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
