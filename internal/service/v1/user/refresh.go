package user

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Refresh service for refresh token
func (s *UserService) Refresh(ctx context.Context, email, refreshToken string) (*userModel.User, string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/Refresh")
	defer segment.End()

	// Find user by email
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "error when finding user by email", logger.Error(err))
		return nil, "", errors.New(response.HttpErrDatabase, err)
	}

	if user.RefreshToken.String != refreshToken {
		s.logger.Error(ctx, "invalid refresh token")
		return nil, "", errors.New(response.HttpErrUnauthorized, fmt.Errorf("invalid refresh token"))
	}

	// generate new refresh token
	refreshToken, errSign := s.JWT.GenerateRefreshToken(ctx, user, time.Now().UTC().Add(constant.REFRESH_EXPIRATION_DURATION))
	if errSign != nil {
		s.logger.Error(ctx, "error generate refresh token", logger.Error(errSign))
		return nil, "", errors.New(response.HttpErrInternal, errSign)
	}
	user.RefreshToken.Valid = true
	user.RefreshToken.String = refreshToken
	user.DeviceIdentifier, _ = ctx.Value(constant.CtxUserDeviceIdentifierKey).(string)

	// update user
	errUpdate := s.userRepo.UpdateRefreshToken(ctx, user.UUID, user.RefreshToken.String)
	if errUpdate != nil {
		s.logger.Error(ctx, "error update refresh token", logger.Error(errUpdate))
		return nil, "", errors.New(response.HttpErrDatabase, errUpdate)
	}

	// generate new token for access token
	accessToken, err := s.JWT.GenerateAccessToken(ctx, user)
	if err != nil {
		s.logger.Error(ctx, "error generate access token", logger.Error(err))
		return nil, "", errors.New(response.HttpErrInternal, err)
	}

	// insert jwt token to redis and set ttl same as jwt token expiration
	s.redis.Set(
		ctx,
		fmt.Sprintf("backend-portal:access-token:%s:%s", user.Email, user.DeviceIdentifier),
		accessToken, time.Now().UTC().Add(constant.LOGIN_EXPIRATION_DURATION).Sub(time.Now().UTC()),
	)
	return user, accessToken, nil
}
