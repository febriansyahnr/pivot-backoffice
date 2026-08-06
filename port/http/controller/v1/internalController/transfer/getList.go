package transfer

import (
	"context"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *TransferInternalController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/transfer/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))

	var err error

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	queryParams := r.URL.Query()

	payload := transfer.GetTransferListRequest{
		Type:         queryParams.Get("referenceType"),
		ReferenceID:  queryParams.Get("referenceId"),
		StrStartDate: queryParams.Get("startDate"),
		StrEndDate:   queryParams.Get("endDate"),
		Page:         constant.DefaultPage,
		PerPage:      constant.DefaultPaginationPageSize,
	}

	payload.MerchantID = merchant.MerchantId
	httputil.BindSubmerchantID(r, &payload.MerchantID)

	if err := c.validator.StructCtx(ctx, payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrFieldValidation(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	if payload.StrStartDate != "" { // If the start date has value then the end date must have value.
		payload.StartDate = util.ParseISO8601DatetimeToUTC(payload.StrStartDate)
		payload.EndDate = util.ParseISO8601DatetimeToUTC(payload.StrEndDate)
	}

	if strPage := queryParams.Get("page"); strPage != "" {
		if payload.Page, err = strconv.ParseInt(strPage, 10, 64); err != nil || payload.Page < 1 {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("page"))
			response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	if strPageSize := queryParams.Get("perPage"); strPageSize != "" {
		if payload.PerPage, err = strconv.ParseInt(strPageSize, 10, 64); err != nil || payload.PerPage < 1 {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("perPage"))
			response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	resp, err := c.transferService.GetList(ctx, &payload, payload.Page, payload.PerPage)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Data, resp.Meta)
}
