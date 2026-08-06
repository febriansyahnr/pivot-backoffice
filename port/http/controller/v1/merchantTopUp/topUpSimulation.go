package merchantTopUp

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/topUpSimulation"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Topup			godoc
// @Summary			Merchant Top-up Simulation
// @Description		Merchant Top-up Simulation
// @ID				api-merchant-top-up-simulation
// @Tags			API - Merchant Top Up
// @Accept			json
// @Produce			json
// @Param			Request	body		snapCoreTopUpSimulationModel.TopupSimulationRequest true "JSON Body for Merchant Simulation Topup"
// @Success			200  	{object}	response.Response{data=merchantTopUp.MerchantTopUpResponse}
// @Failure			500  	{object}	response.Response
// @Router			/api/v1/{usecase}/top-up-simulation-va [post]
// @Security		Bearer
func (c *handler) TopUpSimulation(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchantTopUp/TopUpSimulation")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload model.TopupSimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	payload.MerchantId = user.MerchantId
	payload.AccountName, _ = ctx.Value(constant.CtxAccountName).(string)

	if err := c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	reference, err := c.service.CreateTopupSimulation(ctx, payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, reference)
}
