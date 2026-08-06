package user

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ForgotPassword		godoc
// @Summary				Forgot Password endpoint with email
// @Description			Endpoint Forgot Password with email, this process will send an OTP to the user email as an auth step before resetting the password.
// @ID					api-user-forgot-password
// @Tags				API - User
// @Accept				json
// @Produce				json
// @Param 				Request body			user.UserForgotPasswordRequest true "JSON Body for Forgot Password"
// @Success				200  		{object}	response.ApiResponse{data=user.User2FAResponse}
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/auth/forgot-password [post]
func (c *UserController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/ForgotPassword")
	defer segment.End()

	request := userModel.UserForgotPasswordRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	token, err := c.userSvc.ForgotPassword(ctx, request.Email)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, userModel.User2FAResponse{Token: token, ResendDelaySeconds: config.OTPConfig().ResendDelaySecondsForgotPassword})
}

// ResetPassword		godoc
// @Summary				Reset Password endpoint for users who forgot their password (new password).
// @Description			Reset Password endpoint requires authentication from the OTP verification process.
// @ID					api-user-reset-password
// @Tags				API - User
// @Accept				json
// @Produce				json
// @Param 				Request	body		user.UserResetPasswordRequest true "JSON Body for Reset Password"
// @Success				200  	{object}	response.ApiResponse
// @Failure				500  	{object}	response.ApiResponse
// @Router				/api/v1/auth/reset-password [patch]
func (c *UserController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/ResetPassword")
	defer segment.End()

	claims, _ := ctx.Value(constant.CtxTokenOTPKey).(*otpModel.TokenOTPClaims)

	request := userModel.UserResetPasswordRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.userSvc.ResetPassword(ctx, claims.UUID, request.Password); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, map[string]bool{"updated": true})
}
