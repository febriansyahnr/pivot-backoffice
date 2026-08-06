package merchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UpdateFeeTieringConfig	godoc
// @Summary					Endpoint to add fee tiering configuration
// @Description				Endpoint to add fee tiering configuration
// @ID						crm-merchant-fee-tiering-config
// @Tags					CRM - Merchant
// @Accept					json
// @Produce					json
// @Param 					id		path		string true "Merchant fee ID"
// @Param					Request	body		merchant.FeeTieringRequest true "JSON Body for configuration fee tiering"
// @Success					200  	{object}	response.Response{data=merchant.FeeTieringResponse}
// @Failure					500  	{object}	response.Response
// @Router					/crm/v1/merchants/fee/{id}/tiers [patch]
// @Header       			all     {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) UpdateFeeTieringConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/UpdateFeeTieringConfig")
	defer segment.End()

	request := &merchant.FeeTieringRequest{
		FeeId: r.PathValue("id"),
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err := c.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := c.merchantSvc.UpdateFeeTieringConfig(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}
