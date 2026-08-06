package xbPayoutController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *xbPayoutController) Confirm(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/Confirm")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// get id from url path
	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	requestPayload := xbModel.ConfirmPayoutRequest{
		MerchantId: user.MerchantId,
		PayoutId:   id,
		ApprovedBy: user.UUID,
	}

	resp, err := c.xbPayoutSvc.Confirm(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
