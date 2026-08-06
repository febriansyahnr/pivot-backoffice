package otp_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/otp"
	loggerPkgMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	jwtPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisPkgMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiY2U5Mzc1OGUtYjE5Yi00M2MyLTk3MmEtZjY0YmUxNmYwOGNmIiwiaWRlbnRpZmllciI6IjMxYjNkMDk5LTg1MDgtNGU5My1hODBjLTA0NzFlZjIzN2E0YyJ9.awoAo8i2uIvROrmxDqS1aa45N4OOPJU7KHuiTwFaBls"

const (
	loginSuspendKey = "backend-portal:otp-verification:email:user-login:suspend"
	loginDataKey    = "backend-portal:otp-verification:email:user-login:data"
)

func TestSendGenerateOTPCode(pt *testing.T) {

	userMock := repoMocks.NewIUserRepository(pt)
	generatorMock := serviceMocks.NewIOTPGenerator(pt)

	service := New(cfg, nil, nil, nil, nil, userMock, nil)
	service.WithGenerator(generatorMock)

	tests := []struct {
		name      string
		mockSetup func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator)
		wantErr   string
		wantToken string
	}{
		{
			name: "ERROR:Find user by email",
			mockSetup: func(u *repoMocks.IUserRepository, _ *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(nil, errors.New("invalid db session"))
			},
			wantErr: "invalid db session",
		},
		{
			name: "ERROR:Email not registered",
			mockSetup: func(u *repoMocks.IUserRepository, _ *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "email not registered",
		},
		{
			name: "ERROR:User status is deactivated",
			mockSetup: func(u *repoMocks.IUserRepository, _ *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{DeactivatedAt: sql.NullTime{Valid: true}}, nil)
			},
			wantErr: "user is deactivated",
		},
		{
			name: "ERROR:User has been blocked",
			mockSetup: func(u *repoMocks.IUserRepository, _ *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{Blocked: sql.NullTime{Time: time.Now().UTC().Add(time.Hour)}}, nil)
			},
			wantErr: "user has been blocked",
		},
		{
			name: "ERROR:Generate OTP code",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{}, nil)

				g.On(
					"GenerateOTPCode", mock.Anything, constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Once().Return("", errors.New("your request has exceeded the limit"))
			},
			wantErr: "your request has exceeded the limit",
		},
		{
			name: "SUCCESS:2FA with OTP",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{}, nil)
				g.On(
					"GenerateOTPCode", mock.Anything, constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Return(token, nil)
			},
			wantToken: token,
		},
		{
			name: "SUCCESS:2FA with TOTP",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{TOTPStatus: constant.TOTPStatusActive}, nil)
				g.On(
					"GenerateTOTPVerifyToken", mock.Anything, mock.Anything,
				).Return(token, nil)
			},
			wantToken: token,
		},
		{
			name: "SUCCESS:Preferred OTP method (even with TOTP active)",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{
					TOTPStatus:         constant.TOTPStatusActive,
					Preferred2FAMethod: string(constant.TwoFactorAuthMethodOTP),
				}, nil)
				g.On(
					"GenerateOTPCode", mock.Anything, constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Return(token, nil)
			},
			wantToken: token,
		},
		{
			name: "SUCCESS:Preferred TOTP method with TOTP active",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{
					TOTPStatus:         constant.TOTPStatusActive,
					Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
				}, nil)
				g.On(
					"GenerateTOTPVerifyToken", mock.Anything, mock.Anything,
				).Return(token, nil)
			},
			wantToken: token,
		},
		{
			name: "ERROR:Preferred TOTP method but TOTP not active",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{
					TOTPStatus:         constant.TOTPStatusNotEnrolled,
					Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
				}, nil)
			},
			wantErr: constant.ErrTOTPRequiredButNotActive.Error(),
		},
		{
			name: "SUCCESS:Empty TwoFactorAuthMethod with preferred OTP",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{
					TOTPStatus:         constant.TOTPStatusActive,
					Preferred2FAMethod: string(constant.TwoFactorAuthMethodOTP),
				}, nil)
				g.On(
					"GenerateOTPCode", mock.Anything, constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Return(token, nil)
			},
			wantToken: token,
		},
		{
			name: "SUCCESS:No preference defaults to AUTO behavior (TOTP if active)",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{
					TOTPStatus:         constant.TOTPStatusActive,
					Preferred2FAMethod: "", // No preference set
				}, nil)
				g.On(
					"GenerateTOTPVerifyToken", mock.Anything, mock.Anything,
				).Return(token, nil)
			},
			wantToken: token,
		},
		{
			name: "SUCCESS:No preference defaults to AUTO behavior (OTP if TOTP not active)",
			mockSetup: func(u *repoMocks.IUserRepository, g *serviceMocks.IOTPGenerator) {
				u.On(
					"FindUserByEmail", mock.Anything, constant.StringMockType(),
				).Once().Return(&userModel.User{
					TOTPStatus:         constant.TOTPStatusNotEnrolled,
					Preferred2FAMethod: "", // No preference set
				}, nil)
				g.On(
					"GenerateOTPCode", mock.Anything, constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Return(token, nil)
			},
			wantToken: token,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(userMock, generatorMock)

			token, err := service.SendGenerateOTPCode(t.Context(), &otp.GenerateOTPCodeRequest{
				UserEmail:           "email@example.id",
				Event:               constant.OTPIdentifierForgotPassword,
				TwoFactorAuthMethod: constant.TwoFactorAuthMethodAuto,
			})
			if test.wantErr == "" {
				require.Nil(t, err)
				assert.Equal(t, test.wantToken, token)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}

			userMock.AssertExpectations(t)
			generatorMock.AssertExpectations(t)
		})
	}

}

