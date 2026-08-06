package user

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *UserService) CreateMerchantUser(ctx context.Context, payload *userModel.MerchantUserRequest) (*userModel.User, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/CreateMerchantUser")
	defer segment.End()

	existingUser, err := s.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrCheckUser)
	}
	if existingUser != nil {
		return nil, errors.New(response.HttpErrDupCheck, constant.ErrUserAlreadyExists)
	}

	user := userModel.NewMerchantUser(payload)
	err = s.Create(ctx, user)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrCreateUser)
	}

	role, err := s.roleSvc.FindRoleBySlug(ctx, strings.ToLower(payload.Role))
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrCreateRole)
	}
	if role == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrRoleNotFound)
	}

	userRole := userRole.New(user.UUID, role.UUID)
	err = s.userRoleSvc.Create(ctx, userRole)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrAssignUserToRole)
	}

	if payload.Invitation {
		ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, payload.MerchantId)

		err = s.SendGeneratedInvitationURL(ctx, &userModel.SendGeneratedInvitationRequest{
			Inviter:      payload.MerchantName,
			Email:        user.Email,
			MerchantName: payload.MerchantName,
			MerchantID:   payload.MerchantId,
			UserID:       user.UUID,
			UserName:     user.Name,
		})
		if err != nil {
			// Pass through so typed downstream errors (502/503/504) from paper-communication
			// propagate instead of collapsing to an opaque 500.
			return nil, err
		}
	}
	return user, nil
}
