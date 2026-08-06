package merchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// CreateFeeOnBehalf	godoc
// @Summary				Create transaction fee config on behalf of sub-merchant
// @Description			Create transaction fee config on behalf of sub-merchant
// @ID					crm-merchant-create-fee-on-behalf
// @Tags				CRM - Merchant
// @Accept				json
// @Produce				json
// @Param				Request	body		merchant.CreateFeeConfigOnBehalfRequest true "JSON Body for create fee config on behalf"
// @Success				200  	{object}	response.Response{data=merchant.CreateFeeConfigOnBehalfResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/merchants/fee-on-behalf [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *CRMMerchantController) CreateFeeConfigOnBehalf(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/CreateFeeConfigOnBehalf")
	defer segment.End()

	request := &merchant.CreateFeeConfigOnBehalfRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	id, err := h.merchantSvc.CreateFeeConfigOnBehalf(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, &merchant.CreateFeeConfigOnBehalfResponse{
		Id: id, CreateFeeConfigOnBehalfRequest: request,
	})
}

// GetFeeConfigOnBehalf	godoc
// @Summary				Retrieving transaction fee config on behalf of sub-merchants
// @Description			Retrieving transaction fee config on behalf of sub-merchants
// @ID					crm-merchant-get-fee-on-behalf
// @Tags				CRM - Merchant
// @Accept				json
// @Produce				json
// @Param				merchantId		query	string  true  "Main merchant ID"  Format(uuid)
// @Param				reference		query	string  true  "Reference fee config such as DISBURSEMENT or PAYMENT"
// @Param				paymentMethod	query	string  false "Specific payment method type for PAYMENT reference"
// @Success				200  	{object}	response.Response{data=merchant.GetFeeConfigOnBehalfResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/merchants/fee-on-behalf/details [get]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *CRMMerchantController) GetFeeConfigOnBehalf(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/GetFeeConfigOnBehalf")
	defer segment.End()

	request := &merchant.GetFeeConfigOnBehalfRequest{
		MerchantId: r.URL.Query().Get("merchantId"),
		Reference:  r.URL.Query().Get("reference"),
	}
	if request.Reference == constant.ReferencePayment {
		request.PaymentMethod = util.ValueToPtr(r.URL.Query().Get("paymentMethod"))
	}
	if err := uuid.Validate(request.MerchantId); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, constant.ErrInvalidMerchantId))
		return
	}
	if err := h.validate.Var(request.Reference, "required,oneof=PAYMENT DISBURSEMENT ACCOUNT_INQUIRY"); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, constant.ErrInvalidReference))
		return
	}

	configs, err := h.merchantSvc.GetFeeConfigOnBehalf(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, &merchant.GetFeeConfigOnBehalfResponse{
		GetFeeConfigOnBehalfRequest: request, Configs: configs,
	})
}

// UpdateFeeConfigOnBehalf	godoc
// @Summary					Endpoint for updating transaction fee config on behalf of sub-merchants
// @Description				Endpoint for updating transaction fee config on behalf of sub-merchants
// @ID						crm-merchant-update-fee-on-behalf
// @Tags					CRM - Merchant
// @Accept					json
// @Produce					json
// @Param 					id		path		string true "On-behalf fee ID"
// @Param					Request	body		merchant.UpdateFeeConfigOnBehalfRequest true "JSON Body for fee config"
// @Success					200  	{object}	response.Response{data=map[string]string}
// @Failure					500  	{object}	response.Response
// @Router					/crm/v1/merchants/fee-on-behalf/{id} [patch]
// @Header       			all     {string}  X-CRM-Key "{"key": "value"}"
func (h *CRMMerchantController) UpdateFeeConfigOnBehalf(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/UpdateFeeConfigOnBehalf")
	defer segment.End()

	id := r.PathValue("id")
	if err := uuid.Validate(id); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	request := &merchant.UpdateFeeConfigOnBehalfRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	if err := h.merchantSvc.UpdateFeeConfigOnBehalf(ctx, id, request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, map[string]string{"message": "data successfully updated"})
	}
}
