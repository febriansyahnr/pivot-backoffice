package charges

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Export		godoc
// @Summary		Download charge history
// @Description	This endpoint is used to download charge history reports
// @ID			api-charge-download-history
// @Tags		API - charge
// @Accept		json
// @Produce		json
// @Param		Request	body		unifiedPaymentModel.FilterChargeRequest true "JSON body for charge list filter"
// @Success		200  	{object}	response.ApiResponse{data=unifiedPaymentModel.chargeDownloadHistoryResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/payments/export [post]
// @Security	Bearer
func (h *ChargesController) Export(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/Export")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &unifiedPaymentModel.FilterChargeRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	request.MerchantID = user.MerchantId

	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(request, "startCreatedAt", "endCreatedAt", "startPaymentDate", "endPaymentDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.unifiedPaymentService.ExportCharge(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
