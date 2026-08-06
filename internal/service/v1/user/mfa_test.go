package user

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindUserTOTPDataByID(t *testing.T) {
	log := loggerMock.NewILogger(t)
	userRepo := repoMocks.NewIUserRepository(t)

	service := New(nil, nil, log, userRepo, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *model.UserTOTPData
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed while find user TOTP data by id", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, constant.ErrUserNotFound),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, mock.Anything).Once().Return(&model.UserTOTPData{}, nil)
			},
			wantError:  nil,
			wantResult: &model.UserTOTPData{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.FindUserTOTPDataByID(t.Context(), "")
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			log.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestEnrollTOTP(t *testing.T) {

	log := loggerMock.NewILogger(t)
	userRepo := repoMocks.NewIUserRepository(t)
	encryptionKey := vaultMock.NewIVaultKeyValue(t)

	rdb, clientMock := redismock.NewClientMock()
	defer rdb.Close()

	clientMock.MatchExpectationsInOrder(false)

	config := &config.Config{
		MultiFactorAuth: config.MultiFactorAuthConfig{
			TimeBasedOTP: config.TimeBasedOTPConfig{
				TOTPIssuer:     "PivotPayment", // NOSONAR
				TOTPSecretSize: 32,             // NOSONAR
			},
		},
	}
	service := New(config, nil, log, userRepo, nil, WithEncryptionKey(encryptionKey), WithRedisClient(redisExt.WrapRedisClient(rdb, nil)))

	userID := "bfb430cc-cc71-429f-b9ef-e23bf6773ee5"
	userData := &model.UserTOTPData{
		UserId:     userID,
		Email:      "john.doe@example.com", // NOSONAR
		TOTPStatus: constant.TOTPStatusActive,
	}
	errInternalService := fmt.Errorf(constant.InternalErrorFmt, "")
	enrollmentCacheKey := fmt.Sprintf(constant.TOTPEnrollmentCacheKeyFmt, userID)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult func(t *testing.T, result *model.EnrollTOTPResponse)
	}{
		{
			name: "ERROR:User not found",
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, userID).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, constant.ErrUserNotFound),
		},
		{
			name: "ERROR:Get secret key",
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, userID).Return(userData, nil)
				encryptionKey.On("GetSecretKeyString", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed while getting user encryption key", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, errInternalService),
		},
		{
			name: "ERROR:Secret value is not base64 format",
			setupMock: func() {
				encryptionKey.On("GetSecretKeyString", mock.Anything, mock.Anything).Once().Return(&vault.SecretKey[string]{Value: "123456"}, nil)
				log.On("Error", mock.Anything, "Failed while base64 decode user encryption key", logger.Error(base64.CorruptInputError(4))).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, errInternalService),
		},
		{
			name: "ERROR:Update user TOTP data",
			setupMock: func() {
				encryptionKey.On("GetSecretKeyString", mock.Anything, mock.Anything).Return(&vault.SecretKey[string]{Version: 1, Value: "U9JqU15dTbHFDhnvTYCYmvcbrFcm1oHvhq/WXW6cPcI="}, nil) // NOSONAR
				clientMock.CustomMatch(func(expected, actual []any) error {
					if expected[1] != actual[1] || expected[4] != actual[4] {
						return errors.New("key or duration does not match") // NOSONAR
					}
					return nil
				}).ExpectSet(enrollmentCacheKey, "any", constant.TOTPEnrollmentCacheDuration).SetErr(assert.AnError)
				log.On("Error", mock.Anything, "Failed to store TOTP enrollment in cache", mock.Anything).Once().Return()

			},
			wantError: pkgErrs.New(response.HttpErrDatabase, errInternalService),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				clientMock.CustomMatch(func(expected, actual []any) error {
					if expected[1] != actual[1] || expected[4] != actual[4] {
						return errors.New("key or duration does not match") // NOSONAR
					}
					return nil
				}).ExpectSet(enrollmentCacheKey, "any", constant.TOTPEnrollmentCacheDuration).SetVal("")
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *model.EnrollTOTPResponse) {
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.QRCodeDataURL)
				assert.NotEmpty(t, result.SecretKey)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientMock.ClearExpect()

			test.setupMock()

			result, err := service.EnrollTOTP(t.Context(), model.EnrollTOTPRequest{
				UserId: userID,
			})
			assert.Equal(t, test.wantError, err)
			if test.wantResult == nil {
				assert.Nil(t, result)
			} else {
				test.wantResult(t, result)
			}

			log.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			encryptionKey.AssertExpectations(t)
			require.NoError(t, clientMock.ExpectationsWereMet())
		})
	}
}

