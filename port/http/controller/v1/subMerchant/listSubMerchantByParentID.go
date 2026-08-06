package subMerchant

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantController) ListSubMerchantByParentID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/api/v1/subMerchant/listSubMerchantByParentID")
	defer segment.End()

	var (
		startCreatedAt *time.Time // default nil
		endCreatedAt   *time.Time // default nil
		page           int64      = 1
		perPage        int64      = constant.DefaultPaginationPageSize
		err            error
	)

	merchant, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	urlQuery := r.URL.Query()
	startCreatedAtStr := urlQuery.Get("startCreatedAt")
	endCreatedAtStr := urlQuery.Get("endCreatedAt")
	pageStr := urlQuery.Get("page")
	perPageStr := urlQuery.Get("perPage")

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

	filter := &merchantModel.SubMerchantListFilter{
		ParentId:       merchant.MerchantId,
		MID:            urlQuery.Get("mid"),
		Name:           urlQuery.Get("name"),
		ShortName:      urlQuery.Get("shortName"),
		Keywords:       urlQuery.Get("keywords"),
		Status:         urlQuery.Get("status"),
		StartCreatedAt: startCreatedAt,
		EndCreatedAt:   endCreatedAt,
	}

	list, err := c.merchantSvc.ListSubMerchantByParentID(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}
