package withdrawalController

import (
	"errors"
	"net/http"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// GetById		godoc
// @Summary		Endpoint to view withdrawal details
// @Description	Endpoint to view withdrawal details
// @ID			withdrawal-get-detail
// @Tags		API - Withdrawal
// @Accept		json
// @Produce		json
// @Param 		account	path		string true "Account name for payments or payouts"
// @Param 		id		path		string true "Withdrawal ID"
// @Success		200  	{object}	response.ApiResponse{data=[]withdrawal.WithdrawalDetailResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/withdrawals/{account}/{id} [get]
// @Security 	Bearer
func (h *handler) GetById(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/withdrawal/GetList")
	defer segment.End()

	account := r.PathValue("account")
	if !slices.Contains(pathAccountNames, account) {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrNotFound, constant.ErrInvalidPath))
		return
	}

	id := r.PathValue("id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid transaction id format")))
		return
	}

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := &withdrawal.WithdrawalDetailRequest{
		Id:          id,
		MerchantId:  user.MerchantId,
		AccountName: constant.TypeDisbursement,
	}
	if account == "payments" {
		request.AccountName = constant.TypePayment
	}
	if resp, err := h.service.GetById(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
