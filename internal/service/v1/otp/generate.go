package otp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	commServicePb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/commService"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/random"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const totalResendOtpField = "total_resend"

func (s *service) SendGenerateOTPCode(ctx context.Context, request *otpModel.GenerateOTPCodeRequest) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/otp/SendGenerateOTPCode")
	defer segment.End()

	user, err := s.userRepo.FindUserByEmail(ctx, request.UserEmail)
	if err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if user == nil {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("email not registered"))

	} else if user.DeactivatedAt.Valid {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("user is deactivated"))

	} else if user.Blocked.Time.After(time.Now().UTC()) {
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user has been blocked"))

	} else if request.TwoFactorAuthMethod == constant.TwoFactorAuthMethodTOTP && user.TOTPStatus != constant.TOTPStatusActive {
		return "", pkgErrs.New(response.HttpErrRequest, constant.ErrTOTPNotActivated)
	}

	ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, user.MerchantId)

	twoFactorMethod := request.TwoFactorAuthMethod

	// If method is AUTO or empty, use user's preferred method
	if twoFactorMethod == constant.TwoFactorAuthMethodAuto || twoFactorMethod == "" {
		if user.Preferred2FAMethod != "" {
			twoFactorMethod = constant.TwoFactorMethod(user.Preferred2FAMethod)
		} else {
			twoFactorMethod = constant.TwoFactorAuthMethodAuto
		}
	}

	if twoFactorMethod == constant.TwoFactorAuthMethodTOTP {
		if user.TOTPStatus != constant.TOTPStatusActive {
			return "", pkgErrs.New(response.HttpErrRequest, constant.ErrTOTPRequiredButNotActive)
		}
		verifyTokenRequest := otpModel.GenerateTOTPVerifyTokenRequest{
			UserId:    user.UUID,
			UserEmail: user.Email,
			Feature:   request.Event,
		}
		return s.generator.GenerateTOTPVerifyToken(ctx, verifyTokenRequest)
	}

	if twoFactorMethod == constant.TwoFactorAuthMethodOTP {
		return s.generator.GenerateOTPCode(ctx, user.UUID, user.Email, request.Event)
	}

	// Default: Use TOTP if active, otherwise OTP
	if user.TOTPStatus == constant.TOTPStatusActive {
		verifyTokenRequest := otpModel.GenerateTOTPVerifyTokenRequest{
			UserId:    user.UUID,
			UserEmail: user.Email,
			Feature:   request.Event,
		}
		return s.generator.GenerateTOTPVerifyToken(ctx, verifyTokenRequest)
	}
	return s.generator.GenerateOTPCode(ctx, user.UUID, user.Email, request.Event)
}

