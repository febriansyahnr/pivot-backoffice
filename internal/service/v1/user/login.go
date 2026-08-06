package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	commServicePb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/commService"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	userLoggedInDeviceModel "github.com/paper-indonesia/pivot-backoffice/internal/model/userLoggedInDevice"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *UserService) Login(ctx context.Context, request *userModel.UserLoginRequest) (user *userModel.User, signedToken string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/Login")
	defer segment.End()

	if user, err = s.userRepo.FindUserByEmail(ctx, request.Email); err != nil {
		s.logger.Error(ctx, "failed to find user by email", logger.Error(err))
		return nil, "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if user == nil {
		s.logger.Info(ctx, "Trying to login using an unregistered email", logger.String("email", request.Email))
		return nil, "", pkgErrs.New(response.HttpErrNotFound, constant.ErrInvalidEmailOrPassword)
	}

	attempts := s.redis.CustomIncr(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", request.Email), 24*time.Hour)

	if user.DeactivatedAt.Valid || user.Status == constant.UserStatusInactive {
		s.logFailedLoginActivity(ctx, user)
		return nil, "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user is not activated"))

	} else if !user.Blocked.Time.IsZero() || user.Status == constant.UserStatusBlocked {
		s.logFailedLoginActivity(ctx, user)
		return nil, "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user has been blocked"))

	} else if user.Status == constant.UserStatusInvited {
		s.logFailedLoginActivity(ctx, user)
		return nil, "", pkgErrs.New(response.HttpStatusErrorConflict, constant.ErrUserInvitedStatus)
	}

	if err = s.validateMerchantStatus(ctx, user); err != nil {
		return nil, "", err
	}

	// check if login attempt is more than 3 times
	if intAttempt, _ := strconv.Atoi(attempts.Val()); intAttempt >= 3 {
		// block user
		user.Blocked = sql.NullTime{Time: time.Now().UTC().Add(constant.BLOCKED_DURATION), Valid: true}
		errUpdate := s.userRepo.Update(ctx, user)
		if errUpdate != nil {
			s.logger.Error(ctx, "failed to update user", logger.Error(errUpdate))
			return nil, "", pkgErrs.New(response.HttpErrDatabase, errUpdate)
		}

		// delete login attempt from redis
		s.redis.Del(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", request.Email))
		s.logFailedLoginActivity(ctx, user)
		return user, "", pkgErrs.New(response.HttpErrUnauthorized, constant.ErrBlockedTooManyAttempts)
	}

	if !user.ComparePassword(request.Password) {
		s.logFailedLoginActivity(ctx, user)
		return nil, "", pkgErrs.New(response.HttpErrUnauthorized, constant.ErrInvalidEmailOrPassword)
	}

	// Legacy deviceID from generator
	deviceID := util.GenerateDeviceID(ctx.Value(constant.CtxUserAgentKey).(string))

	// Check if this is account that can bypass OTP
	bypassIsTrue, err := constant.IsAccountBypassOTP(ctx, config.Environment(), user.UUID, user.Email)
	if err != nil {
		s.logger.Error(ctx, "Failed to check account for OTP bypass", logger.Error(err))
	}

	// Validate device identifier
	if deviceIdentifierFromHeader, ok := ctx.Value(constant.CtxUserDeviceIdentifierKey).(string); ok {
		if err = s.userLoggedInDeviceSvc.Validate(ctx, user.UUID, deviceIdentifierFromHeader, request.IsRemember); err != nil {
			// Skip device validation error if this is a account with flag OTP bypass
			if !bypassIsTrue {
				return nil, "", err
			}
			s.logger.Info(ctx, "Bypass account detected, bypassing device validation for login", logger.String("email", user.Email))
		}

		if deviceIdentifierFromHeader != "" {
			deviceID = deviceIdentifierFromHeader
		}
	}
	user.DeviceIdentifier = deviceID

	// generate access token
	accessToken, errSign := s.JWT.GenerateAccessToken(ctx, user)
	if errSign != nil {
		s.logger.Error(ctx, "failed to generate access token", logger.Error(errSign))
		return nil, "", pkgErrs.New(response.HttpErrInternal, errSign)
	}

	// generate refresh token
	refreshToken, err := s.JWT.GenerateRefreshToken(ctx, user, time.Now().UTC().Add(constant.REFRESH_EXPIRATION_DURATION))
	if err != nil {
		s.logger.Error(ctx, "failed to generate refresh token", logger.Error(err))
		return nil, "", pkgErrs.New(response.HttpErrInternal, err)
	}
	user.RefreshToken.Valid = true
	user.RefreshToken.String = refreshToken

	// update user refresh token
	errUpdate := s.userRepo.UpdateRefreshToken(ctx, user.UUID, user.RefreshToken.String)
	if errUpdate != nil {
		s.logger.Error(ctx, "failed to update user refresh token", logger.Error(errUpdate))
		return nil, "", pkgErrs.New(response.HttpErrDatabase, errUpdate)
	}

	user.LastLoginAt = commonModel.CustomNullTime{NullTime: sql.NullTime{Time: time.Now().UTC(), Valid: true}}
	errUpdateUser := s.userRepo.Update(ctx, user)
	if errUpdateUser != nil {
		s.logger.Error(ctx, "failed to update user", logger.Error(errUpdate))
		return nil, "", pkgErrs.New(response.HttpErrDatabase, errUpdate)
	}

	// insert jwt token to redis and set ttl same as jwt token expiration
	s.redis.Set(
		ctx,
		fmt.Sprintf("backend-portal:access-token:%s:%s", user.Email, deviceID),
		accessToken, time.Now().UTC().Add(constant.LOGIN_EXPIRATION_DURATION).Sub(time.Now().UTC()))

	_ = s.JWT.TerminateTokenOtherDevices(ctx, user.Email, deviceID)

	// remove login attempt from redis
	s.redis.Del(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", request.Email))

	return user, accessToken, nil
}

func (s *UserService) LoginWithOTP(ctx context.Context, email, password string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/LoginWithOTP")
	defer segment.End()

	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "failed to find user by email", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if user == nil {
		s.logger.Info(ctx, "Trying to login using an unregistered email", logger.String("email", email))
		return "", pkgErrs.New(response.HttpErrUnauthorized, constant.ErrInvalidEmailOrPassword)
	}

	attempts := s.redis.CustomIncr(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", email), 24*time.Hour)

	if user.DeactivatedAt.Valid {
		s.logFailedLoginActivity(ctx, user)
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user is deactivated"))

	} else if user.Status == constant.UserStatusInvited {
		s.logFailedLoginActivity(ctx, user)
		return "", pkgErrs.New(response.HttpStatusErrorConflict, constant.ErrUserInvitedStatus)

	} else if user.Status != constant.UserStatusActive {
		s.logFailedLoginActivity(ctx, user)
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user is not activated"))

	} else if !user.Blocked.Time.IsZero() {
		s.logFailedLoginActivity(ctx, user)
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user has been blocked"))
	}

	// Check if login attempt is more than 3 times, if true then block user.
	if intAttempt, _ := strconv.Atoi(attempts.Val()); intAttempt >= 3 {

		user.Blocked = sql.NullTime{Time: time.Now().UTC().Add(constant.BLOCKED_DURATION), Valid: true}

		if err := s.userRepo.Update(ctx, user); err != nil {
			s.logger.Error(ctx, "failed to update user", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrDatabase, err)
		}

		_ = s.redis.Del(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", email))

		// Revoke all user's token
		_ = s.JWT.RemoveIterateTokenFromRedis(ctx, user.Email)
		s.logFailedLoginActivity(ctx, user)
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user is blocked, too many login attempts"))
	}

	if !user.ComparePassword(password) {
		s.logFailedLoginActivity(ctx, user)
		return "", pkgErrs.New(response.HttpErrUnauthorized, constant.ErrInvalidEmailOrPassword)
	}
	ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, user.MerchantId)

	// Remove login attempt from redis
	_ = s.redis.Del(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", email))

	// Check user's preferred 2FA method
	preferred2FA := user.Preferred2FAMethod

	// If preferred method is explicitly set to TOTP
	if preferred2FA == string(constant.TwoFactorAuthMethodTOTP) {
		if user.TOTPStatus != constant.TOTPStatusActive {
			return "", pkgErrs.New(response.HttpErrRequest, constant.ErrTOTPRequiredButNotActive)
		}
		verifyTokenRequest := otp.GenerateTOTPVerifyTokenRequest{
			UserId:    user.UUID,
			UserEmail: user.Email,
			Feature:   constant.OTPIdentifierUserLogin,
		}
		return s.otpSvc.GenerateTOTPVerifyToken(ctx, verifyTokenRequest)
	}

	// If preferred method is explicitly set to OTP
	if preferred2FA == string(constant.TwoFactorAuthMethodOTP) {
		return s.otpSvc.GenerateOTPCode(ctx, user.UUID, user.Email, constant.OTPIdentifierUserLogin)
	}

	// Default behavior (AUTO or empty): Use TOTP if active, otherwise OTP
	if user.TOTPStatus == constant.TOTPStatusActive {
		verifyTokenRequest := otp.GenerateTOTPVerifyTokenRequest{
			UserId:    user.UUID,
			UserEmail: user.Email,
			Feature:   constant.OTPIdentifierUserLogin,
		}
		return s.otpSvc.GenerateTOTPVerifyToken(ctx, verifyTokenRequest)
	}
	return s.otpSvc.GenerateOTPCode(ctx, user.UUID, user.Email, constant.OTPIdentifierUserLogin)
}

func (s *UserService) GenerateTokenFromLogin2FA(ctx context.Context, id string) (user *userModel.User, signedToken string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/GenerateTokenFromLogin2FA")
	defer segment.End()

	if user, err = s.userRepo.FindUserByID(ctx, id); err != nil {
		s.logger.Error(ctx, "failed to find user by id", logger.Error(err))
		return nil, "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if user == nil {
		return nil, "", pkgErrs.New(response.HttpErrUnauthorized, constant.ErrInvalidEmailOrPassword)
	}

	// Legacy deviceID from generator
	deviceID := util.GenerateDeviceID(ctx.Value(constant.CtxUserAgentKey).(string))

	// Validate device identifier
	if deviceIdentifierFromHeader, ok := ctx.Value(constant.CtxUserDeviceIdentifierKey).(string); ok {
		if deviceIdentifierFromHeader != "" {
			deviceID = deviceIdentifierFromHeader
		}
	}
	user.DeviceIdentifier = deviceID

	// Update isRemember device
	if isRememberStr, ok := ctx.Value(constant.CtxIsRemember).(string); ok {
		isRemember := false
		if isRememberBool, err := strconv.ParseBool(isRememberStr); err == nil {
			isRemember = isRememberBool
		}

		metadata := &userLoggedInDeviceModel.UserLoggedInDeviceMetadata{
			IsRemember:    isRemember,
			RememberUntil: time.Now().UTC().Add(time.Duration(s.config.UserOTPConfig.UserLoginRememberInMinute) * time.Minute),
		}
		metadataInJson, _ := json.Marshal(metadata)

		if errSet := s.userLoggedInDeviceRepo.SetRememberDevice(ctx, user.UUID, user.DeviceIdentifier, string(metadataInJson)); errSet != nil {
			return nil, "", pkgErrs.New(response.HttpErrDatabase, errSet)
		}
	}

	accessToken, err := s.JWT.GenerateAccessToken(ctx, user)
	if err != nil {
		s.logger.Error(ctx, "failed to generate access token", logger.Error(err))
		return nil, "", pkgErrs.New(response.HttpErrInternal, err)
	}

	user.LastLoginAt = commonModel.CustomNullTime{NullTime: sql.NullTime{Time: time.Now().UTC(), Valid: true}}
	user.RefreshToken.Valid = true
	user.RefreshToken.String, err = s.JWT.GenerateRefreshToken(ctx, user, time.Now().UTC().Add(constant.REFRESH_EXPIRATION_DURATION))
	if err != nil {
		s.logger.Error(ctx, "failed to generate refresh token", logger.Error(err))
		return nil, "", pkgErrs.New(response.HttpErrInternal, err)
	}

	if err = s.userRepo.UpdateRefreshToken(ctx, user.UUID, user.RefreshToken.String); err != nil {
		s.logger.Error(ctx, "failed to update user refresh token", logger.Error(err))
		return nil, "", pkgErrs.New(response.HttpErrDatabase, err)
	}

	if err = s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error(ctx, "failed to update user", logger.Error(err))
		return nil, "", pkgErrs.New(response.HttpErrDatabase, err)
	}

	// insert jwt token to redis and set ttl same as jwt token expiration
	s.redis.Set(
		ctx,
		fmt.Sprintf("backend-portal:access-token:%s:%s", user.Email, deviceID),
		accessToken, time.Now().UTC().Add(constant.LOGIN_EXPIRATION_DURATION).Sub(time.Now().UTC()),
	)

	_ = s.JWT.TerminateTokenOtherDevices(ctx, user.Email, deviceID)

	// Publish activity, do nothing on error
	_ = s.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagAccount,
		constant.ActivityUserLogin,
		map[string]string{
			"deviceIdentifier": deviceID,
			"email":            user.Email,
			"password":         "********",
		},
	)

	var ipAddress string
	locationJSON, ok := ctx.Value(constant.CtxUserIPKey).(string)
	if ok && locationJSON != "" {
		var locationData map[string]interface{}
		if err := json.Unmarshal([]byte(locationJSON), &locationData); err == nil {
			if ip, exists := locationData["ip"]; exists {
				ipAddress, _ = ip.(string)
			}
		}
	}

	formattedDateTime := util.ConvertToJakartaString(time.Now().UTC())

	content, _ := structpb.NewStruct(map[string]any{
		"Location": fmt.Sprintf("IP Address %s", ipAddress),
		"Device":   ctx.Value(constant.CtxUserAgentKey),
		"DateTime": formattedDateTime,
		"LogoURL":  s.config.MerchantPortalConfig.LogoURL,
	})

	emailRequest := &commServicePb.EmailRequest{
		Event:                constant.UserLoginActivityEvent,
		From:                 config.DefaultEmailSender(),
		To:                   user.Email,
		Subject:              "Login Information",
		Content:              content,
		Priority:             commServicePb.EmailPriority_L0,
		ToBeRetriedOnFailure: true,
	}
	payload, _ := proto.Marshal(emailRequest)

	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.CommServiceEmailRoutingKey, nil, payload)
	return user, accessToken, nil
}

