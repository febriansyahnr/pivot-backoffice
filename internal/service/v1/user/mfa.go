package user

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/encrypt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"golang.org/x/sync/errgroup"
)

// FindUserTOTPDataByID is a function used to retrieve TOTP data along with additional information.
func (s *UserService) FindUserTOTPDataByID(ctx context.Context, userId string) (*model.UserTOTPData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/FindUserTOTPDataByID")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	user, err := s.userRepo.FindUserTOTPDataByID(ctx, userId)
	if err != nil {
		s.logger.Error(ctx, "Failed while find user TOTP data by id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if user == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrUserNotFound)
	}
	return user, nil
}

// EnrollTOTP is a function that registers a user to enable multi-factor authentication using Time-based One-Time Password (TOTP).
// To activate this feature, the user must first confirm the TOTP code during the confirmation step.
func (s *UserService) EnrollTOTP(ctx context.Context, request model.EnrollTOTPRequest) (*model.EnrollTOTPResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/EnrollTOTP")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	user, err := s.FindUserTOTPDataByID(ctx, request.UserId)
	if err != nil {
		return nil, err
	}

	// Set to default value for qr code size is not specified
	if request.QrCodeSize == 0 {
		request.QrCodeSize = 256
	}

	groups := new(errgroup.Group)

	var (
		otpKey                   *otp.Key
		encryptVersion           int
		encryptionKey, qrCodePNG []byte
	)

	// Generate TOTP secret & QR code
	// Note: This process will be wrapped as an adapter to streamline operations and minimize the core services
	//       awareness of the underlying dependencies.
	groups.Go(func() (errGroup error) {

		// Generate TOTP secret
		otpKey, errGroup = totp.Generate(totp.GenerateOpts{
			Issuer:      s.config.MultiFactorAuth.TimeBasedOTP.TOTPIssuer,
			AccountName: user.Email,
			SecretSize:  s.config.MultiFactorAuth.TimeBasedOTP.TOTPSecretSize,
			Period:      s.config.MultiFactorAuth.TimeBasedOTP.TOTPPeriodInSeconds,
			Algorithm:   otp.AlgorithmSHA1, // Support for multiple authenticator apps and manual key input
			Digits:      otp.DigitsSix,
		})
		if errGroup != nil {
			s.logger.Error(ctx, "Failed while generate new a totp key", logger.Error(errGroup))
			return pkgErrs.New(response.HttpErrInternal, constant.ErrCreateTOTPSecretKey)
		}

		// Generate QR code
		qrCodePNG, errGroup = qrcode.Encode(otpKey.String(), request.GetQrCodeLevel(), request.QrCodeSize)
		if errGroup != nil {
			s.logger.Error(ctx, "Failed while encode TOTP data URL to QR image", logger.Error(errGroup))
			return pkgErrs.New(response.HttpErrInternal, constant.ErrGenerateQrImageFromTOTPDataURL)
		}
		return nil
	})

	// Getting the encryption key in Vault
	groups.Go(func() (errGroup error) {
		secret, errGroup := s.encryptionKey.GetSecretKeyString(ctx, s.config.Vault.Secrets.UserEncryptionKey.KeyName)
		if errGroup != nil {
			s.logger.Error(ctx, "Failed while getting user encryption key", logger.Error(errGroup))
			return pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}
		encryptVersion = secret.Version

		if encryptionKey, errGroup = base64.StdEncoding.DecodeString(secret.Value); errGroup != nil {
			s.logger.Error(ctx, "Failed while base64 decode user encryption key", logger.Error(errGroup))
			return pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}
		return nil
	})

	if err := groups.Wait(); err != nil {
		return nil, err
	}

	// Encrypt the secret for storage
	wrappedSecret, err := encrypt.AesGcmEncryptToBase64String(otpKey.Secret(), encryptionKey)
	if err != nil {
		s.logger.Error(ctx, "Failed while encrypt TOTP secret key", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	// The newly generated secret key will be stored in cache for a limited period, and the changes will be applied to persistent data
	// once the user confirms the enrollment.
	enrollmentData := model.TOTPEnrollmentData{
		WrappedSecretKey: wrappedSecret,
		EncryptVersion:   encryptVersion,
	}
	enrollmentKey := fmt.Sprintf(constant.TOTPEnrollmentCacheKeyFmt, user.UserId)
	if err = s.redis.Set(ctx, enrollmentKey, enrollmentData, constant.TOTPEnrollmentCacheDuration).Err(); err != nil {
		s.logger.Error(ctx, "Failed to store TOTP enrollment in cache", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	// Response containing the QR code and secret key to be displayed to the user.
	return &model.EnrollTOTPResponse{
		QRCodeDataURL: fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(qrCodePNG)), // Convert PNG to data URL
		SecretKey:     otpKey.Secret(),
	}, nil
}

func (s *UserService) ConfirmTOTP(ctx context.Context, request model.ConfirmTOTPRequest) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ConfirmTOTP")
	defer segment.End()

	enrollmentData := model.TOTPEnrollmentData{}
	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	// During the TOTP enrollment process, the enrollment data is stored in cache
	// and prepared for confirmation by entering the OTP code generated by the Authenticator App
	enrollmentKey := fmt.Sprintf(constant.TOTPEnrollmentCacheKeyFmt, request.UserId)
	if err := s.redis.Get(ctx, enrollmentKey).Scan(&enrollmentData); err != nil {
		if errors.Is(err, redisExt.ErrNil) {
			return false, pkgErrs.New(response.HttpErrNotFound, errors.New("TOTP enrollment has not been completed or has expired"))
		}
		s.logger.Error(ctx, "Failed to get TOTP enrollment data in cache", logger.Error(err))
		return false, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	// Verify the OTP code generated by the Authenticator App against the one generated by the server
	// using the previously configured (shared) secret key.
	verifyOtpRequest := &otpModel.VerifyTOTPRequest{
		WrappedSecret:  &enrollmentData.WrappedSecretKey,
		EncryptVersion: &enrollmentData.EncryptVersion,
		Code:           request.OTP,
	}
	if valid, err := s.otpSvc.ValidateTOTPCode(ctx, verifyOtpRequest); err != nil || !valid {
		return valid, err
	}

	// If the OTP code is valid, the secret key and the encryption version used will be permanently stored in the database.
	updateRequest := &model.UpdateUserTOTPDataRequest{
		UserId:         request.UserId,
		WrappedSecret:  enrollmentData.WrappedSecretKey,
		EncryptVersion: enrollmentData.EncryptVersion,
		Status:         constant.TOTPStatusActive,
	}
	if err := s.userRepo.UpdateUserTOTPData(ctx, updateRequest); err != nil {
		s.logger.Error(ctx, "Failed while update user TOTP data on confirm enrollment", logger.Error(err))
		return false, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	_ = s.redis.Del(ctx, enrollmentKey)

	return true, nil
}

// SetPreferred2FAMethod sets the user's preferred 2FA method
func (s *UserService) SetPreferred2FAMethod(ctx context.Context, request model.SetPreferred2FAMethodRequest) (*model.SetPreferred2FAMethodResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/SetPreferred2FAMethod")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	// Validate the 2FA method
	if request.Preferred2FAMethod != string(constant.TwoFactorAuthMethodOTP) &&
		request.Preferred2FAMethod != string(constant.TwoFactorAuthMethodTOTP) {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidPreferred2FAMethod)
	}

	if request.Preferred2FAMethod == string(constant.TwoFactorAuthMethodTOTP) {
		totpData, err := s.FindUserTOTPDataByID(ctx, request.UserId)
		if err != nil {
			return nil, err
		}
		if totpData.TOTPStatus != constant.TOTPStatusActive {
			return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrTOTPRequiredButNotActive)
		}
	}

	// Update preferred 2FA method
	if err := s.userRepo.UpdateUserPreferred2FAMethod(ctx, request.UserId, request.Preferred2FAMethod); err != nil {
		s.logger.Error(ctx, "Failed while updating user preferred 2FA method", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	return &model.SetPreferred2FAMethodResponse{
		Preferred2FAMethod: request.Preferred2FAMethod,
		Updated:            true,
	}, nil
}
