package submerchant

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantInternalController) AssignAdmin(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/AssignAdmin")
	defer segment.End()

	var (
		err           error
		submerchantId string
		payload       merchant.SubMerchantAdminRequest
	)

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendGeneralResponseError))

	httputil.BindSubmerchantID(r, &submerchantId)
	if submerchantId == "" {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrRequiredField(constant.HeaderXSubMerchantID))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrMissingSubMerchantId))
		return
	}
	payload.MerchantId = submerchantId
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidPayload(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}
	if err = c.validate.Struct(payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrFieldValidation(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.merchantSvc.AssignSubMerchantAdmin(ctx, &payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendGeneralResponseOK(w, nil)
}
