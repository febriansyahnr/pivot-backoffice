package user

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
)

// GetInvitationURL	godoc
// @Summary			Endpoint to get invitation URL to support automated testing (QA)
// @Description		Endpoint to get invitation URL to support automated testing (QA)
// @ID				tests-users-invitations
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Param 			encoded_email	path	string true "Encoded base64 user email"
// @Success			200  	{object}	response.ApiResponse{data=user.InvitationURLResp}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/tests/users/invitations/{encoded_email} [get]
// @Security 		Bearer
// @Header       	All     {string}  X-Automated-Test-Key "{"key": "value"}"
func (c *UserController) GetInvitationURL(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/user/GetInvitationURL")
	defer segment.End()

	rawEmail, err := base64.RawURLEncoding.DecodeString(chi.URLParam(r, "encoded_email"))
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid format")))
		return
	}

	merchantId := r.Header.Get(constant.HeaderXMerchantId)
	if url, err := c.userSvc.GetInvitationURL(ctx, merchantId, string(rawEmail)); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, user.InvitationURLResp{URL: url})
	}
}
