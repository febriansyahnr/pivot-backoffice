package constant_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func TestOTPEmailSender(t *testing.T) {
	val := OTPIdentifierForgotPassword
	assert.Equal(t, "sender name <email@example.com>", val.EmailSender())
}

func TestConstOTPConversion(t *testing.T) {
	tests := []struct {
		input                   OTPIdentifier
		path                    string
		wantFeatureName         string
		wantEvent               string
		wantMaxSendOTP          int
		wantMaxFailed           int
		wantExpire              time.Duration
		wantNumWaitAfterSendOTP int
		wantWaitTimeDuration    time.Duration
		wantFeatureValidation   bool
		wantResendDelaySeconds  int
	}{
		{wantResendDelaySeconds: 30},
		{
			input:                  OTPIdentifierForgotPassword,
			path:                   "/api/v1/auth/reset-password",
			wantFeatureName:        "forgot-password",
			wantEvent:              ResetPasswordEvent,
			wantMaxSendOTP:         5,
			wantMaxFailed:          3,
			wantExpire:             5 * time.Minute,
			wantFeatureValidation:  true,
			wantResendDelaySeconds: 241,
		},
		{
			input:                  OTPIdentifierResetPIN,
			path:                   "/api/v1/auth/reset-pin",
			wantFeatureName:        "reset-pin",
			wantEvent:              ResetPINEvent,
			wantMaxSendOTP:         10,
			wantMaxFailed:          6,
			wantExpire:             5 * time.Minute,
			wantFeatureValidation:  true,
			wantResendDelaySeconds: 242,
		},
		{
			input:                   OTPIdentifierFirstTimeLogin,
			path:                    "/api/v1/users/activate",
			wantFeatureName:         "first-time-login",
			wantEvent:               FirstTimeLoginEvent,
			wantMaxSendOTP:          21,
			wantMaxFailed:           22,
			wantExpire:              5 * time.Minute,
			wantNumWaitAfterSendOTP: 23,
			wantWaitTimeDuration:    24 * time.Minute,
			wantFeatureValidation:   true,
			wantResendDelaySeconds:  245,
		},
		{
			input:                   OTPIdentifierUserLogin,
			path:                    "/users/2fa/token",
			wantFeatureName:         "user-login",
			wantEvent:               UserLoginEvent,
			wantMaxSendOTP:          10,
			wantMaxFailed:           3,
			wantExpire:              5 * time.Minute,
			wantNumWaitAfterSendOTP: 3,
			wantWaitTimeDuration:    60 * time.Minute,
			wantFeatureValidation:   true,
			wantResendDelaySeconds:  244,
		},
		{
			input:                  OTPIdentifierChangePassword,
			path:                   "/api/v1/change-password",
			wantFeatureName:        "change-password",
			wantEvent:              ChangePasswordEvent,
			wantMaxSendOTP:         5,
			wantMaxFailed:          3,
			wantExpire:             5 * time.Minute,
			wantFeatureValidation:  true,
			wantResendDelaySeconds: 243,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantFeatureName, test.input.FeatureName())
		assert.Equal(t, test.wantEvent, test.input.Event())
		assert.Equal(t, test.wantMaxSendOTP, test.input.MaxSendOTP())
		assert.Equal(t, test.wantMaxFailed, test.input.MaxFailedAttempts())
		assert.Equal(t, test.wantExpire, test.input.ExpireDuration())
		assert.Equal(t, test.wantNumWaitAfterSendOTP, test.input.NumWaitAfterSendOTP())
		assert.Equal(t, test.wantWaitTimeDuration, test.input.WaitTimeDuration())
		assert.Equal(t, test.wantFeatureValidation, test.input.FeatureValidation(test.path))
		assert.Equal(t, test.wantResendDelaySeconds, test.input.GetResendDelaySeconds())
	}
}

func TestIntegrationIsTestingEmail(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	ctx := context.Background()

	logger, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)
	consulContainer, consulURL, err := test.SetupConsul(ctx)
	assert.NoError(t, err)
	test.SetupFeatureFlag(consulURL)
	test.SetupGoff(ctx, consulURL, pdkLogger)

	defer logger.Sync()
	defer pdkLogger.Sync()
	defer ffclient.Close()
	defer consulContainer.Terminate(ctx)

	tests := []struct {
		name     string
		env      string
		email    string
		want     bool
		wantErr  error
		modifier func()
	}{
		{
			name:    "Given when env production and email is not whitelist then return false",
			env:     "production",
			email:   "notwhitelisted@example.com",
			want:    false,
			wantErr: nil,
		},
		{
			name:    "Given when env production and email is whitelist then return true",
			env:     "production",
			email:   "whitelisted@example.com",
			want:    true,
			wantErr: nil,
		},
		{
			name:    "Given when env staging then return true",
			env:     "test",
			email:   "random@example.com",
			want:    true,
			wantErr: nil,
		},
		{
			name:    "Given error when check otp bypass then return error",
			env:     "error",
			email:   "random@example.com",
			want:    false,
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.modifier != nil {
				test.modifier()
			}
			bypassIsTrue, err := IsTestingAccount(context.Background(), test.env, uuid.NewString(), test.email)

			assert.Equal(t, err, test.wantErr)
			assert.Equal(t, test.want, bypassIsTrue)
		})
	}
}
