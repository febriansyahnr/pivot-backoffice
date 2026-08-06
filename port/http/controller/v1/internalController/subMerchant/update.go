package submerchant

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (c *SubMerchantInternalController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendGeneralResponseError))

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if err := uuid.Validate(id); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidFieldFmt("id"))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("sub merchant id is required")))
		return
	}

	var payload merchantModel.UpdateMerchantOpenApiRequest
	payload.RequesterID = merchant.MerchantId
	payload.RequesterType = constant.UserTypeMerchant
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidPayload(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}
	payload.ID = id
	payload.ParentId = merchant.MerchantId

	if err := c.validate.Struct(payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrFieldValidation(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.merchantSvc.UpdateSubMerchantOpenApi(ctx, &payload)
	if err != nil {
		if e.Is(err, constant.ErrMerchantNotFound) {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("sub-account", id))
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
