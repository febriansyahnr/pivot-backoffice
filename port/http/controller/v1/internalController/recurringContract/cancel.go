package recurringContractHandler

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

func (h *handler) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/recurringContract/Cancel")
	defer span.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	request := model.CancelRecurringContractRequest{
		MerchantID:  merchantAuth.MerchantId,
		RecurringID: r.PathValue("uuid"),
		UpdatedBy:   merchantAuth.MerchantId,
	}
	if err := uuid.Validate(request.RecurringID); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, constant.NewErrInvalidFieldFmt("recurringId"))
		return
	}
	httputil.BindSubmerchantID(r, &request.MerchantID)

	if err := h.service.Cancel(ctx, request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, model.CancelRecurringContractResponse{
		RecurringID: request.RecurringID,
		Status:      constant.RecurringContractStatusInactive,
	})
}
