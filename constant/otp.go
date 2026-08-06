package constant

import (
	"context"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

type OTPIdentifier string

const (
	OTPIdentifierForgotPassword OTPIdentifier = "00e2a8f568ff592a6f19d15f6fece067"
	OTPIdentifierResetPIN       OTPIdentifier = "aaa85d70d6320537095bf5661e3feb53"
	OTPIdentifierUserLogin      OTPIdentifier = "63bd25c8db9724f610636fdf20d9a52a"
	OTPIdentifierFirstTimeLogin OTPIdentifier = "194fbeea2d29e7bd8106c33e12103b00"
	OTPIdentifierChangePassword OTPIdentifier = "717c10b2cccb36ca57ed94c149694d2b"
	TokenOTPNamespace                         = "token-otp"
	TokenFeature2FANamespace                  = "token-feature"
	OTPKeyFormatting                          = "backend-portal:otp-verification:%s:%s"
	OTPForgotPasswordName                     = "forgot-password"
	OTPResetPINName                           = "reset-pin"
	OTPUserLoginName                          = "user-login"
	OTPFirstTimeLoginName                     = "first-time-login"
	OTPChangePasswordName                     = "change-password"
	OTPHardCode                               = "123456"
)

func (o *OTPIdentifier) FeatureValidation(path string) bool {
	switch *o {
	default:
		return false

	case OTPIdentifierForgotPassword:
		return strings.Contains(path, "/auth/reset-password")

	case OTPIdentifierResetPIN:
		return strings.Contains(path, "/auth/reset-pin")

	case OTPIdentifierUserLogin:
		return strings.Contains(path, "/users/2fa/token")

	case OTPIdentifierFirstTimeLogin:
		return strings.Contains(path, "/users/activate")

	case OTPIdentifierChangePassword:
		return strings.Contains(path, "/change-password")
	}
}

func (o *OTPIdentifier) EmailSender() string {
	return config.DefaultEmailSender()
}

func (o *OTPIdentifier) FeatureName() string {
	switch *o {
	default:
		return ""

	case OTPIdentifierForgotPassword:
		return OTPForgotPasswordName

	case OTPIdentifierResetPIN:
		return OTPResetPINName

	case OTPIdentifierUserLogin:
		return OTPUserLoginName

	case OTPIdentifierFirstTimeLogin:
		return OTPFirstTimeLoginName

	case OTPIdentifierChangePassword:
		return OTPChangePasswordName
	}
}

func (o *OTPIdentifier) Event() string {
	switch *o {
	default:
		return ""

	case OTPIdentifierForgotPassword:
		return ResetPasswordEvent

	case OTPIdentifierResetPIN:
		return ResetPINEvent

	case OTPIdentifierUserLogin:
		return UserLoginEvent

	case OTPIdentifierFirstTimeLogin:
		return FirstTimeLoginEvent

	case OTPIdentifierChangePassword:
		return ChangePasswordEvent
	}
}

func (o *OTPIdentifier) MaxSendOTP() int {
	switch *o {
	default:
		return 0

	case OTPIdentifierForgotPassword:
		return config.OTPConfig().MaxSendResetPwd

	case OTPIdentifierResetPIN:
		return config.OTPConfig().MaxSendResetPIN

	case OTPIdentifierUserLogin:
		return config.OTPConfig().MaxSendUserLogin

	case OTPIdentifierFirstTimeLogin:
		return config.OTPConfig().FirstTimeLoginMaxSend

	case OTPIdentifierChangePassword:
		return config.OTPConfig().MaxSendChangePwd
	}
}

func (o *OTPIdentifier) MaxFailedAttempts() int {
	switch *o {
	default:
		return 0

	case OTPIdentifierForgotPassword:
		return config.OTPConfig().MaxFailedVerifyResetPwd

	case OTPIdentifierResetPIN:
		return config.OTPConfig().MaxFailedVerifyResetPIN

	case OTPIdentifierUserLogin:
		return config.OTPConfig().MaxFailedVerifyUserLogin

	case OTPIdentifierFirstTimeLogin:
		return config.OTPConfig().FirstTimeLoginMaxFailedVerify

	case OTPIdentifierChangePassword:
		return config.OTPConfig().MaxFailedVerifyChangePwd
	}
}

func (o *OTPIdentifier) ExpireDuration() time.Duration {
	switch *o {
	default:
		return 0

	case OTPIdentifierForgotPassword:
		return time.Duration(config.OTPConfig().ExpirationSecondsForgotPassword) * time.Second

	case OTPIdentifierResetPIN:
		return time.Duration(config.OTPConfig().ExpirationSecondsResetPIN) * time.Second

	case OTPIdentifierChangePassword:
		return time.Duration(config.OTPConfig().ExpirationSecondsChangePassword) * time.Second

	case OTPIdentifierUserLogin:
		return time.Duration(config.OTPConfig().ExpirationSecondsUserLogin) * time.Second

	case OTPIdentifierFirstTimeLogin:
		return time.Duration(config.OTPConfig().ExpirationSecondsFirstTimeLogin) * time.Second
	}
}

func (o *OTPIdentifier) GetResendDelaySeconds() int {
	switch *o {
	default:
		return config.OTPConfig().ResendDelaySecondsDefault

	case OTPIdentifierForgotPassword:
		return config.OTPConfig().ResendDelaySecondsForgotPassword

	case OTPIdentifierResetPIN:
		return config.OTPConfig().ResendDelaySecondsResetPIN

	case OTPIdentifierChangePassword:
		return config.OTPConfig().ResendDelaySecondsChangePassword

	case OTPIdentifierUserLogin:
		return config.OTPConfig().ResendDelaySecondsUserLogin

	case OTPIdentifierFirstTimeLogin:
		return config.OTPConfig().ResendDelaySecondsFirstTimeLogin
	}
}

func (o *OTPIdentifier) NumWaitAfterSendOTP() int {
	switch *o {
	default:
		return 0

	case OTPIdentifierUserLogin:
		return config.OTPConfig().UserLoginWaitAfterSend

	case OTPIdentifierFirstTimeLogin:
		return config.OTPConfig().FirstTimeLoginWaitAfterSend
	}
}

func (o *OTPIdentifier) WaitTimeDuration() time.Duration {
	switch *o {
	default:
		return 0

	case OTPIdentifierUserLogin:
		return time.Duration(config.OTPConfig().UserLoginWaitTimeMinute) * time.Minute

	case OTPIdentifierFirstTimeLogin:
		return time.Duration(config.OTPConfig().FirstTimeLoginWaitTimeMinute) * time.Minute
	}
}

func IsTestingAccount(ctx context.Context, env string, userId string, email string) (bool, error) {
	userEmailFlag := ffcontext.NewEvaluationContext(userId)
	userEmailFlag.AddCustomAttribute("environment", env)
	userEmailFlag.AddCustomAttribute("email", email)

	if bypassIsTrue, err := ffclient.BoolVariation("backend-portal-otp-bypass", userEmailFlag, false); err != nil {
		return false, err
	} else if bypassIsTrue {
		return true, nil
	}

	return ffclient.BoolVariation("backend-portal-otp-bypass-whitelist", userEmailFlag, false)
}

func IsAccountBypassOTP(ctx context.Context, env string, userId string, email string) (bool, error) {
	userEmailFlag := ffcontext.NewEvaluationContext(userId)
	userEmailFlag.AddCustomAttribute("environment", env)
	userEmailFlag.AddCustomAttribute("email", email)

	// Use boolean variation with targeting rules that check email patterns in Consul
	return ffclient.BoolVariation("backend-portal-account-otp-bypass", userEmailFlag, false)
}
