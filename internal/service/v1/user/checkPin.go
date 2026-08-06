package user

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const (
	checkPINName = "check-pin"
)

func (s *UserService) CheckCurrentPin(ctx context.Context, userID, pin string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/CreatePin")
	defer segment.End()

	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return constant.ErrUserNotFound
	}

	// throw error once pin exist in database
	if !user.PinHash.Valid {
		s.logger.Error(ctx, constant.ErrPINNotCreatedYet.Error(), logger.Error(constant.ErrPINNotCreatedYet))
		return pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrPINNotCreatedYet)
	}

	isPinCorrect := user.PinHash.String == util.HashString(pin)
	rateLimitReq := ratelimiter.RateLimit{
		Attribute:            userID,
		IsCheckResultCorrect: isPinCorrect,
		FeatureName:          checkPINName,
		Timestamp:            time.Now(),
	}
	err = s.rateLimiter.RateLimitFailedAttempt(ctx, &rateLimitReq)
	if err != nil {
		if err == constant.ErrRateLimiterExceedFailedAttempts {
			return pkgErrors.New(httpResponse.HttpErrTooManyRequest, err)
		}
		return err
	}

	// hash pin
	if !isPinCorrect {
		s.logger.Error(ctx, constant.ErrInvalidPIN.Error(), logger.Error(constant.ErrInvalidPIN))
		s.ActivityLog(ctx, &user.MerchantId, &user.ID, checkPINName, constant.ActivityUserFailedCheckPin, map[string]string{
			"email": user.Email,
			"pin":   "********",
		})
		return pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrInvalidPIN)
	}
	return nil
}
