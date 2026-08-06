package withdrawalController

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	withdrawalModel "github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// TransferBalance		godoc
// @Summary		Endpoint for withdrawal transfer balance
// @Description	Endpoint for withdrawal transfer balance
// @ID			withdrawal-transfer-balance
// @Tags		API - Withdrawal
// @Accept		json
// @Produce		json
// @Param		Request	body		withdrawal.WithdrawalTransferBalanceRequest true "JSON Body for withdrawal transfer balance request"
// @Success		200  	{object}	response.ApiResponse{data=withdrawal.WithdrawalTransferBalanceResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/withdrawals/balance [post]
// @Security 	Bearer
func (h *handler) TransferBalance(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/withdrawal/TransferBalance")
	defer segment.End()

	userClaims, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := withdrawalModel.WithdrawalTransferBalanceRequest{
		UserID:     userClaims.UUID,
		MerchantID: userClaims.MerchantId,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.StructCtx(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	rawPIN, _ := base64.StdEncoding.DecodeString(r.Header.Get(constant.HeaderXRequestPIN))
	if err := h.userSvc.CheckCurrentPin(ctx, userClaims.UUID, string(rawPIN)); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if resp, err := h.service.TransferBalance(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
