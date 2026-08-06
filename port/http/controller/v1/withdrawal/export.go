package withdrawalController

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Export		godoc
// @Summary		Endpoint for download withdrawal history
// @Description	Endpoint for download withdrawal history
// @ID			withdrawal-download-list
// @Tags		API - Withdrawal
// @Accept		json
// @Produce		json
// @Param 		account	path		string true "Account name for payments or payouts"
// @Param		Request	body		withdrawal.WithdrawalListRequest true "JSON body for withdrawal list filter"
// @Success		200  	{object}	response.ApiResponse{data=[]withdrawal.WithdrawalDownloadResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/withdrawals/{account}/export	[post]
// @Security 	Bearer
func (h *handler) Export(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/withdrawal/Export")
	defer segment.End()

	account := r.PathValue("account")
	if !slices.Contains(pathAccountNames, account) {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrNotFound, constant.ErrInvalidPath))
		return
	}

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &withdrawal.WithdrawalListRequest{
		MerchantId: user.MerchantId,
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.preparationGetList(r, account, request); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if err := h.validator.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if res, err := h.service.Export(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, res)
	}
}
