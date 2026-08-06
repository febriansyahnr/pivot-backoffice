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
	checkCurrentPasswordName = "check-password"
)

func (s *UserService) CheckCurrentPassword(ctx context.Context, userID string, password string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/CheckCurrentPassword")
	defer segment.End()

	userData, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	} else if userData == nil {
		return constant.ErrUserNotFound
	}

	isPasswordCorrect := userData.Password == util.HashString(password)
	rateLimitReq := ratelimiter.RateLimit{
		Attribute:            userID,
		IsCheckResultCorrect: isPasswordCorrect,
		FeatureName:          checkCurrentPasswordName,
		Timestamp:            time.Now(),
	}
	err = s.rateLimiter.RateLimitFailedAttempt(ctx, &rateLimitReq)
	if err != nil {
		if err == constant.ErrRateLimiterExceedFailedAttempts {
			return pkgErrors.New(httpResponse.HttpErrTooManyRequest, err)
		}
		return err
	}

	// validate password
	if userData.Password != util.HashString(password) {
		s.logger.Error(ctx, constant.ErrInvalidPassword.Error(), logger.Error(constant.ErrInvalidPassword))
		return pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrInvalidPassword)
	}

	return nil
}
