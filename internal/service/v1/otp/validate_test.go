package otp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/service/v1/otp"
	jwtPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	redisPkgMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	userRepoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidateOTPCode(t *testing.T) {
	// Set up defaults for test cases
	const (
		defEmail     = "user@example.id"
		defSessionID = "user@example.id"
		defCode      = "123456"
		badCode      = "654321"
	)

	// Use OTPIdentifierUserLogin for SUCCESS test case as it might have a higher max attempts value
	defaultIdentifier := constant.OTPIdentifierForgotPassword

	// Get the feature name that will be used in Redis keys
	defFeatureName := constant.OTPForgotPasswordName

	defDataKey := fmt.Sprintf(
		"backend-portal:otp-verification:%s:%s:data",
		defSessionID,
		defFeatureName,
	)
	defLockKey := fmt.Sprintf(
		"backend-portal:otp-verification:%s:%s:lock",
		defSessionID,
		defFeatureName,
	)

	defData := &otpModel.VerifyOTP{
		ID:                  defEmail,
		Email:               defEmail,
		OTPCode:             defCode,
		Identifier:          defaultIdentifier,
		TwoFactorAuthMethod: constant.TwoFactorAuthMethodOTP,
	}

	client, clientMock := redismock.NewClientMock()
	jwtMock := jwtPkgMock.NewIJwt(t)
	limiterMock := redisPkgMocks.NewILimiter(t)
	userRepoMock := userRepoMocks.NewIUserRepository(t)

	service := otp.New(
		cfg,
		pdkLoggerMock,
		redisExt.WrapRedisClient(client, nil),
		jwtMock,
		nil,
		userRepoMock,
		limiterMock,
	)

	testCases := []struct {
		name      string
		data      *otpModel.VerifyOTP
		mockSetup func(*jwtPkgMock.IJwt, redismock.ClientMock, *redisPkgMocks.ILimiter, *userRepoMocks.IUserRepository)
		wantErr   string
		wantToken string
	}{
		{
			name: "ERROR:Getting OTP data/Invalid session",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetErr(errors.New("invalid db session"))
			},
			wantErr: "ERROR_DATABASE | invalid db session",
		},
		{
			name: "ERROR:Getting OTP data/Data not found",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetErr(redisExt.ErrNil)
			},
			wantErr: "ERROR_REQUEST | OTP data is not registered",
		},
		{
			name: "ERROR:Max attempts has been exceeded",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetVal(`[
					{"otp":"123456","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("3")
			},
			wantErr: "ERROR_TOO_MANY_REQUEST | max attempts limit has been exceeded",
		},
		{
			name: "ERROR:[TOTP]Find user totp data by id",
			data: &otpModel.VerifyOTP{
				Identifier:          constant.OTPIdentifierUserLogin,
				TwoFactorAuthMethod: constant.TwoFactorAuthMethodTOTP,
			},
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, userRepo *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(true)
				userRepo.On("FindUserTOTPDataByID", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: "ERROR_DATABASE | an error occurred on the server. please try again later",
		},
		{
			name: "ERROR:[TOTP]User not found",
			data: &otpModel.VerifyOTP{
				Identifier:          constant.OTPIdentifierUserLogin,
				TwoFactorAuthMethod: constant.TwoFactorAuthMethodTOTP,
			},
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, userRepo *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(true)
				userRepo.On("FindUserTOTPDataByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantErr: "ERROR_NOT_FOUND | user not found",
		},
		{
			name: "ERROR:[TOTP]User not enrolled",
			data: &otpModel.VerifyOTP{
				Identifier:          constant.OTPIdentifierUserLogin,
				TwoFactorAuthMethod: constant.TwoFactorAuthMethodTOTP,
			},
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, userRepo *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(true)
				userRepo.On("FindUserTOTPDataByID", mock.Anything, mock.Anything).Once().Return(&user.UserTOTPData{}, nil)
			},
			wantErr: "ERROR_INTERNAL | an error occurred on the server. please try again later",
		},
		{
			name: "ERROR:OTP has expired",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetVal(`[
					{"otp":"123456","expired_at":"2020-01-01T00:00:00Z","verify":false}
				]`)
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(true)
				r.ExpectDel(defLockKey).SetVal(1)
				r.ExpectHSet(defDataKey, "total_attempts", 1).SetVal(1)
			},
			wantErr: "ERROR_UNPROCESSABLE_CONTENT | the code is expired",
		},
		{
			name: "ERROR:Exclusive lock/Invalid db session",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetVal(`[
					{"otp":"123456","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetErr(errors.New("invalid db session"))
			},
			wantErr: "ERROR_DATABASE | invalid db session",
		},
		{
			name: "ERROR:Exclusive lock/Same process is underway",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetVal(`[
					{"otp":"123456","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(false)
			},
			wantErr: "ERROR_REQUEST | the same request is in progress",
		},
		{
			name: "ERROR:Incorrect password #1",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetVal(`[
					{"otp":"654321","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(true)
				r.ExpectDel(defLockKey).SetVal(1)
				r.ExpectHSet(defDataKey, "total_attempts", 1).SetVal(1)
			},
			wantErr: "ERROR_UNPROCESSABLE_CONTENT | wrong code",
		},
		{
			name: "ERROR:Incorrect password #2",
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetVal(`[
					{"otp":"654321","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("2")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(true)
				r.ExpectDel(defLockKey).SetVal(1)
				r.ExpectHSet(defDataKey, "total_attempts", 3).SetVal(1)
			},
			wantErr: "ERROR_UNPROCESSABLE_CONTENT | wrong code",
		},
		{
			name: "ERROR:Wrong code/User blocked/Invalid session",
			data: &otpModel.VerifyOTP{
				ID:                  "unique-id",
				Email:               "email@example.id",
				OTPCode:             "123456",
				Identifier:          constant.OTPIdentifierUserLogin,
				TwoFactorAuthMethod: constant.TwoFactorAuthMethodOTP,
			},
			mockSetup: func(_ *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, userRepoMock *userRepoMocks.IUserRepository) {
				loginDataKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:data", "email@example.id", constant.OTPUserLoginName)
				loginLockKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:lock", "email@example.id", constant.OTPUserLoginName)

				r.ExpectHGet(loginDataKey, "otp").SetVal(`[
					{"otp":"654321","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(loginDataKey, "total_attempts").SetVal("2")
				r.ExpectSetNX(loginLockKey, "lock", 30*time.Second).SetVal(true)
				userRepoMock.On("BlockUser", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.Anything).Return(errors.New("invalid session"))
				r.ExpectDel(loginLockKey).SetVal(1)
				r.ExpectHSet(loginDataKey, "total_attempts", 3).SetVal(1)
			},
			wantErr: "ERROR_DATABASE | invalid session",
		},
		{
			name: "ERROR:Wrong code/User blocked/Success",
			data: &otpModel.VerifyOTP{
				ID:                  "unique-id",
				Email:               "email@example.id",
				OTPCode:             "123456",
				Identifier:          constant.OTPIdentifierUserLogin,
				TwoFactorAuthMethod: constant.TwoFactorAuthMethodOTP,
			},
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, userRepoMock *userRepoMocks.IUserRepository) {
				loginDataKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:data", "email@example.id", constant.OTPUserLoginName)
				loginLockKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:lock", "email@example.id", constant.OTPUserLoginName)

				r.ExpectHGet(loginDataKey, "otp").SetVal(`[
					{"otp":"654321","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(loginDataKey, "total_attempts").SetVal("2")
				r.ExpectSetNX(loginLockKey, "lock", 30*time.Second).SetVal(true)
				userRepoMock.On("BlockUser", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.Anything).Return(nil)
				j.On("RemoveIterateTokenFromRedis", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).Return(nil)
				r.ExpectDel(loginLockKey).SetVal(1)
				r.ExpectHSet(loginDataKey, "total_attempts", 3).SetVal(1)
			},
			wantErr: "ERROR_TOO_MANY_REQUEST | user is blocked",
		},
		{
			name: "ERROR:Generate token feature 2FA",
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock, _ *redisPkgMocks.ILimiter, _ *userRepoMocks.IUserRepository) {
				r.ExpectHGet(defDataKey, "otp").SetVal(`[
					{"otp":"123456","expired_at":"2999-12-31T23:59:59Z","verify":false},
					{"otp":"654321","expired_at":"2999-12-31T23:59:59Z","verify":false}
				]`)
				r.ExpectHGet(defDataKey, "total_attempts").SetVal("0")
				r.ExpectSetNX(defLockKey, "lock", 30*time.Second).SetVal(true)
				j.On(
					"GenerateTokenForFeature2FA",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					defSessionID,
					defFeatureName,
				).Return("", errors.New("incorrect data format"))
				r.ExpectDel(defLockKey).SetVal(1)
				r.ExpectHSet(defDataKey, "total_attempts", 1).SetVal(1)
				r.ExpectHSet(defDataKey, "total_resend_otp", 0).SetVal(1)
				r.ExpectHSet(defDataKey, "otp", mock.Anything).SetVal(1)
			},
			wantErr: "ERROR_INTERNAL | incorrect data format",
		},
		{
			name: "SUCCESS",
			data: &otpModel.VerifyOTP{
				ID:                  defEmail,
				Email:               defEmail,
				OTPCode:             constant.OTPHardCode,
				Identifier:          constant.OTPIdentifierUserLogin,
				TwoFactorAuthMethod: constant.TwoFactorAuthMethodOTP,
			},
			mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock, l *redisPkgMocks.ILimiter, u *userRepoMocks.IUserRepository) {
				r.ClearExpect()

				// Basic setup
				featureName := constant.OTPUserLoginName
				dataKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:data", defEmail, featureName)
				lockKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:lock", defEmail, featureName)
				tokenKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:feature-2fa:jwt-token", defEmail, featureName)
				suspendKey := fmt.Sprintf("backend-portal:otp-verification:%s:%s:suspend", defEmail, featureName)

				// Create valid OTP JSON with proper expiration (future date)
				futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
				otpJson := fmt.Sprintf(`[{"otp":"%s","expired_at":"%s","verify":false}]`, constant.OTPHardCode, futureTime)

				// 1. Initial data checks
				r.ExpectHGet(dataKey, "otp").SetVal(otpJson)
				r.ExpectHGet(dataKey, "total_attempts").SetVal("0")

				// 2. Lock handling - updated to use SetNX instead of Exists + Set
				r.ExpectSetNX(lockKey, "lock", 30*time.Second).SetVal(true)

				// 3. Session validation
				r.ExpectHGetAll(dataKey).SetVal(map[string]string{
					"email":          defEmail,
					"otp":            otpJson,
					"total_attempts": "0",
					"session_id":     defEmail,
					"total_delivery": "1",
					"total_resend":   "0",
				})

				// 4. OTP Verification - Update OTP to mark as verified
				updatedOtpJson := fmt.Sprintf(`[{"otp":"%s","expired_at":"%s","verify":true}]`, constant.OTPHardCode, futureTime)
				r.ExpectHSet(dataKey, "otp", updatedOtpJson).SetVal(1)

				// 5. Reset resend counter
				r.ExpectHSet(dataKey, "total_resend", "0").SetVal(1)

				// 6. JWT Token generation
				j.On(
					"GenerateTokenForFeature2FA",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					defEmail,
					constant.OTPIdentifierUserLogin,
				).Return(token, nil).Once()

				// 7. Set token in Redis
				r.ExpectSet(tokenKey, defEmail, 30*time.Minute).SetVal("OK")

				// 8. Reset limiter
				l.On(
					"Reset",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					fmt.Sprintf("backend-portal:otp-verification:%s:%s", defEmail, featureName),
				).Return(nil).Once()

				// 9. Delete suspend key if exists
				r.ExpectDel(suspendKey).SetVal(0)

				// 10. Update attempts counter (in deferred function)
				r.ExpectHSet(dataKey, "total_attempts", 1).SetVal(1)

				// 11. Delete lock (in deferred function)
				r.ExpectDel(lockKey).SetVal(1)
			},
			wantToken: token,
		},
	}

	errStringContains := map[string]string{
		"ERROR:Expired_OTP":                             "ERROR_UNPROCESSABLE_CONTENT | the code is expired",
		"ERROR:Invalid_code_length":                     "ERROR_UNPROCESSABLE_CONTENT | invalid code length",
		"ERROR:Incorrect_password_#1":                   "ERROR_UNPROCESSABLE_CONTENT | wrong code",
		"ERROR:Incorrect_password_#2":                   "ERROR_UNPROCESSABLE_CONTENT | wrong code",
		"ERROR:Wrong_code/User_blocked/Invalid_session": "ERROR_DATABASE | invalid session",
		"ERROR:Wrong_code/User_blocked/Success":         "ERROR_TOO_MANY_REQUEST | max attempts limit has been exceeded",
		"ERROR:Generate_token_feature_2FA":              "ERROR_INTERNAL | incorrect data format",
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// Reset mocks before each test case
			clientMock.ClearExpect()
			clientMock.MatchExpectationsInOrder(true)
			jwtMock.ExpectedCalls = nil
			limiterMock.ExpectedCalls = nil
			userRepoMock.ExpectedCalls = nil

			test.mockSetup(jwtMock, clientMock, limiterMock, userRepoMock)

			if test.data == nil {
				test.data = defData
			}

			token, err := service.ValidateOTPCode(context.Background(), test.data)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), errStringContains[test.name])
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantToken, token)
		})
	}
}
