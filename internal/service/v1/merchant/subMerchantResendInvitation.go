package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) SubMerchantResendInvitation(ctx context.Context, request *merchant.ResendInvitationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/SubMerchantResendInvitation")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, request.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "Failed when find merchant by id", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.String != request.ParentMerchantId {
		return pkgErrs.New(response.HttpErrForbidden, constant.ErrMerchantNotAllowedPerformAction)
	}

	user, err := s.UserSvc.FindUserByEmail(ctx, request.Email)
	if err != nil {
		s.logger.Error(ctx, "Failed when find user by email", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)

	} else if user == nil {
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrUserNotFound)

	} else if user.MerchantId != request.MerchantId {
		return pkgErrs.New(response.HttpErrForbidden, constant.ErrMerchantNotAllowedPerformAction)

	} else if user.Status != constant.UserStatusInvited {
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrUserAlreadyActivated)
	}

	ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, merchant.UUID)

	err = s.UserSvc.SendGeneratedInvitationURL(ctx, &userModel.SendGeneratedInvitationRequest{
		Inviter:      merchant.Name,
		Email:        user.Email,
		MerchantName: merchant.Name,
		MerchantID:   merchant.UUID,
		UserID:       user.UUID,
		UserName:     user.Name,
	})
	if err != nil {
		s.logger.Error(ctx, "Failed when send generated invitation URL", logger.Error(err))
		return err
	}
	return nil
}