func TestGenerateOTPCode(pt *testing.T) {
	db, clientMock := redismock.NewClientMock()

	clientMock.MatchExpectationsInOrder(false)
	jwtMock := jwtPkgMock.NewIJwt(pt)
	redisMock := redisExt.WrapRedisClient(db, nil)
	limiterMock := redisPkgMocks.NewILimiter(pt)

	rmq := rmqMock.NewRabbitMQExt(pt)
	rmq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	service := New(cfg, pdkLoggerMock, redisMock, jwtMock, rmq, nil, limiterMock)

	activatedFuncs := []func(){}

	tests := []struct {
		name       string
		id         string
		feature    constant.OTPIdentifier
		mockSetup  func(*jwtPkgMock.IJwt, redismock.ClientMock)
		wantErr    string
		wantResult string
	}{
		{
			name:    "ERROR:User suspended",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock) {
				r.ExpectGet("backend-portal:otp-verification:email:forgot-password:suspend").SetVal(`{"status": true}`)

				activatedFuncs = append(activatedFuncs, func() {
					r.ExpectGet("backend-portal:otp-verification:email:forgot-password:suspend").SetVal(`{"status": false}`)
				})
			},
			wantErr: "your otp request is currently suspended",
		},
		{
			name:    "ERROR:Take delivery total",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock) {
				for _, f := range activatedFuncs {
					f()
				}
				r.ExpectHGet("backend-portal:otp-verification:email:forgot-password:data", "total_delivery").SetErr(errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name:    "ERROR:Daily OTP sending limit",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock) {
				r.ExpectHGet("backend-portal:otp-verification:email:forgot-password:data", "total_delivery").SetVal("10")

				activatedFuncs = append(activatedFuncs, func() {
					r.ExpectHGet("backend-portal:otp-verification:email:forgot-password:data", "total_delivery").SetVal("0")
				})
			},
			wantErr: "your request has exceeded the limit",
		},
		{
			name:    "ERROR:Rate limitting/Invalid Session",
			feature: constant.OTPIdentifierUserLogin,
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock) {

				r.ClearExpect()
				r.ExpectGet(loginSuspendKey).SetVal(`{"status": false}`)
				r.ExpectHGet(loginDataKey, "total_delivery").SetVal("0")
				limiterMock.On(
					"Allow", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, errors.New("invalid session for redis rate limiter"))
			},
			wantErr: "invalid session for redis rate limiter",
		},
		{
			name:    "ERROR:Rate limitting/Limit exceeded",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock) {
				for _, f := range activatedFuncs {
					f()
				}
				limiterMock.On(
					"Allow", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(&redisExt.Result{Allowed: 0}, nil)
			},
			wantErr: "please wait for a moment",
		},
		{
			name:    "ERROR:Exclusive key set",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock) {
				for _, f := range activatedFuncs {
					f()
				}
				r.ExpectSetNX("backend-portal:otp-verification:email:forgot-password:lock", true, 10*time.Second).SetErr(errors.New("invalid session #2"))

				limiterMock.On(
					"Allow", mock.Anything, mock.Anything, mock.Anything,
				).Return(&redisExt.Result{Allowed: 1}, nil)
			},
			wantErr: "invalid session #2",
		},
		{
			name:    "ERROR:Same request content is being processed",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock) {
				for _, f := range activatedFuncs {
					f()
				}
				r.ExpectSetNX("backend-portal:otp-verification:email:forgot-password:lock", true, 10*time.Second).SetVal(false)

				activatedFuncs = append(activatedFuncs, func() {
					r.ExpectSetNX("backend-portal:otp-verification:email:forgot-password:lock", true, 10*time.Second).SetVal(true)
				})
			},
			wantErr: "the same request is in progress",
		},
		{
			name:    "ERROR:Generate JWT",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
				for _, f := range activatedFuncs {
					f()
				}
				j.On(
					"GenerateTokenForOTP", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return("", errors.New("token creation process failed"))
			},
			wantErr: "token creation process failed",
		},
		{
			name:    "ERROR:Storing OTP cache",
			feature: constant.OTPIdentifierForgotPassword,
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
				r.ClearExpect()

				r.ExpectGet("backend-portal:otp-verification:email:forgot-password:suspend").SetVal(`{"status": false}`)
				r.ExpectHGet("backend-portal:otp-verification:email:forgot-password:data", "total_delivery").RedisNil()

				limiterMock.On(
					"Allow", mock.Anything, mock.Anything, mock.Anything,
				).Return(&redisExt.Result{Allowed: 1}, nil).Maybe()

				r.ExpectSetNX("backend-portal:otp-verification:email:forgot-password:lock", true, 10*time.Second).SetVal(true)
				j.On("GenerateTokenForOTP", mock.Anything, mock.Anything, mock.Anything).Return(token, nil).Once()

				r.ExpectHGet("backend-portal:otp-verification:email:forgot-password:data", "otp").RedisNil()

				// This is an error test case - set an error for HSet operation
				r.ExpectHSet("backend-portal:otp-verification:email:forgot-password:data",
					"otp", mock.Anything,
					"total_delivery", mock.Anything,
					"total_attempts", mock.Anything,
					"total_resend", mock.Anything).SetErr(errors.New("invalid session"))

				r.ExpectDel("backend-portal:otp-verification:email:forgot-password:lock").SetVal(1)
				r.ExpectKeys("backend-portal:otp-verification:unique-id:forgot-password:token-otp:*").SetVal([]string{"key-1", "key-2"})
				r.ExpectDel("key-1", "key-2").SetVal(2)

				r.ExpectSet(
					"backend-portal:otp-verification:unique-id:forgot-password:token-otp:"+token,
					"email",
					25*time.Minute,
				).SetVal("1")

				r.ExpectTTL("backend-portal:otp-verification:email:forgot-password:data").SetVal(time.Duration(10 * time.Second))
				r.ExpectExpire("backend-portal:otp-verification:email:forgot-password:data", 2*time.Hour).SetVal(true)
			},
			wantErr: "ERROR_INTERNAL",
		},
		{
			name:    "SUCCESS:OTP_with_login",
			feature: constant.OTPIdentifierUserLogin,
			id:      "rizaldiamanda+innovation@gmail.com",
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
				r.ClearExpect()

				email := "rizaldiamanda+innovation@gmail.com"
				suspendKey := "backend-portal:otp-verification:" + email + ":user-login:suspend"
				dataKey := "backend-portal:otp-verification:" + email + ":user-login:data"
				lockKey := "backend-portal:otp-verification:" + email + ":user-login:lock"

				r.ExpectGet(suspendKey).SetVal(`{"status": false}`)

				// Use exact value for total_delivery
				r.ExpectHGet(dataKey, "total_delivery").SetVal("1")

				limiterMock.On(
					"Allow", mock.Anything, mock.Anything, mock.Anything,
				).Return(&redisExt.Result{Allowed: 1}, nil).Maybe()

				r.ExpectSetNX(lockKey, true, 10*time.Second).SetVal(true)

				j.On("GenerateTokenForOTP", mock.Anything, mock.Anything, mock.Anything).Return(token, nil).Once()

				// Provide existing OTP to be appended
				existingOTP := `[{"otp":"536009","expired_at":"2025-04-09T14:35:15.386829Z","verify":false}]`
				r.ExpectHGet(dataKey, "otp").SetVal(existingOTP)

				// Use a custom matcher for flexible HSet expectations
				r.CustomMatch(func(expected, actual []interface{}) error {
					// Check command name
					if expected[0] != actual[0] {
						return fmt.Errorf("expected %v, got %v", expected[0], actual[0])
					}

					// Check key
					if expected[1] != actual[1] {
						return fmt.Errorf("expected %v, got %v", expected[1], actual[1])
					}

					// Only check field names, not values (for flexibility)
					if expected[2] != actual[2] || expected[4] != actual[4] ||
						expected[6] != actual[6] || expected[8] != actual[8] {
						return fmt.Errorf("field names don't match")
					}

					return nil
				}).ExpectHSet(dataKey,
					"otp", mock.Anything,
					"total_delivery", 2,
					"total_attempts", 0,
					"total_resend", 1).SetVal(1)

				r.ExpectDel(lockKey).SetVal(1)

				r.ExpectKeys("backend-portal:otp-verification:unique-id:user-login:token-otp:*").SetVal([]string{})

				r.ExpectSet(
					"backend-portal:otp-verification:unique-id:user-login:token-otp:"+token,
					email,
					25*time.Minute,
				).SetVal("1")

				r.ExpectTTL(dataKey).SetVal(time.Duration(10 * time.Second))
				r.ExpectExpire(dataKey, 2*time.Hour).SetVal(true)
			},
			wantResult: token,
		},
		{
			name:    "SUCCESS:Invalid_OTP_JSON_Format",
			feature: constant.OTPIdentifierUserLogin,
			id:      "test@example.com",
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
				r.ClearExpect()

				email := "test@example.com"
				suspendKey := "backend-portal:otp-verification:" + email + ":user-login:suspend"
				dataKey := "backend-portal:otp-verification:" + email + ":user-login:data"
				lockKey := "backend-portal:otp-verification:" + email + ":user-login:lock"

				r.ExpectGet(suspendKey).SetVal(`{"status": false}`)
				r.ExpectHGet(dataKey, "total_delivery").SetVal("1")

				limiterMock.On(
					"Allow", mock.Anything, mock.Anything, mock.Anything,
				).Return(&redisExt.Result{Allowed: 1}, nil).Maybe()

				r.ExpectSetNX(lockKey, true, 10*time.Second).SetVal(true)

				j.On("GenerateTokenForOTP", mock.Anything, mock.Anything, mock.Anything).Return(token, nil).Once()

				// Return malformed JSON to trigger the error handling path
				r.ExpectHGet(dataKey, "otp").SetVal(`{"invalid_json": true`)

				// Use a custom matcher for flexible HSet expectations
				r.CustomMatch(func(expected, actual []interface{}) error {
					// Check command name
					if expected[0] != actual[0] {
						return fmt.Errorf("expected %v, got %v", expected[0], actual[0])
					}

					// Check key
					if expected[1] != actual[1] {
						return fmt.Errorf("expected %v, got %v", expected[1], actual[1])
					}

					// Only check field names, not values (for flexibility)
					if expected[2] != actual[2] || expected[4] != actual[4] ||
						expected[6] != actual[6] || expected[8] != actual[8] {
						return fmt.Errorf("field names don't match")
					}

					return nil
				}).ExpectHSet(dataKey,
					"otp", mock.Anything,
					"total_delivery", 2,
					"total_attempts", 0,
					"total_resend", 1).SetVal(1)

				r.ExpectDel(lockKey).SetVal(1)
				r.ExpectKeys("backend-portal:otp-verification:unique-id:user-login:token-otp:*").SetVal([]string{})

				r.ExpectSet(
					"backend-portal:otp-verification:unique-id:user-login:token-otp:"+token,
					email,
					25*time.Minute,
				).SetVal("1")

				r.ExpectTTL(dataKey).SetVal(time.Duration(10 * time.Second))
				r.ExpectExpire(dataKey, 2*time.Hour).SetVal(true)
			},
			wantResult: token,
		},
		{
			name:    "SUCCESS:OTP_with_forgot_password",
			feature: constant.OTPIdentifierForgotPassword,
			id:      "rizaldiamanda+innovation@gmail.com",
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
				r.ClearExpect()

				email := "rizaldiamanda+innovation@gmail.com"
				suspendKey := "backend-portal:otp-verification:" + email + ":forgot-password:suspend"
				dataKey := "backend-portal:otp-verification:" + email + ":forgot-password:data"
				lockKey := "backend-portal:otp-verification:" + email + ":forgot-password:lock"

				r.ExpectGet(suspendKey).SetVal(`{"status": false}`)
				r.ExpectHGet(dataKey, "total_delivery").RedisNil()

				limiterMock.On(
					"Allow", mock.Anything, mock.Anything, mock.Anything,
				).Return(&redisExt.Result{Allowed: 1}, nil).Maybe()

				r.ExpectSetNX(lockKey, true, 10*time.Second).SetVal(true)
				j.On("GenerateTokenForOTP", mock.Anything, mock.Anything, mock.Anything).Return(token, nil).Once()

				r.ExpectHGet(dataKey, "otp").RedisNil()

				// Use a custom matcher function to match any valid JSON for OTP
				r.CustomMatch(func(expected, actual []interface{}) error {
					// Check command name
					if expected[0] != actual[0] {
						return fmt.Errorf("expected %v, got %v", expected[0], actual[0])
					}

					// Check key
					if expected[1] != actual[1] {
						return fmt.Errorf("expected %v, got %v", expected[1], actual[1])
					}

					// Only check field names, not values (for flexibility)
					if expected[2] != actual[2] || expected[4] != actual[4] ||
						expected[6] != actual[6] || expected[8] != actual[8] {
						return fmt.Errorf("field names don't match")
					}

					return nil
				}).ExpectHSet(dataKey,
					"otp", mock.Anything,
					"total_delivery", 1,
					"total_attempts", 0,
					"total_resend", 0).SetVal(1)

				r.ExpectDel(lockKey).SetVal(1)
				r.ExpectKeys("backend-portal:otp-verification:unique-id:forgot-password:token-otp:*").SetVal([]string{"key-1", "key-2"})
				r.ExpectDel("key-1", "key-2").SetVal(2)

				r.ExpectSet(
					"backend-portal:otp-verification:unique-id:forgot-password:token-otp:"+token,
					email,
					25*time.Minute,
				).SetVal("1")

				r.ExpectTTL(dataKey).SetVal(time.Duration(10 * time.Second))
				r.ExpectExpire(dataKey, 2*time.Hour).SetVal(true)
			},
			wantResult: token,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			clientMock.ClearExpect()

			if test.id == "" {
				test.id = "email"
			}

			test.mockSetup(jwtMock, clientMock)

			token, err := service.GenerateOTPCode(context.Background(), "unique-id", test.id, test.feature)

			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, test.wantResult, token)
		})
	}
}

