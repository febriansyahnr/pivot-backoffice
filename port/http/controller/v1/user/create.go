package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/dictionary"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
)

// Create		godoc
// @Summary		Create user endpoint
// @Description	Create user endpoint
// @ID			user-create
// @Tags		API - User
// @Accept		json
// @Produce		json
// @Param 		id		path		string true "Merchant ID to get user list"
// @Param		Request	body		user.UserCreateRequest true "JSON Body for Create User"
// @Success		200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/users/ [post]
// @Security 	Bearer
func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/Create")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload userModel.UserCreateRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	// Authorization is handled by the permission middleware (team-members.create permission)
	// which properly supports both default admin role and custom roles with the required permission

	// check if merchant exists
	merchant, err := c.merchantSvc.FindMerchantByID(r.Context(), user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	} else if merchant == nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound))
		return
	}

	// Check existed user by email
	existedUser, err := c.userSvc.FindUserByEmail(ctx, strings.ToLower(payload.Email))
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	} else if existedUser != nil {
		errMessage := dictionary.TranslationAPIErrEmailAlreadyRegistered
		if dictionary.Dict != nil {
			errMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationAPIErrEmailAlreadyRegistered, payload.Email)
		}
		response.SendApiResponseError(
			ctx,
			w,
			pkgErrors.New(
				response.HttpErrRequest,
				errors.New(errMessage),
			),
		)
		return
	}

	userData := &userModel.User{
		UUID:             uuid.New().String(),
		MerchantId:       merchant.UUID,
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
		Inviter:      user.Name,
		Email:        userData.Email,
		MerchantName: merchant.Name,
		MerchantID:   merchant.UUID,
		UserID:       userData.UUID,
		UserName:     userData.Name,
	}); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, userData.ToResponse())
}
