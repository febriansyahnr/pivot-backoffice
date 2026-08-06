package recurringContractHandler

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/recurringContract/Create")
	defer span.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	payload := recurringContractModel.CreateRecurringContractRequest{
		MerchantID: merchantAuth.MerchantId,
		CreatedBy:  merchantAuth.MerchantId,
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, constant.NewErrStringRequest(response.HttpErrRequest, constant.ErrCodeV2APIValidationError, "malformed request body payload"))
		return
	}
	httputil.BindSubmerchantID(r, &payload.MerchantID)

	if err := h.validate.StructCtx(ctx, payload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, constant.NewErrFieldValidation(err))
		return
	}
	if err := payload.Validate(); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, constant.NewErrStringRequest(response.HttpErrRequest, constant.ErrCodeV2APIValidationError, err.Error()))
		return
	}

	result, err := h.service.Create(ctx, payload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, result)
}
