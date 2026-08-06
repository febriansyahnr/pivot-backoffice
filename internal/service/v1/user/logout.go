package user

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) Logout(ctx context.Context, email string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/Logout")
	defer segment.End()

	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "failed to find user by email", logger.Error(err))
		return errors.New(response.HttpErrDatabase, err)
	}

	if user == nil {
		s.logger.Error(ctx, "email not registered")
		return errors.New(response.HttpErrNotFound, fmt.Errorf("email not registered"))
	}

	// insert jwt token to redis and set ttl same as jwt token expiration
	deviceID := util.GenerateDeviceID(ctx.Value(constant.CtxUserAgentKey).(string))
	s.redis.Del(ctx, fmt.Sprintf("backend-portal:access-token:%s:%s", user.Email, deviceID))

	// remove login attempt from redis
	s.redis.Del(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", email))

	return nil
}
