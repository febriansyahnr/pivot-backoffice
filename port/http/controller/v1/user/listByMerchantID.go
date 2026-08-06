package user

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ListByMerchantID	godoc
// @Summary			Get user list by merchant_id
// @Description		Get user list by merchant_id
// @ID				user-list
// @Tags			API - User
// @Accept			json
// @Produce			json
// @Success			200  	{object}	response.ApiResponse{data=user.UserResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/users [get]
// @Security 		Bearer
func (c *UserController) ListByMerchantID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/user/ListByMerchantID")
	defer segment.End()

	var (
		startCreatedAt *time.Time // default nil
		endCreatedAt   *time.Time // default nil
		filterName     string
		filterRoleId   string
		page           int64 = 1
		perPage        int64 = constant.DefaultPaginationPageSize
		err            error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Get query params
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")
	startCreatedAtStr := r.URL.Query().Get("startCreatedAt")
	endCreatedAtStr := r.URL.Query().Get("endCreatedAt")
	filterName = r.URL.Query().Get("name")
	filterRoleId = r.URL.Query().Get("roleId")

	// Validation and parsing
	if startCreatedAtStr != "" {
		parsedStartCreatedAt, err := time.Parse(util.UTCLayout, startCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid startCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		startCreatedAt = &parsedStartCreatedAt
	}
	if endCreatedAtStr != "" {
		parsedEndCreatedAt, err := time.Parse(util.UTCLayout, endCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid endCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		endCreatedAt = &parsedEndCreatedAt
	}
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}
	if perPageStr != "" {
		perPage, err = strconv.ParseInt(perPageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead")))
			return
		}
	}

	filter := &userModel.ListUsersByMerchantIDRequest{
		MerchantID:     user.MerchantId,
		StartCreatedAt: startCreatedAt,
		EndCreatedAt:   endCreatedAt,
		Name:           filterName,
		RoleID:         filterRoleId,
	}

	list, err := c.userSvc.ListUsersByMerchantID(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}
