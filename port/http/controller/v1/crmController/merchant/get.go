package merchant

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Get		godoc
// @Summary		Get merchant by ID endpoint
// @Description	Get merchant by ID endpoint
// @ID			crm-merchant-get
// @Tags		CRM - Merchant
// @Accept		json
// @Produce		json
// @Param 		merchantId	path		string true "Merchant ID"
// @Success		200  		{object}	response.Response{data=merchant.MerchantResponse}
// @Failure		500  		{object}	response.Response
// @Router		/crm/v1/merchants/{merchantId} [get]
// @Header      all     	{string}  X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) Get(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/Get")
	defer segment.End()

	merchantId := r.PathValue("merchantId")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	merchant, err := c.merchantSvc.FindMerchantByID(ctx, merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	if merchant == nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound))
		return
	}

	response.SendGeneralResponseOK(w, merchant.ToCRMMerchantResponse())
}
