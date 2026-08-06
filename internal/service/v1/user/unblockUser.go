package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *UserService) UnblockUser(ctx context.Context, email string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/UnblockUser")
	defer segment.End()

	existedUser, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	} else if existedUser == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrUserNotFound)
	}

	// remove user-login:data to remove limit on total_delivery, total_resend and otp_code
	dataKey := redisOTPKey(email, constant.OTPIdentifierUserLogin, ":data")
	_ = s.redis.Del(ctx, dataKey)

	// Remove suspend key
	suspendKey := redisOTPKey(email, constant.OTPIdentifierUserLogin, ":suspend")
	_ = s.redis.Del(ctx, suspendKey)

	// Remove login attempt key
	_ = s.redis.Del(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", email))

	// Update user
	existedUser.Blocked = sql.NullTime{Time: time.Time{}, Valid: false}
	if errUpdate := s.userRepo.Update(ctx, existedUser); errUpdate != nil {
		return pkgErrors.New(response.HttpErrDatabase, errUpdate)
	}

	return nil
}

func redisOTPKey(identifier string, feature constant.OTPIdentifier, addition ...string) (key string) {
	key = fmt.Sprintf(
		constant.OTPKeyFormatting, identifier, feature.FeatureName(),
	)
	for _, s := range addition {
		key += s
	}
	return
}
