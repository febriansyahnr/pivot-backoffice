package platform

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *PlatformController) GetSubMerchantUserList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "/port/http/controller/v1/platform/GetSubMerchantUserList")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrUnauthorized, constant.ErrInvalidAccess))
		return
	}

	query := r.URL.Query()

	request := platform.GetSubMerchantUsersRequest{
		ParentMerchantId: user.MerchantId,
		MerchantId:       query.Get("merchantId"),
	}
	if request.MerchantId == "" {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidMerchantId))
		return
	}
	request.Keyword = query.Get("keyword")
	request.RoleId = query.Get("roleId")

	// page
	request.Page = constant.DefaultPage
	page := query.Get("page")
	if page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid page value")))
			return
		}
		request.Page = int64(p)
	}

	request.PerPage = constant.DefaultPageSize
	perPage := query.Get("perPage")
	if perPage != "" {
		p, err := strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid perPage value")))
			return
		}
		request.PerPage = int64(p)
	}

	// sort
	request.SortBy = query.Get("sortBy")
	request.SortOrder = query.Get("sortOrder")

	data, err := c.platformSvc.GetSubMerchantUserList(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, data)
}
