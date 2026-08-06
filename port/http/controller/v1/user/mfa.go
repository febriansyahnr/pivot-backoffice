package user

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *UserController) EnrollTOTP(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/EnrollTOTP")
	defer segment.End()

	userAuth, ok := ctx.Value(constant.CtxUserInfoKey).(*model.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := model.EnrollTOTPRequest{
		UserId: userAuth.UUID,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.userSvc.EnrollTOTP(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}

func (h *UserController) ConfirmTOTP(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/ConfirmTOTP")
	defer segment.End()

	userAuth, ok := ctx.Value(constant.CtxUserInfoKey).(*model.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := model.ConfirmTOTPRequest{
		UserId: userAuth.UUID,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrMalformedRequestBodyPayload))
		return
	}
	if err := h.validate.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if valid, err := h.userSvc.ConfirmTOTP(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else if !valid {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidTOTPCode))

	} else {
		response.SendApiResponseOK(w, model.ConfirmTOTPResponse{Status: constant.TOTPStatusActive})
	}
}

func (h *UserController) SetPreferred2FAMethod(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/SetPreferred2FAMethod")
	defer segment.End()

	userAuth, ok := ctx.Value(constant.CtxUserInfoKey).(*model.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := model.SetPreferred2FAMethodRequest{
		UserId: userAuth.UUID,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.userSvc.SetPreferred2FAMethod(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}
