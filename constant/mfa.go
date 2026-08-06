package constant

import (
	"errors"
	"time"
)

type TwoFactorMethod string

const (
	TOTPStatusNotEnrolled = "NOT_ENROLLED"
	TOTPStatusEnrolled    = "ENROLLED"
	TOTPStatusActive      = "ACTIVE"
	TOTPStatusDisabled    = "DISABLED"
	TOTPTokenPrefixID     = "totp-token/"

	TwoFactorAuthMethodOTP  TwoFactorMethod = "OTP"
	TwoFactorAuthMethodTOTP TwoFactorMethod = "TOTP"
	TwoFactorAuthMethodAuto TwoFactorMethod = "AUTO"

	TOTPEnrollmentCacheKeyFmt                 = "backend-portal:authenticator:%s" // Use user id as parameter value
	TOTPEnrollmentCacheDuration time.Duration = 2 * time.Hour
)

var (
	ErrTOTPAlreadyActivated           = errors.New("TOTP is already activated")
	ErrTOTPNotActivated               = errors.New("TOTP is not activated")
	ErrTOTPNotEnrolled                = errors.New("TOTP not enrolled")
	ErrCreateTOTPSecretKey            = errors.New("unable to create TOTP secret for user")
	ErrGenerateQrImageFromTOTPDataURL = errors.New("unable to generate QR image from TOTP data URL")
	ErrInvalidTOTPCode                = errors.New("invalid TOTP code. please try again")
	ErrFeatureNotSupportTOTPAuth      = errors.New("feature does not support authentication using TOTP")
	ErrOTPTokenNotRegistered          = errors.New("token is not registered")
	ErrInvalidPreferred2FAMethod      = errors.New("invalid preferred 2FA method. must be OTP, TOTP, or AUTO")
	ErrTOTPRequiredButNotActive       = errors.New("TOTP method selected but not activated. please enroll TOTP first")
)
