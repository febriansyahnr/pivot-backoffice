package withdrawalController

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Create		godoc
// @Summary		Endpoint for withdrawal process
// @Description	Endpoint for withdrawal process
// @ID			withdrawal-process
// @Tags		API - Withdrawal
// @Accept		json
// @Produce		json
// @Param		Request	body		withdrawal.WithdrawalRequest true "JSON Body for withdrawal request"
// @Success		200  	{object}	response.ApiResponse{data=withdrawal.WithdrawalProcessResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/withdrawals [post]
// @Security 	Bearer
func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/withdrawal/Create")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := withdrawal.WithdrawalRequest{
		UserId:     user.UUID,
		MerchantId: user.MerchantId,
		Type:       constant.WithdrawalManual,
		Source:     constant.SourceDashboard,
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
	if err := h.userSvc.CheckCurrentPin(ctx, request.UserId, string(rawPIN)); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if resp, err := h.service.Create(ctx, &request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
