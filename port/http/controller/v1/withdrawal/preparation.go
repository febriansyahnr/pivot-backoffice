package withdrawalController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Preparation	godoc
// @Summary		Endpoint for preparing a fund withdrawal request
// @Description	Endpoint for preparing a fund withdrawal request
// @ID			withdrawal-preparation
// @Tags		API - Withdrawal
// @Accept		json
// @Produce		json
// @Param       accountName	query     	string  false  "Account name"
// @Success		200  		{object}	response.ApiResponse{data=withdrawal.PreparationResponse}
// @Failure		500  		{object}	response.ApiResponse
// @Router		/api/v1/withdrawals/preparation [get]
// @Security 	Bearer
func (h *handler) Preparation(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/withdrawal/Preparation")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &withdrawal.PreparationRequest{
		MerchantId:  user.MerchantId,
		AccountName: constant.TypePayment,
	}
	if accountName := r.URL.Query().Get("accountName"); accountName != "" {
		request.AccountName = accountName
	}
	if err := h.validator.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		request.MerchantId = subMerchantId
	}

	if resp, err := h.service.Preparation(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
