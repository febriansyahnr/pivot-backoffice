package otp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) ValidateOTPCode(ctx context.Context, data *otpModel.VerifyOTP) (token string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/otp/ValidateOTPCode")
	defer segment.End()

	// Key list
	dataKey := redisOTPKey(data.Email, data.Identifier, ":data")
	lockKey := redisOTPKey(data.Email, data.Identifier, ":lock")

	var (
		// We need to store the OTP list directly instead of in the OTPCache struct
		otpList = []otpModel.OTPList{}
		// Total OTP verification attempts
		totalAttempts int
		// Get the OTP data from Redis - it's stored as a direct JSON array
		otpJson string
	)

	// Retrieve the list of OTP codes generated when using the 2FA method with OTP (random 6 digits code)
	if data.TwoFactorAuthMethod == constant.TwoFactorAuthMethodOTP {
		if err = s.redis.HGet(ctx, dataKey, "otp").Scan(&otpJson); err != nil {
			if err == redisExt.ErrNil {
				return "", pkgErrs.New(response.HttpErrRequest, errors.New("OTP data is not registered"))
			}
			return "", pkgErrs.New(response.HttpErrDatabase, err)
		}
		if err = json.Unmarshal([]byte(otpJson), &otpList); err != nil {
			s.logger.Warn(ctx, "Failed to unmarshal OTP data", logger.Error(err))
		}
	}

	if err = s.redis.HGet(ctx, dataKey, "total_attempts").Scan(&totalAttempts); err != nil && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "Failed to get total attempts from Redis", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrInternal, errors.New("unable to process OTP verification at this time, please try again later"))
	}

	// Try to acquire a lock to prevent race conditions
	acquired, err := s.redis.SetNX(ctx, lockKey, "lock", 30*time.Second).Result()
	if err != nil {
		s.logger.Error(ctx, "Failed to acquire lock", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrInternal, fmt.Errorf("failed to process OTP verification: %w", err))
	}
	defer func() { _ = s.redis.Del(context.Background(), lockKey) }()

	if !acquired {
		return "", pkgErrs.New(response.HttpErrRequest, errors.New("the same request is in progress"))

	} else if (totalAttempts + 1) > data.Identifier.MaxFailedAttempts() {
		return "", pkgErrs.New(response.HttpErrTooManyRequest, errors.New("max attempts limit has been exceeded"))
	}

	// Check if this is a test account that can bypass OTP validation
	bypassIsTrue, err := constant.IsTestingAccount(ctx, config.Environment(), data.ID, data.Email)
	if err != nil {
		s.logger.Warn(ctx, "failed to check testing account", logger.Error(err))
	}

	var (
		otpMatched, otpExpired, otpFound bool
		matchedOTPIndex                  = -1
	)

	if data.TwoFactorAuthMethod == constant.TwoFactorAuthMethodTOTP {
		user, err := s.userRepo.FindUserTOTPDataByID(ctx, data.ID)
		if err != nil {
			s.logger.Error(ctx, "Failed while find user totp data by id on validate otp code", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

		} else if user == nil {
			return "", pkgErrs.New(response.HttpErrNotFound, constant.ErrUserNotFound)
		}

		otpMatched, err = s.ValidateTOTPCode(ctx, &otpModel.VerifyTOTPRequest{
			WrappedSecret:  user.TOTPWrappedSecret,
			EncryptVersion: user.TOTPEncryptVersion,
			Code:           data.OTPCode,
		})
		if err != nil {
			s.logger.Error(ctx, "Failed while validate totp code on validate otp code", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
		}

	} else {
		for i, otpItem := range otpList {
			if bypassIsTrue && data.OTPCode == constant.OTPHardCode {
				matchedOTPIndex = i
				otpMatched, otpFound = true, true
				break

			} else if otpItem.OTP == data.OTPCode {
				otpFound = true
				if time.Now().UTC().After(otpItem.ExpiredAt) {
					otpExpired = true

				} else {
					otpMatched, matchedOTPIndex = true, i
				}
				break
			}
		}
	}

	// Process OTP validation result
	defer func() {

		// Always update the attempts counter
		_ = s.redis.HSet(ctx, dataKey, "total_attempts", totalAttempts+1)

		// Set TTL data for TOTP authentication method
		if data.TwoFactorAuthMethod == constant.TwoFactorAuthMethodTOTP && totalAttempts == 0 {
			now := time.Now().In(loc)
			_ = s.redis.Expire(ctx, dataKey, time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 00, loc).Sub(now))
		}

		// If OTP matched, mark it as verified and save
		if otpMatched && matchedOTPIndex >= 0 {
			otpList[matchedOTPIndex].Verify = true

			// Reset resend counter when verified
			_ = s.redis.HSet(ctx, dataKey, totalResendOtpField, 0)

			if updatedJson, err := json.Marshal(otpList); err != nil {
				s.logger.Warn(ctx, "Failed to marshal updated OTP list", logger.Error(err))
				// Even if marshaling fails, we still need to prevent OTP reuse
				// Add a failsafe by setting a verified flag directly in Redis
				_ = s.redis.HSet(ctx, dataKey, "otp_"+data.OTPCode+"_verified", "true")

			} else {
				_ = s.redis.HSet(ctx, dataKey, "otp", string(updatedJson))
			}
		}
	}()

	if !otpMatched {
		if otpFound && otpExpired {
			return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("OTP expired"))
		}

		remainingChances := data.Identifier.MaxFailedAttempts() - totalAttempts - 1
		if remainingChances <= 0 {
			errMsg := errors.New("wrong code and max attempts exceeded. please request a new password and try again")

			if data.Identifier == constant.OTPIdentifierUserLogin || data.Identifier == constant.OTPIdentifierFirstTimeLogin {
				blocked := sql.NullTime{Time: time.Now().UTC().Add(constant.BLOCKED_DURATION), Valid: true}
				if err := s.userRepo.BlockUser(ctx, data.ID, blocked); err != nil {
					return "", pkgErrs.New(response.HttpErrDatabase, err)
				}
				_ = s.jwt.RemoveIterateTokenFromRedis(ctx, data.Email) // Revoke all user's token

				errMsg = errors.New("user is blocked, too many login attempts")
			}
			return "", pkgErrs.New(response.HttpErrTooManyRequest, errMsg)
		}
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("wrong code, %d more chances to input", remainingChances))
	}

	// Generate a one-time token to access the intended feature.
	if token, err = s.jwt.GenerateTokenForFeature2FA(ctx, data.ID, data.Identifier); err != nil {
		return "", pkgErrs.New(response.HttpErrInternal, err)
	}
	_ = s.redis.Set(
		ctx, redisOTPKey(data.ID, data.Identifier, ":", constant.TokenFeature2FANamespace, ":", token), data.Email, data.Identifier.ExpireDuration(),
	)

	// Reset total attempts
	totalAttempts = -1

	// Reset Rate Limitting
	_ = s.limiter.Reset(ctx, redisOTPKey(data.Email, data.Identifier))

	// Reset Suspend User (If Exists)
	_ = s.redis.Del(ctx, redisOTPKey(data.Email, data.Identifier, ":suspend"))

	return
}
