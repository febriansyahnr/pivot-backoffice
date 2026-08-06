package ipWhitelistController

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *IPWhitelistConfigurationController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/IPWhitelist/Update")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	id := chi.URLParam(r, "id")
	_, err := uuid.Parse(id)
	if err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	var payload ipwhitelistModel.UpdateIPWhitelistConfiguration
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	payload.ID = id
	payload.MerchantID = user.MerchantId
	if err := c.validator.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	config, err := c.svc.Update(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, config.ToResponseModel())

}
