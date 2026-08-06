package tnc

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *TNCSigningController) Status(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/tnc/Status")
	defer segment.End()

	claims, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok || claims == nil {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	if claims.MerchantId == "" {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
		return
	}

	status, err := c.service.GetTNCSigningStatus(ctx, claims.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, status)
}
