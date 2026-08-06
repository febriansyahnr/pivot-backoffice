package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) ActivateUser(ctx context.Context, request *userModel.ActivateUserRequest) (*userModel.UserLoggedInResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ActivateUser")
	defer segment.End()

	existedUser, err := s.userRepo.FindUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if existedUser == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrUserNotFound)
	}

	if existedUser.Status != constant.UserStatusInvited {
		s.logger.Error(ctx, "User already activated", logger.String("uuid", existedUser.UUID))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrUserAlreadyActivated)
	}

	// Passing update user value
	existedUser.Status = constant.UserStatusActive
	existedUser.PinHash = sql.NullString{
		Valid:  true,
		String: util.HashString(request.PIN),
	}
	existedUser.IsChangePassword = 0
	existedUser.Password = util.HashString(request.Password)
	existedUser.UpdatedAt = time.Now().UTC()

	// Update user
	if err := s.userRepo.Update(ctx, existedUser); err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Remove redis activation token
	var keys []string
	_ = s.redis.Keys(ctx, redisTokenKey("", constant.UserIdentifierUserInvitation, ":", constant.UserTokenNamespace, ":", request.Token)).ScanSlice(&keys)
	if len(keys) > 0 {
		_ = s.redis.Del(ctx, keys...)
	}
	_ = s.redis.Del(ctx, fmt.Sprintf("backend-portal:login-attempt:%s", request.Email))

	// Generate access token for first login
	_, signedToken, err := s.GenerateTokenFromLogin2FA(ctx, existedUser.UUID)
	if err != nil {
		return nil, err
	}

	return &userModel.UserLoggedInResponse{
		UserInfo:     existedUser.ToResponse(),
		AccessToken:  signedToken,
		RefreshToken: existedUser.RefreshToken.String,
	}, nil
}
