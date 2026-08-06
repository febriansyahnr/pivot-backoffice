package user

import (
	"context"
	"encoding/json"
	e "errors"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Login		godoc
// @Summary		Login endpoint with email and password
// @Description	Login endpoint with email and password
// @ID			api-user-login
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.UserLoginRequest true "JSON Body for Login"
// @Success		200  	{object}	response.ApiResponse{data=user.UserLoggedInResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/auth/login [post]
func (c *UserController) Login(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/Login")
	defer segment.End()

	var payload userModel.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	deviceIdentifier := r.Header.Get(constant.HeaderDeviceIdentifier)
	if deviceIdentifier == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrDeviceIdentifierRequired))
		return
	}

	// Add user-agent to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserAgentKey, r.UserAgent())

	// Add user device identifier to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserDeviceIdentifierKey, deviceIdentifier)

	// login without login
	user, signedToken, err := c.userSvc.Login(ctx, &payload)
	if err != nil {
		if e.Is(err, constant.ErrNeed2FAChallengeForLogin) {
			// If got 2FA challenge for login
			token, err := c.userSvc.LoginWithOTP(ctx, payload.Email, payload.Password)
			if err != nil {
				response.SendApiResponseError(ctx, w, err)
				return
			}

			// Need to fetch user since it's nil when 2FA is required
			user, err := c.userSvc.FindUserByEmail(ctx, payload.Email)
			if err != nil {
				response.SendApiResponseError(ctx, w, err)
				return
			}

			isTOTPActive := false
			if user != nil {
				totpData, err := c.userSvc.FindUserTOTPDataByID(ctx, user.UUID)
				if err != nil {
					response.SendApiResponseError(ctx, w, err)
					return
				}
				if totpData != nil && totpData.TOTPStatus == constant.TOTPStatusActive {
					isTOTPActive = true
				}
			}

			twoFAMethod, resendDelaySeconds := constant.TwoFactorAuthMethodOTP, config.OTPConfig().ResendDelaySecondsUserLogin

			if strings.HasPrefix(token, constant.TOTPTokenPrefixID) {
				token = strings.Replace(token, constant.TOTPTokenPrefixID, "", 1)
				twoFAMethod, resendDelaySeconds = constant.TwoFactorAuthMethodTOTP, 0
			}

			response.SendApiResponseOK(w, userModel.User2FAResponse{
				Token:               token,
				ResendDelaySeconds:  resendDelaySeconds,
				TwoFactorAuthMethod: twoFAMethod,
				IsTOTPActive:        isTOTPActive,
			})
			return
		}

		if e.Is(err, constant.ErrBlockedTooManyAttempts) {
			// publish activity, do nothing on error
			payload.Password = "********"
			_ = c.rabbitMqExt.PublishActivity(
				ctx,
				&user.MerchantId,
				&user.UUID,
				constant.TagAccount,
				constant.ErrBlockedTooManyAttempts.Error(),
				payload,
			)
		}

		response.SendApiResponseError(ctx, w, err)
		return
	}

	if user == nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, constant.ErrUserNotFound))
		return
	}

	// fill response
	res := userModel.UserLoggedInResponse{
		UserInfo:     user.ToResponse(),
		AccessToken:  signedToken,
		RefreshToken: user.RefreshToken.String,
	}

	// publish activity, do nothing on error
	payload.Password = "********"
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagAccount,
		constant.ActivityUserLogin,
		payload,
	)

	response.SendApiResponseOK(w, res)
}

// LoginWithOTP	godoc
// @Summary		Login endpoint with email and password (2FA).
// @Description	Login endpoint with email and password will send an OTP for 2FA verification if the provided credentials are valid.
// @ID			api-user-login-otp
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.UserLoginRequest true "JSON Body for Login"
// @Success		200  	{object}	response.ApiResponse{data=user.User2FAResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/auth/login 	[post]
func (c *UserController) LoginWithOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/LoginWithOTP")
	defer segment.End()

	var payload userModel.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Add user-agent to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserAgentKey, r.UserAgent())

	// Add user device identifier to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserDeviceIdentifierKey, r.Header.Get(constant.HeaderDeviceIdentifier))

	token, err := c.userSvc.LoginWithOTP(ctx, payload.Email, payload.Password)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Check user's TOTP status for response
	user, err := c.userSvc.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	isTOTPActive := false
	if user != nil {
		totpData, err := c.userSvc.FindUserTOTPDataByID(ctx, user.UUID)
		if err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}
		if totpData != nil && totpData.TOTPStatus == constant.TOTPStatusActive {
			isTOTPActive = true
		}
	}

	twoFAMethod, resendDelaySeconds := constant.TwoFactorAuthMethodOTP, config.OTPConfig().ResendDelaySecondsUserLogin

	if strings.HasPrefix(token, constant.TOTPTokenPrefixID) {
		token = strings.Replace(token, constant.TOTPTokenPrefixID, "", 1)
		twoFAMethod, resendDelaySeconds = constant.TwoFactorAuthMethodTOTP, 0
	}

	response.SendApiResponseOK(w, userModel.User2FAResponse{
		Token:               token,
		ResendDelaySeconds:  resendDelaySeconds,
		TwoFactorAuthMethod: twoFAMethod,
		IsTOTPActive:        isTOTPActive,
	})
}

// SessionFromLogin2FA godoc
// @Summary		SessionFromLogin2FA endpoint is used to obtain access token, refresh token, and user information.
// @Description	SessionFromLogin2FA endpoint is used to obtain access token, refresh token, and user information.						 -
// @ID			api-users-2fa-token
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Success		200  	{object}		response.ApiResponse{data=user.UserLoggedInResponse}
// @Failure		500  	{object}		response.ApiResponse
// @Router		/api/v1/users/2fa/token	[get]
// @Security 	Bearer
func (c *UserController) SessionFromLogin2FA(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/SessionFromLogin2FA")
	defer segment.End()

	claims, _ := ctx.Value(constant.CtxTokenOTPKey).(*otpModel.TokenOTPClaims)

	// Add user-agent to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserAgentKey, r.UserAgent())

	// Add user device identifier to context.WithValue
	ctx = context.WithValue(ctx, constant.CtxUserDeviceIdentifierKey, r.Header.Get(constant.HeaderDeviceIdentifier))
	ctx = context.WithValue(ctx, constant.CtxIsRemember, r.Header.Get(constant.XIsRemember))
	ctx = context.WithValue(ctx, constant.CtxUserIPKey, r.Header.Get(constant.XIPInfoKey))

	user, signedToken, err := c.userSvc.GenerateTokenFromLogin2FA(ctx, claims.UUID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, userModel.UserLoggedInResponse{
		UserInfo:     user.ToResponse(),
		AccessToken:  signedToken,
		RefreshToken: user.RefreshToken.String,
	})
}
