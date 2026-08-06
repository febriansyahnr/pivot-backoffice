package merchant

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// BlockMerchant		godoc
// @Summary		Block merchant endpoint
// @Description	Block merchant and all related static VAs. If parent merchant is blocked, all sub-merchants will also be blocked.
// @ID			crm-merchant-block
// @Tags		CRM - Merchant
// @Accept		json
// @Produce		json
// @Param		id	path	string	true	"Merchant ID"
// @Success		200  	{object}	response.Response{data=merchant.BlockMerchantResponse}
// @Failure		400  	{object}	response.Response
// @Failure		500  	{object}	response.Response
// @Router		/api/v1/merchants/{id}/block [post]
// @Security	Bearer
func (c *CRMMerchantController) BlockMerchant(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/BlockMerchant")
	defer segment.End()

	merchantId := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	blockResponse, err := c.merchantSvc.BlockMerchant(ctx, merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, blockResponse)
}

func (c *CRMMerchantController) UnblockMerchant(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/UnblockMerchant")
	defer segment.End()

	merchantId := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	unblockResponse, err := c.merchantSvc.UnblockMerchant(ctx, merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, unblockResponse)
}
