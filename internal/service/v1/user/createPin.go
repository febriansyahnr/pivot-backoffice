package user

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) CreatePin(ctx context.Context, userID, pin string) error {
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
	if user.PinHash.Valid {
		err = errors.New("pin has already created")
		s.logger.Error(ctx, "pin is not empty when create new", logger.Error(err))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	if err := util.IsValidPin(pin); err != nil {
		s.logger.Error(ctx, "invalid pin format", logger.Error(err))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	// hash pin
	hashedPin := util.HashString(pin)
	return s.userRepo.UpdatePin(ctx, userID, hashedPin)
}
