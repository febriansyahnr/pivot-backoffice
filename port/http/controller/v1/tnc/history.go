package tnc

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *TNCSigningController) History(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/tnc/History")
	defer segment.End()

	claims, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok || claims == nil {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}
	if claims.MerchantId == "" {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
		return
	}

	var (
		page     = constant.DefaultPage
		pageSize = constant.DefaultPaginationPageSize
		err      error
	)

	query := r.URL.Query()
	payload := tncModel.SigningHistoryQuery{
		MerchantID: claims.MerchantId,
		TNCVersion: query.Get("version"),
	}

	strPage := query.Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := query.Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	result, err := c.service.GetSigningHistory(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
