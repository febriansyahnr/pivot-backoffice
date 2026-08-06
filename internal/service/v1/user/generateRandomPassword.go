package user

import (
	"context"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *UserService) GenerateRandomPassword(ctx context.Context, userData *userModel.User) (*userModel.GenerateRandomPasswordResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/GenerateRandomPassword")
	defer segment.End()

	// Generate random password
	// don't need to check the error because the length is fixed
	randomPassword, _ := util.GenerateRandomString(10)

	// hash the random password
	user := userModel.User{}
	randomPasswordHashed := user.EncryptPassword(randomPassword)

	// Change the password
	_, errChange := s.userRepo.ChangePassword(ctx, userData.UUID, randomPasswordHashed)
	if errChange != nil {
		return nil, errors.New(responseHttp.HttpErrInternal, errChange)

	}

	return &userModel.GenerateRandomPasswordResponse{
		Email:    userData.Email,
		Password: randomPassword,
	}, nil
}
