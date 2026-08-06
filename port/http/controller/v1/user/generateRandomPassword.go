package user

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GenerateRandomPassword	godoc
// @Summary					Generate random password
// @Description				Generate random password
// @ID						api-user-generate-random-password
// @Tags					API - User
// @Accept					json
// @Produce					json
// @Success					200  	{object}	response.ApiResponse{data=user.GenerateRandomPasswordResponse}
// @Failure					500  	{object}	response.ApiResponse
// @Router					/api/v1/generate-random-password [post]
// @Security				Bearer
func (c *UserController) GenerateRandomPassword(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/GenerateRandomPassowrd")
	defer segment.End()

	var (
		err error
	)

	// Get User Info from jwt token
	userInfoFromCtx := ctx.Value(constant.CtxUserInfoKey)
	userClaims, ok := userInfoFromCtx.(*userModel.UserTokenClaims)
	if !ok {
		err = fmt.Errorf("user not found")
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	// var payload user_model.GenerateRandomPasswordRequest
	// payload.Email = user.Email
	// payload.uuid = user.UUID

	// if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
	// 	response.SendApiResponseError(ctx, w, response.HttpErrRequest, err)
	// 	return
	// }

	// if err = c.validate.Struct(payload); err != nil {
	// 	response.SendApiResponseError(ctx, w, response.HttpErrRequest, err)
	// 	return
	// }

	// convert userClaims to user_model.User
	user := &userModel.User{
		UUID:         userClaims.UUID,
		Email:        userClaims.Email,
		Name:         userClaims.Name,
		Blocked:      sql.NullTime{Time: userClaims.Blocked, Valid: true},
		MerchantId:   userClaims.MerchantId,
		RefreshToken: sql.NullString{String: userClaims.RefreshToken, Valid: true},
		Role:         sql.NullString{String: userClaims.Role, Valid: true},
	}

	isSuccess, err := c.userSvc.GenerateRandomPassword(r.Context(), user)

	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&userClaims.MerchantId,
		&userClaims.UUID,
		constant.TagAccount,
		constant.ActivityUserGenerateRandomPassword,
		user,
	)

	response.SendApiResponseOK(w, isSuccess)
}
