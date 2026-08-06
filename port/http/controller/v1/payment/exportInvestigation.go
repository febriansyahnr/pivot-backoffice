package payment

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// ExportInvestigation	godoc
// @Summary		Download investigation history
// @Description	This endpoint is used to download investigation history reports
// @ID			api-cases-export
// @Tags		API - Cases Management
// @Accept		json
// @Produce		json
// @Param		Request	body		paymentModel.InvestigationDownloadHistoryRequest true "JSON body for investigation list filter"
// @Success		200  	{object}	response.ApiResponse{data=paymentModel.InvestigationDownloadHistoryResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/cases/export [post]
// @Security	Bearer
func (h *PaymentController) ExportInvestigation(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/ExportInvestigation")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &paymentModel.InvestigationDownloadHistoryRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	request.MerchantId = user.MerchantId

	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(request, "fromDate", "toDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := h.paymentService.ExportInvestigatedPayments(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)
	} else {
		response.SendApiResponseOK(w, resp)
	}
}