func (s *UserService) logFailedLoginActivity(ctx context.Context, user *userModel.User) {
	s.ActivityLog(ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagAccount,
		constant.ActivityUserFailedLogin,
		map[string]string{
			"email":    user.Email,
			"password": "********",
		})
}

func (s *UserService) validateMerchantStatus(ctx context.Context, user *userModel.User) error {
	if user.MerchantStatus.String == constant.MerchantStatusClosed {
		s.logFailedLoginActivity(ctx, user)
		return pkgErrs.New(response.HttpErrUnauthorized, errors.New("Merchant status is closed. Reason: "+user.ReasonStatus.String))
	}

	if user.MerchantStatus.String == constant.MerchantStatusBlocked {
		s.logFailedLoginActivity(ctx, user)
		return pkgErrs.New(response.HttpErrUnauthorized, errors.New("merchant is blocked"))
	}

	if user.MerchantStatus.String == constant.MerchantStatusInactive {
		s.logFailedLoginActivity(ctx, user)
		return pkgErrs.New(response.HttpErrUnauthorized, errors.New("merchant is inactive"))
	}

	if user.MerchantStatus.String == constant.MerchantStatusDeactivated {
		return pkgErrs.New(response.HttpErrForbidden, errors.New("merchant is deactivated"))
	}

	if user.MerchantStatus.String == constant.MerchantStatusCreated {
		return pkgErrs.New(response.HttpErrForbidden, errors.New("merchant is not activated yet"))
	}

	return nil
}
