package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *TransferInternalController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/Create")
	defer segment.End()

	var (
		submerchantId string
		payload       transfer.TransferRequest
	)

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendGeneralResponseError))

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidPayload(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrRequest, err))
		return
	}
	payload.ParentMerchantID = util.ParseUUID(merchant.MerchantId)
	payload.SourceMerchantID = payload.ParentMerchantID
	httputil.BindSubmerchantID(r, &submerchantId)
	if submerchantId != "" {
		payload.SourceMerchantID = util.ParseUUID(submerchantId)
	}

	if err := c.validator.Struct(payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrFieldValidation(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.transferService.Transfer(ctx, &payload)
	if err != nil {
		if errors.Is(err, constant.ErrRecipientIdNotFound) {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("recipient", payload.RecipientID))
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp.ToTransferResponse())
}
