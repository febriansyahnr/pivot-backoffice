package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ResendInvitation		godoc
// @Summary		ResendInvitation user endpoint
// @Description	ResendInvitation user endpoint
// @ID			crm-user-resend-invitation
// @Tags		CRM - User
// @Accept		json
// @Produce		json
// @Param 		Request	body		user.ResendEmailInvitationRequest true "JSON body to resend invitation"
// @Success		200  	{object}	response.Response{data=user.UserResponse}
// @Failure		500  	{object}	response.Response
// @Router		/crm/v1/users/{id}/resend-invitation [post]
// @Security 	Bearer
func (c *CRMUserController) ResendInvitation(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/user/ResendInvitation")
	defer segment.End()

	request := &userModel.ResendEmailInvitationRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("cannot unmarshal: %w", err)))
		return
	}
	if err := c.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	// check if user exists
	invitedUser, err := c.userSvc.FindUserByEmail(ctx, request.Email)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	} else if invitedUser == nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrUserNotFound))
		return
	} else if invitedUser.Status != constant.UserStatusInvited {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrUserAlreadyActivated))
		return
	}

	// send generated invitation URL
	ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, invitedUser.MerchantId)
	if errInvitation := c.userSvc.SendGeneratedInvitationURL(ctx, &userModel.SendGeneratedInvitationRequest{
		Inviter:      invitedUser.MerchantName,
		Email:        invitedUser.Email,
		MerchantName: invitedUser.MerchantName,
		MerchantID:   invitedUser.MerchantId,
		UserID:       invitedUser.UUID,
		UserName:     invitedUser.Name,
		IsResend:     true,
	}); errInvitation != nil {
		response.SendGeneralResponseError(w, errInvitation)
		return
	}

	response.SendGeneralResponseOK(w, map[string]any{
		"id":    invitedUser.UUID,
		"email": request.Email,
	})
}
