package submerchant

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantInternalController) ListSubMerchantByParentID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/ListSubMerchantByParentID")
	defer segment.End()

	var (
		startCreatedAt *time.Time // default nil
		endCreatedAt   *time.Time // default nil
		page           int64      = 1
		perPage        int64      = constant.DefaultPaginationPageSize
		err            error
	)
	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))

	merchantInfo := r.Context().Value(constant.CtxMerchantInfo)
	merchant, ok := merchantInfo.(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	loggedInMerchantId := merchant.MerchantId
	httputil.BindSubmerchantID(r, &loggedInMerchantId)
	loggedInUserType := constant.UserTypeMerchant
	httputil.BindLoggedInUserType(r, &loggedInUserType)

	startCreatedAtStr := r.URL.Query().Get("startCreatedAt")
	endCreatedAtStr := r.URL.Query().Get("endCreatedAt")
	statusStr := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")

	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("page"))
			response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}
	if perPageStr != "" {
		perPage, err = strconv.ParseInt(perPageStr, 10, 64)
		if err != nil {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("perPage"))
			response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead")))
			return
		}
	}

	if statusStr != "" {
		if statusStr == "ACTIVE" {
			statusStr = "1"
		} else {
			statusStr = "0"
		}
	}

	if startCreatedAtStr != "" {
		parsedStartCreatedAt, err := time.Parse(util.UTCLayout, startCreatedAtStr)
		if err != nil {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("startCreatedAt"))
			response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid startCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		startCreatedAt = &parsedStartCreatedAt
	}
	if endCreatedAtStr != "" {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("endCreatedAt"))
		parsedEndCreatedAt, err := time.Parse(util.UTCLayout, endCreatedAtStr)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid endCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		endCreatedAt = &parsedEndCreatedAt
	}

	filter := &merchantModel.SubMerchantListFilter{
		ParentId:       merchant.MerchantId,
		MID:            r.URL.Query().Get("mid"),
		Name:           r.URL.Query().Get("name"),
		ShortName:      r.URL.Query().Get("shortName"),
		Keywords:       r.URL.Query().Get("keywords"),
		Email:          r.URL.Query().Get("email"),
		Status:         statusStr,
		StartCreatedAt: startCreatedAt,
		EndCreatedAt:   endCreatedAt,
	}

	list, err := c.merchantSvc.ListSubMerchantByParentID(ctx, filter, page, perPage)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	responses, err := buildSubmerchantResponses(list.Data)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponsePaginationOK(w, responses, list.Meta)

}

func buildSubmerchantResponses(data interface{}) ([]*merchantModel.SubMerchantResponse, error) {
	merchants, ok := data.([]*merchantModel.Merchant)
	if !ok {
		return nil, errors.New(response.HttpErrInternal, fmt.Errorf("unexpected data format"))
	}
	var responses []*merchantModel.SubMerchantResponse
	for _, merchant := range merchants {
		responses = append(responses, merchant.ToSubMerchantResponse())
	}

	return responses, nil
}
