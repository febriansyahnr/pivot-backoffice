package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

const (
	checkOldPassword = "check-old-password"
)

func (s *UserService) ChangePassword(
	ctx context.Context, userID string, OldPassword string, NewPassword string) (*userModel.ChangePasswordResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ChangePassword")
	defer segment.End()

	userData, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, pkgError.New(responseHttp.HttpErrInternal, err)
	}

	//Hash the incoming payload
	user := userModel.User{}
	oldPasswordHashed := user.EncryptPassword(OldPassword)
	newPasswordHashed := user.EncryptPassword(NewPassword)
	// Validate the old password
	isOldPasswordCorrect := (userData != nil) && (userData.Password == oldPasswordHashed)

	rateLimitReq := ratelimiter.RateLimit{
		Attribute:            userID,
		IsCheckResultCorrect: isOldPasswordCorrect,
		FeatureName:          checkOldPassword,
		Timestamp:            time.Now(),
	}
	err = s.rateLimiter.RateLimitFailedAttempt(ctx, &rateLimitReq)
	if err != nil {
		if errors.Is(err, constant.ErrRateLimiterExceedFailedAttempts) {
			return nil, pkgError.New(responseHttp.HttpErrTooManyRequest, err)
		}

		return nil, err
	}

	if !isOldPasswordCorrect {
		return nil, pkgError.New(responseHttp.HttpErrUnauthorized, fmt.Errorf("invalid password"))
	}

	countLastPassword := 4
	previousHistories, err := s.phRepo.FindByUserID(ctx, userID, &countLastPassword)
	if err != nil {
		return nil, pkgError.New(responseHttp.HttpErrDatabase, err)
	}
	for _, passwordHistory := range previousHistories {
		if passwordHistory.PasswordHashes == newPasswordHashed {
			return nil, pkgError.New(responseHttp.HttpErrDupCheck, fmt.Errorf("password has been used before"))
		}
	}

	// Change the password
	affected, errChange := s.userRepo.ChangePassword(ctx, userID, newPasswordHashed)
	if errChange != nil {
		return nil, pkgError.New(responseHttp.HttpErrInternal, err)
	}

	// Insert the new password to the password histories
	phId := uuid.New().String()
	err = s.phRepo.Insert(ctx, phId, userID, newPasswordHashed)
	if err != nil {
		return nil, pkgError.New(responseHttp.HttpErrInternal, err)
	}

	// Change user status from INVITED to ACTIVE
	if userData.Status == constant.UserStatusInvited {
		userData.Status = constant.UserStatusActive
	}

	// Update isChangePassword = 0
	userData.IsChangePassword = 0
	userData.Password = newPasswordHashed
	if err = s.userRepo.Update(ctx, userData); err != nil {
		return nil, err
	}

	return &userModel.ChangePasswordResponse{
		Updated: affected,
	}, nil
}
