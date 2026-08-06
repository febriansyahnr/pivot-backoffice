package user

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) ChangePin(ctx context.Context, userID, pin, newPin string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ChangePin")
	defer segment.End()

	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	} else if user == nil {
		return constant.ErrUserNotFound
	}

	// throw error once pin exist in database
	if !user.PinHash.Valid {
		s.logger.Error(ctx, constant.ErrPINNotCreatedYet.Error(), logger.Error(constant.ErrPINNotCreatedYet))
		return pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrPINNotCreatedYet)
	}

	if user.PinHash.String != util.HashString(pin) {
		s.logger.Error(ctx, constant.ErrInvalidPIN.Error(), logger.Error(constant.ErrInvalidPIN))
		return pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrInvalidPIN)
	}

	if err := util.IsValidPin(newPin); err != nil {
		s.logger.Error(ctx, "invalid new pin format", logger.Error(err))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	return s.userRepo.UpdatePin(ctx, userID, util.HashString(newPin))
}

func (s *UserService) ResetPIN(ctx context.Context, id, pin string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ResetPIN")
	defer segment.End()

	user, err := s.userRepo.FindUserByID(ctx, id)
	if err != nil {
		return pkgErrors.New(httpResponse.HttpErrDatabase, err)

	} else if user == nil {
		return pkgErrors.New(httpResponse.HttpErrUnprocessableContent, constant.ErrUserNotFound)
	}

	if err := util.IsValidPin(pin); err != nil {
		s.logger.Error(ctx, "invalid pin format", logger.Error(err))
		return pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	return s.userRepo.UpdatePin(ctx, user.UUID, util.HashString(pin))
}