func TestConfirmTOTP(t *testing.T) {
	log := loggerMock.NewILogger(t)
	otpSvc := serviceMocks.NewIOTP(t)
	userRepo := repoMocks.NewIUserRepository(t)

	rdb, clientMock := redismock.NewClientMock()
	defer rdb.Close()

	clientMock.MatchExpectationsInOrder(false)

	service := New(nil, nil, log, userRepo, nil, WithOTPService(otpSvc), WithRedisClient(redisExt.WrapRedisClient(rdb, nil)))

	userId := "2f964647-5a70-43a8-8ff0-c1fafdcdcb4c"
	enrollmentKey := fmt.Sprintf(constant.TOTPEnrollmentCacheKeyFmt, userId)
	enrollmentDataStr := `{"wrappedSecretKey":"","encryptVersion":1}` // NOSONAR

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult bool
	}{
		{
			name: "ERROR:Get enrollment data",
			setupMock: func() {
				clientMock.ExpectGet(enrollmentKey).SetErr(assert.AnError)
				log.On("Error", mock.Anything, "Failed to get TOTP enrollment data in cache", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name: "ERROR:Enrollment data not found",
			setupMock: func() {
				clientMock.ExpectGet(enrollmentKey).SetErr(redisExt.ErrNil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, errors.New("TOTP enrollment has not been completed or has expired")), // NOSONAR
		},
		{
			name: "SUCCESS:Invalid TOTP code",
			setupMock: func() {
				clientMock.ExpectGet(enrollmentKey).SetVal(enrollmentDataStr)
				otpSvc.On("ValidateTOTPCode", mock.Anything, mock.Anything).Once().Return(false, nil)
			},
			wantError: nil, wantResult: false,
		},
		{
			name: "ERROR:Update user TOTP data",
			setupMock: func() {
				clientMock.ExpectGet(enrollmentKey).SetVal(enrollmentDataStr)
				otpSvc.On("ValidateTOTPCode", mock.Anything, mock.Anything).Return(true, nil)
				userRepo.On("UpdateUserTOTPData", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed while update user TOTP data on confirm enrollment", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name: "SUCCESS:Valid OTP code",
			setupMock: func() {
				clientMock.ExpectGet(enrollmentKey).SetVal(enrollmentDataStr)
				userRepo.On("UpdateUserTOTPData", mock.Anything, mock.MatchedBy(func(params *model.UpdateUserTOTPDataRequest) bool {
					return params.UserId == userId &&
						params.EncryptVersion == 1 &&
						params.Status == constant.TOTPStatusActive
				})).Once().Return(nil)
				clientMock.ExpectDel(enrollmentKey)
			},
			wantError: nil, wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientMock.ClearExpect()

			test.setupMock()

			result, err := service.ConfirmTOTP(t.Context(), model.ConfirmTOTPRequest{
				UserId: userId,
			})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			log.AssertExpectations(t)
			otpSvc.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			require.NoError(t, clientMock.ExpectationsWereMet())
		})
	}
}

func TestSetPreferred2FAMethod(t *testing.T) {
	log := loggerMock.NewILogger(t)
	userRepo := repoMocks.NewIUserRepository(t)

	service := New(nil, nil, log, userRepo, nil)

	userId := "bfb430cc-cc71-429f-b9ef-e23bf6773ee5"

	tests := []struct {
		name       string
		request    model.SetPreferred2FAMethodRequest
		setupMock  func()
		wantError  error
		wantResult *model.SetPreferred2FAMethodResponse
	}{
		{
			name: "ERROR: Invalid 2FA method",
			request: model.SetPreferred2FAMethodRequest{
				UserId:             userId,
				Preferred2FAMethod: "INVALID_METHOD",
			},
			setupMock: func() {},
			wantError: pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidPreferred2FAMethod),
		},
		{
			name: "ERROR: TOTP selected but user not found",
			request: model.SetPreferred2FAMethodRequest{
				UserId:             userId,
				Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
			},
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, userId).Once().Return(nil, nil)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, constant.ErrUserNotFound),
		},
		{
			name: "ERROR: TOTP selected but not active",
			request: model.SetPreferred2FAMethodRequest{
				UserId:             userId,
				Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
			},
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, userId).Once().Return(&model.UserTOTPData{
					UserId:     userId,
					TOTPStatus: constant.TOTPStatusEnrolled,
				}, nil)
			},
			wantError: pkgErrs.New(response.HttpErrRequest, constant.ErrTOTPRequiredButNotActive),
		},
		{
			name: "ERROR: Database error when updating",
			request: model.SetPreferred2FAMethodRequest{
				UserId:             userId,
				Preferred2FAMethod: string(constant.TwoFactorAuthMethodOTP),
			},
			setupMock: func() {
				userRepo.On("UpdateUserPreferred2FAMethod", mock.Anything, userId, string(constant.TwoFactorAuthMethodOTP)).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed while updating user preferred 2FA method", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, "")),
		},
		{
			name: "SUCCESS: Set OTP as preferred method",
			request: model.SetPreferred2FAMethodRequest{
				UserId:             userId,
				Preferred2FAMethod: string(constant.TwoFactorAuthMethodOTP),
			},
			setupMock: func() {
				userRepo.On("UpdateUserPreferred2FAMethod", mock.Anything, userId, string(constant.TwoFactorAuthMethodOTP)).Once().Return(nil)
			},
			wantError: nil,
			wantResult: &model.SetPreferred2FAMethodResponse{
				Preferred2FAMethod: string(constant.TwoFactorAuthMethodOTP),
				Updated:            true,
			},
		},
		{
			name: "SUCCESS: Set TOTP as preferred method",
			request: model.SetPreferred2FAMethodRequest{
				UserId:             userId,
				Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
			},
			setupMock: func() {
				userRepo.On("FindUserTOTPDataByID", mock.Anything, userId).Once().Return(&model.UserTOTPData{
					UserId:     userId,
					TOTPStatus: constant.TOTPStatusActive,
				}, nil)
				userRepo.On("UpdateUserPreferred2FAMethod", mock.Anything, userId, string(constant.TwoFactorAuthMethodTOTP)).Once().Return(nil)
			},
			wantError: nil,
			wantResult: &model.SetPreferred2FAMethodResponse{
				Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
				Updated:            true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.SetPreferred2FAMethod(t.Context(), test.request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			log.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}
