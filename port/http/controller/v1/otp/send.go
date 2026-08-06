package otp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Send					godoc
// @Summary				Send endpoint for send an OTP email.
// @Description			This endpoint is used to send an OTP email according to the selected event.
// @ID					api-user-otp-send
// @Tags				API - User
// @Accept				json
// @Produce				json
// @Param 				Request	body		otp.SendOTPReq true "JSON Body for Send"
// @Success				200  	{object}	response.ApiResponse{data=otp.User2FARes}
// @Failure				500  	{object}	response.ApiResponse
// @Router				/api/v1/auth/otp/send [post]
func (h *handler) Send(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/otp/Send")
	defer segment.End()

	request := otpModel.SendOTPReq{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if request.TwoFactorAuthMethod == "" {
		request.TwoFactorAuthMethod = constant.TwoFactorAuthMethodOTP
	}
	if err := h.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	token, err := h.service.SendGenerateOTPCode(ctx, &otpModel.GenerateOTPCodeRequest{
		UserEmail:           request.Email,
		Event:               request.Event,
		TwoFactorAuthMethod: request.TwoFactorAuthMethod,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// Check user's TOTP status
	user, err := h.userSvc.FindUserByEmail(ctx, request.Email)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	isTOTPActive := false
	if user != nil {
		totpData, err := h.userSvc.FindUserTOTPDataByID(ctx, user.UUID)
		if err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}
		if totpData != nil && totpData.TOTPStatus == constant.TOTPStatusActive {
			isTOTPActive = true
		}
	}

	twoFAMethod, resendDelaySeconds := constant.TwoFactorAuthMethodOTP, request.Event.GetResendDelaySeconds()

	if strings.HasPrefix(token, constant.TOTPTokenPrefixID) {
		token = strings.Replace(token, constant.TOTPTokenPrefixID, "", 1)
		twoFAMethod, resendDelaySeconds = constant.TwoFactorAuthMethodTOTP, 0
	}
	response.SendApiResponseOK(w, otpModel.User2FARes{
		Token:               token,
		ResendDelaySeconds:  resendDelaySeconds,
		TwoFactorAuthMethod: twoFAMethod,
		IsTOTPActive:        isTOTPActive,
	})
}
