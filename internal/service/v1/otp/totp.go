package otp

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/encrypt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	// TOTPSkew allows ±1 time step (30 seconds) for verification
	TOTPSkew = 1
	// Period is number of seconds a TOTP hash is valid for.
	TOTPPeriod = 30
)

// ValidateTOTPCode verifies a TOTP code using constant-time comparison
func (s *service) ValidateTOTPCode(ctx context.Context, request *otpModel.VerifyTOTPRequest) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/otp/ValidateTOTPCode")
	defer segment.End()

	if request.EncryptVersion == nil || request.WrappedSecret == nil {
		return false, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("user has not registered for TOTP"))
	}

	// Decrypt the secret key
	secret, err := s.userEncryptionKey.GetSecretKeyVersionString(ctx, *request.EncryptVersion, s.config.Vault.Secrets.UserEncryptionKey.KeyName)
	if err != nil {
		return false, pkgErrs.New(response.HttpErrInternal, err)
	}
	encryptionKey, err := base64.StdEncoding.DecodeString(secret.Value)
	if err != nil {
		return false, pkgErrs.New(response.HttpErrInternal, err)
	}
	totpSecretKey, err := encrypt.AesGcmDecryptBase64String(*request.WrappedSecret, encryptionKey)
	if err != nil {
		return false, pkgErrs.New(response.HttpErrInternal, err)
	}

	period := s.config.MultiFactorAuth.TimeBasedOTP.TOTPPeriodInSeconds
	if period == 0 {
		period = TOTPPeriod
	}

	// Validates a TOTP given a user specified time and custom options
	validCode, err := totp.ValidateCustom(request.Code, totpSecretKey, time.Now().UTC(), totp.ValidateOpts{
		Period:    period,
		Skew:      TOTPSkew,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1, // Support for multiple authenticator apps and manual key input
	})
	if err != nil {
		s.logger.Error(ctx, "Failed while validate totp code", logger.Error(err))
		return false, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}
	return validCode, nil
}
