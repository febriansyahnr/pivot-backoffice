package user

import (
	"context"
	"database/sql"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) UserDetail(ctx context.Context, id string) (*userModel.UserResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/UserDetail")
	defer segment.End()

	user, err := s.userRepo.FindUserByID(ctx, id)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}

	if user == nil {
		return nil, errors.New(responseHttp.HttpErrNotFound, nil)
	}

	// Get last password change from password_histories
	var lastPasswordChange *sql.NullTime
	limit := 1
	passwordHistories, err := s.phRepo.FindByUserID(ctx, id, &limit)
	if err != nil {
		s.logger.Error(ctx, "Failed to get password history", logger.Error(err))
	} else if len(passwordHistories) > 0 {
		lastPasswordChange = &sql.NullTime{
			Time:  passwordHistories[0].CreatedAt,
			Valid: true,
		}
	}

	response := user.ToResponse()

	if lastPasswordChange != nil && lastPasswordChange.Valid {
		response.LastChangePassword = &lastPasswordChange.Time
	}

	return response, nil
}
