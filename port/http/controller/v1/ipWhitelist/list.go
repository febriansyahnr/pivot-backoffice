package ipWhitelistController

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *IPWhitelistConfigurationController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/IPWhitelist/GetList")
	defer segment.End()
	var (
		page     int = constant.DefaultPage
		pageSize int = constant.DefaultPaginationPageSize
		err      error
	)

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload ipwhitelistModel.GetIPWhitelistConfiguration
	payload.MerchantID = user.MerchantId
	payload.IP = r.URL.Query().Get("ip")
	payload.Status = r.URL.Query().Get("status")

	strPage := r.URL.Query().Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendOpenApiResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := r.URL.Query().Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendOpenApiResponseError(w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	if err := c.validator.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	configs, err := c.svc.List(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	list := []*ipwhitelistModel.IPWhitelistConfigurationResponse{}
	for _, config := range configs.Data.([]*ipwhitelistModel.IPWhitelistConfiguration) {
		list = append(list, config.ToResponseModel())
	}
	configs.Data = list

	response.SendApiResponseOK(w, configs)
}
