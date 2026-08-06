package user

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ResendInvitation		godoc
// @Summary				Resend invitation for expiry invited user
// @Description			Resend invitation for expiry invited user
// @ID					api-user-resend-invitation
// @Tags				API - User
// @Accept				json
// @Produce				json
// @Param 				id		path		string true "User ID for resend invitation"
// @Success				200  	{object}	response.ApiResponse
// @Failure				500  	{object}	response.ApiResponse
// @Router				/api/v1/users/{id}/resend-invitation [post]
func (c *UserController) ResendInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/ResendInvitation")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	invitedUserID := chi.URLParam(r, "user_id")
	if err := uuid.Validate(invitedUserID); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	// check if merchant exists
	merchant, err := c.merchantSvc.FindMerchantByID(r.Context(), user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	} else if merchant == nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound))
		return
	}

	// check if user exists
	invitedUser, err := c.userSvc.FindUserByID(ctx, invitedUserID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	} else if invitedUser.Status != constant.UserStatusInvited {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrUserAlreadyActivated))
		return
	} else if invitedUser.MerchantId != user.MerchantId {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrForbidden, constant.ErrMerchantNotAllowedPerformAction))
		return
	}

	// send generated invitation URL
	ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, user.MerchantId)
	if errInvitation := c.userSvc.SendGeneratedInvitationURL(ctx, &userModel.SendGeneratedInvitationRequest{
		Inviter:      user.Name,
		Email:        invitedUser.Email,
		MerchantName: merchant.Name,
		MerchantID:   merchant.UUID,
		UserID:       invitedUser.UUID,
		UserName:     invitedUser.Name,
		IsResend:     true,
	}); errInvitation != nil {
		response.SendApiResponseError(ctx, w, errInvitation)
		return
	}

	resp := map[string]any{"id": invitedUserID}
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagAccount,
		constant.ActivityUserResendInvitation,
		resp,
	)

	response.SendApiResponseOK(w, resp)
}