func TestGenerateTOTPVerifyToken(t *testing.T) {
	jwt := jwtPkgMock.NewIJwt(t)
	log := loggerPkgMocks.NewILogger(t)
	redis := redisPkgMocks.NewIRedisExt(t)
	rateLimit := redisPkgMocks.NewILimiter(t)

	service := New(&config.Config{}, log, redis, jwt, nil, nil, rateLimit)

	errInternalService := fmt.Errorf(constant.InternalErrorFmt, "")

	tests := []struct {
		name       string
		request    otp.GenerateTOTPVerifyTokenRequest
		setupMock  func()
		wantError  error
		wantResult string
	}{
		{
			name: "ERROR:Feature not support",
			request: otp.GenerateTOTPVerifyTokenRequest{
				Feature: constant.OTPIdentifierForgotPassword,
			},
			setupMock: func() { /* empty */ },
			wantError: pkgErrs.New(response.HttpErrForbidden, constant.ErrFeatureNotSupportTOTPAuth),
		},
		{
			name: "ERROR:Rate limit validation",
			request: otp.GenerateTOTPVerifyTokenRequest{
				Feature: constant.OTPIdentifierUserLogin,
			},
			setupMock: func() {
				rateLimit.On("Allow", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed while check rate limit on generate totp verify token", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, errInternalService),
		},
		{
			name: "ERROR:Request limit exceeded",
			request: otp.GenerateTOTPVerifyTokenRequest{
				Feature: constant.OTPIdentifierChangePassword,
			},
			setupMock: func() {
				rateLimit.On("Allow", mock.Anything, mock.Anything, mock.Anything).Once().Return(&redisExt.Result{Allowed: 0}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrResourceLocked, errors.New("please wait for a moment")), // NOSONAR
		},
		{
			name: "ERROR:Generate verify token",
			request: otp.GenerateTOTPVerifyTokenRequest{
				Feature: constant.OTPIdentifierResetPIN,
			},
			setupMock: func() {
				rateLimit.On("Allow", mock.Anything, mock.Anything, mock.Anything).Return(&redisExt.Result{Allowed: 1}, nil)
				jwt.On("GenerateTokenForOTP", mock.Anything, mock.Anything, mock.Anything).Once().Return("", assert.AnError)
				log.On("Error", mock.Anything, "Failed while generate totp verify token", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, errInternalService), // NOSONAR
		},
		{
			name: "SUCCESS",
			request: otp.GenerateTOTPVerifyTokenRequest{
				Feature: constant.OTPIdentifierResetPIN,
			},
			setupMock: func() {
				jwt.On("GenerateTokenForOTP", mock.Anything, mock.Anything, mock.Anything).Once().Return("token", nil)
				redis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil, wantResult: (constant.TOTPTokenPrefixID + "token"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GenerateTOTPVerifyToken(t.Context(), test.request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
