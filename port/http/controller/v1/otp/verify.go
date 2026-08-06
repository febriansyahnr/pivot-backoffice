package otp

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Verify				godoc
// @Summary				Verify endpoint for validate OTP
// @Description			Verify endpoint for validate OTP. This endpoint requires a one-time token generated from the OTP request process.
// @ID					api-user-otp-verify
// @Tags				API - User
// @Accept				json
// @Produce				json
// @Param 				Request	body		otp.VerifyOTPReq true "JSON Body for Verify OTP"
// @Success				200  	{object}	response.ApiResponse{data=otp.User2FARes}
// @Failure				500  	{object}	response.ApiResponse
// @Router				/api/v1/auth/otp/verify [post]
func (h *handler) Verify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/otp/Verify")
	defer segment.End()

	claims, _ := ctx.Value(constant.CtxTokenOTPKey).(*otpModel.TokenOTPClaims)

	request := otpModel.VerifyOTPReq{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	data := &otpModel.VerifyOTP{
		ID:                  claims.UUID, // User ID
		OTPCode:             request.OTP,
		Identifier:          claims.Identifier,
		Email:               claims.VerifyOTP.Email,
		TwoFactorAuthMethod: claims.VerifyOTP.TwoFactorAuthMethod,
	}
	if data.TwoFactorAuthMethod == "" {
		data.TwoFactorAuthMethod = constant.TwoFactorAuthMethodOTP
	}
	token, err := h.service.ValidateOTPCode(ctx, data)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, otpModel.User2FARes{Token: token})
}