func (s *service) GenerateTOTPVerifyToken(ctx context.Context, request otpModel.GenerateTOTPVerifyTokenRequest) (token string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/otp/GenerateTOTPVerifyToken")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	switch request.Feature {
	default:
		return "", pkgErrs.New(response.HttpErrForbidden, constant.ErrFeatureNotSupportTOTPAuth)

	case constant.OTPIdentifierUserLogin, constant.OTPIdentifierChangePassword, constant.OTPIdentifierResetPIN:
		// Features that are allowed using TOTP-based authentication
	}

	// Rate limitting
	limitRequest := &redisExt.Limit{
		Rate:   s.totpRateLimit.RequestLimit,
		Burst:  s.totpRateLimit.RequestLimit,
		Period: s.totpRateLimit.RequestWindow,
	}
	if resp, err := s.limiter.Allow(ctx, redisOTPKey(request.UserEmail, request.Feature, ":totp"), limitRequest); err != nil {
		s.logger.Error(ctx, "Failed while check rate limit on generate totp verify token", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if resp.Allowed == 0 {
		return "", pkgErrs.New(response.HttpErrResourceLocked, errors.New("please wait for a moment"))
	}

	// Generate verify token
	if token, err = s.jwt.GenerateTokenForOTP(ctx, request.UserId, request.Feature); err != nil {
		s.logger.Error(ctx, "Failed while generate totp verify token", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	// Store generated tokens in cache with TTL
	verifyOTPToken := otpModel.VerifyOTPToken{
		Token:               token,
		Email:               request.UserEmail,
		TwoFactorAuthMethod: constant.TwoFactorAuthMethodTOTP,
	}
	tokenKey := redisOTPKey(request.UserId, request.Feature, ":", constant.TokenOTPNamespace)

	_ = s.redis.Set(ctx, tokenKey, verifyOTPToken, request.Feature.ExpireDuration()+time.Minute)

	return (constant.TOTPTokenPrefixID + token), nil // totp-token/<verify-token>
}

func (s *service) GenerateOTPCode(ctx context.Context, id, email string, feature constant.OTPIdentifier) (token string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/otp/GenerateOTPCode")
	defer segment.End()

	// User Suspend Check
	suspend := otpModel.SuspendUser{}
	suspendKey := redisOTPKey(email, feature, ":suspend")
	if _ = s.redis.Get(ctx, suspendKey).Scan(&suspend); suspend.Status {
		return "", pkgErrs.New(
			response.HttpErrResourceLocked,
			fmt.Errorf("your otp request is currently suspended. please try again after %s", suspend.RetryAfter.In(loc).Format(constant.DatetimeFormat)),
		)
	}

	// Accumulated daily deliveries
	total := 0
	dataKey := redisOTPKey(email, feature, ":data")
	if err = s.redis.HGet(ctx, dataKey, "total_delivery").Scan(&total); err != nil && err != redisExt.ErrNil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if (total + 1) > feature.MaxSendOTP() {
		return "", pkgErrs.New(response.HttpErrTooManyRequest, errors.New("your request has exceeded the limit"))
	}

	totalResendOtp := 0
	if feature.NumWaitAfterSendOTP() > 0 {
		_ = s.redis.HGet(ctx, dataKey, totalResendOtpField).Scan(&totalResendOtp)
	}

	// Rate Limitting
	limit := &redisExt.Limit{
		Rate:   1,
		Burst:  1,
		Period: time.Duration(feature.GetResendDelaySeconds()) * time.Second,
	}
	if resp, err := s.limiter.Allow(ctx, redisOTPKey(email, feature), limit); err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if resp.Allowed == 0 {
		return "", pkgErrs.New(response.HttpErrResourceLocked, errors.New("please wait for a moment"))
	}

	lockKey := redisOTPKey(email, feature, ":lock")
	if can, err := s.redis.SetNX(ctx, lockKey, true, exclusiveLockDuration).Result(); err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if !can {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("the same request is in progress"))
	}
	defer func() {
		_ = s.redis.Del(ctx, lockKey)
	}()

	if token, err = s.jwt.GenerateTokenForOTP(ctx, id, feature); err != nil {
		return "", pkgErrs.New(response.HttpErrInternal, err)
	}

	total += 1
	totalResendOtp += 1
	otpCode := random.GenerateOTP(otpMaxDigits)

	var otpList []otpModel.OTPList
	if err = redisExt.ScanHashField(ctx, s.redis, dataKey, "otp", &otpList); err != nil {
		if err != redisExt.ErrNil {
			s.logger.Error(ctx, "Failed to retrieve OTP from Redis", logger.Error(err))
		}
		otpList = []otpModel.OTPList{}
	}

	existingOTPs := otpList
	// Keep the most recent OTPs up to the maximum allowed for this feature
	maxOTPs := feature.MaxSendOTP() - 1 // We subtract 1 because we're adding a new OTP right after
	if len(existingOTPs) > maxOTPs {
		existingOTPs = existingOTPs[len(existingOTPs)-maxOTPs:]
	}

	newOTP := otpModel.OTPList{
		OTP:       otpCode,
		ExpiredAt: time.Now().UTC().Add(feature.ExpireDuration()),
		Verify:    false,
	}

	updatedOTPList := append(existingOTPs, newOTP)

	serializedOTPList, err := json.Marshal(updatedOTPList)
	if err != nil {
		s.logger.Error(ctx, "Failed to marshal OTP list", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrInternal, err)
	}

	if err = s.redis.HSet(
		ctx,
		dataKey,
		"otp", string(serializedOTPList),
		"total_delivery", total,
		"total_attempts", 0,
		totalResendOtpField, totalResendOtp,
	).Err(); err != nil {
		s.logger.Error(ctx, "Failed to store OTP data in Redis", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrInternal, errors.New("unable to process OTP generation at this time, please try again later"))
	}

	if num := feature.NumWaitAfterSendOTP(); num > 0 && (totalResendOtp%num) == 0 {
		suspend.Status = true
		suspend.RetryAfter = time.Now().UTC().Add(feature.WaitTimeDuration())
		_ = s.redis.Set(ctx, suspendKey, suspend, feature.WaitTimeDuration())
	}

	content, _ := structpb.NewStruct(map[string]any{
		"OTPCode":             otpCode,
		"LogoURL":             s.config.PaperCommunication.EmailLogoURL,
		"DashboardGuideURL":   s.config.MerchantPortalConfig.DashboardGuideURL,
		"ExpiresAfterMinutes": math.Floor(feature.ExpireDuration().Minutes()),
	})

	emailRequest := &commServicePb.EmailRequest{
		Event:                feature.Event(),
		From:                 feature.EmailSender(),
		To:                   email,
		Subject:              fmt.Sprintf("Your %s OTP code is %s", s.config.PaperCommunication.PlatformName, otpCode),
		Content:              content,
		Priority:             commServicePb.EmailPriority_L0,
		ToBeRetriedOnFailure: true,
	}
	payload, _ := proto.Marshal(emailRequest)

	_ = s.rmq.Publish(ctx, rabbitMqExt.CommServiceEmailRoutingKey, nil, payload)

	verifyOTPToken := otpModel.VerifyOTPToken{
		Token:               token,
		Email:               email,
		TwoFactorAuthMethod: constant.TwoFactorAuthMethodOTP,
	}
	_ = s.redis.Set(ctx, redisOTPKey(id, feature, ":", constant.TokenOTPNamespace), verifyOTPToken, feature.ExpireDuration()+time.Minute)

	if ttl := s.redis.Client().TTL(ctx, dataKey).Val().Seconds(); ttl < 1 {
		now := time.Now().In(loc)
		_ = s.redis.Expire(ctx, dataKey, time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 00, loc).Sub(now))
	}
	return
}
