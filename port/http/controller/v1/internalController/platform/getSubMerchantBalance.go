package platformInternalController

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *platformController) GetSubMerchantBalance(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/platform/GetSubMerchantBalance")
	defer segment.End()

	var (
		err     error
		request *platform.GetBulkBalanceRequest
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("invalid request body")))
		return
	}
	err = c.validate.Struct(request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	if request.PerPage == 0 {
		request.PerPage = constant.DefaultPlatformSubMerchantBalancePageSize
	}
	if request.Page == 0 {
		request.Page = constant.DefaultPage
	}

	request.MerchantID = merchantAuth.MerchantId
	resp, err := c.platformSvc.GetSubMerchantBalances(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponsePaginationOK(w, resp.Data, resp.Meta)
}
