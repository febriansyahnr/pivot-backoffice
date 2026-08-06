package transfer

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *TransferInternalController) GetById(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/GetById")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if _, err := uuid.Parse(id); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("transferId"))
		response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	merchantId := merchant.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	resp, err := c.transferService.GetById(ctx, id, merchantId)
	if err != nil {
		if errors.Is(err, constant.ErrTransferNotFound) {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("transfer", id))
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp.ToTransferResponse())
}
