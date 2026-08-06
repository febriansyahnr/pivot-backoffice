package user

import (
	"encoding/json"
	"net/http"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Invite	godoc
// @Summary		Invite new user endpoint
// @Description	Invite new user endpoint
// @ID			invite-new-user
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Success		200
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/users/validate-invitation [post]
func (c *UserController) ValidateInvitationToken(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/ValidateInvitationToken")
	defer segment.End()

	request := userModel.ValidateInvitationRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	result, err := c.userSvc.ValidateInvitationToken(ctx, request.Token)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
