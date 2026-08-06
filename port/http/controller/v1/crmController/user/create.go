package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/dictionary"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
)

// Create		godoc
// @Summary		Create user endpoint
// @Description	Create user endpoint
// @ID			crm-user-create
// @Tags		CRM - User
// @Accept		json
// @Produce		json
// @Param		Request	body		user.CRMUserCreateRequest true "JSON Body for Create User"
// @Success		200  	{object}	response.Response{data=user.UserResponse}
// @Failure		500  	{object}	response.Response
// @Router		/crm/v1/users/ [post]
// @Security 	Bearer
func (c *CRMUserController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/crmController/v1/user/Create")
	defer segment.End()

	var (
		err                      error
		merchantId, merchantName string
		merchantModel            *merchant.Merchant
	)

	var payload userModel.CRMUserCreateRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrValidation, err))
		return
	}

	// check if merchant_id is submitted
	if payload.MerchantId != "" {
		merchantModel, err = c.merchantSvc.FindMerchantByID(r.Context(), payload.MerchantId)
		if err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		} else if merchantModel == nil {
			response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound))
			return
		}

		merchantId = merchantModel.UUID
		merchantName = merchantModel.Name
	}

	// Check existed user by email
	existedUser, err := c.userSvc.FindUserByEmail(ctx, strings.ToLower(payload.Email))
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	} else if existedUser != nil {
		response.SendApiResponseError(
			ctx,
			w,
			errors.New(
				response.HttpErrRequest,
				fmt.Errorf("%s", dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationAPIErrEmailAlreadyRegistered, payload.Email)),
			),
		)
		return
	}

	userData := &userModel.User{
		UUID:             uuid.New().String(),
		MerchantId:       merchantId,
		Email:            strings.ToLower(payload.Email),
		Status:           constant.UserStatusInvited,
		Name:             payload.Name,
		IsChangePassword: 1,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// Generate random password
	randomPassword, _ := util.GenerateRandomString(10)

	// hash password
	userData.Password = userData.EncryptPassword(randomPassword)

	if err = c.userSvc.Create(ctx, userData); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check if role exists
	existedRole, err := c.roleSvc.FindRoleBySlug(ctx, payload.RoleSlug)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// assign role to the user
	newUserRole := &userRole.UserRole{
		UUID:      uuid.New().String(),
		UserID:    userData.UUID,
		RoleID:    existedRole.UUID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err = c.userRoleSvc.Create(ctx, newUserRole); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// assign role to response
	userData.Role = sql.NullString{String: existedRole.Name, Valid: true}

	// send generated invitation URL
	ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, userData.MerchantId)
	if err = c.userSvc.SendGeneratedInvitationURL(ctx, &userModel.SendGeneratedInvitationRequest{
		Inviter:      "System",
		Email:        userData.Email,
		MerchantName: merchantName,
		MerchantID:   merchantId,
		UserID:       userData.UUID,
		UserName:     userData.Name,
	}); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, userData.ToResponse())
}
