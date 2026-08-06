package merchant

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// TransactionConfig	godoc
// @Summary				Endpoint for setting transaction configurations for merchants.
// @Description			Endpoint for setting transaction configurations such as max and min disbursement amount, etc.
// @ID					crm-merchant-transaction-configs
// @Tags				CRM - Merchant
// @Accept				json
// @Produce				json
// @Param 				id		path		string true "Merchant Id or Sub-Merchant Id"
// @Param				Request	body		merchant.TransactionConfigs true "JSON Body for transaction configs per merchant"
// @Success				200  	{object}	response.Response{data=merchant.TransactionConfigResp}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/merchants/{id}/transaction-configs [patch]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) TransactionConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/TransactionConfig")
	defer segment.End()

	merchantId := r.PathValue("id")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	var request merchantModel.TransactionConfigs
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("parser: %v", err)))
		return
	}

	if err := c.merchantSvc.TransactionConfig(ctx, merchantId, &request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, merchantModel.TransactionConfigResp{
			MerchantId:         merchantId,
			TransactionConfigs: request,
		})
	}
}

// FDSConfig godoc
// @Summary      Update FDS configuration for merchant
// @Description  Endpoint for updating FDS (Fraud Detection System) configuration for a specific merchant
// @ID			 crm-merchant-fds-configs
// @Tags         CRM - Merchant Config
// @Accept       json
// @Produce      json
// @Param        id path string true "Merchant UUID"
// @Param        request body merchant.FDSConfigRequest true "FDS Configuration"
// @Success      200 {object} response.Response{data=merchant.FDSConfigResponse}
// @Failure      400 {object} response.Response
// @Failure      422 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /crm/v1/merchants/{id}/fds-configs [put]
// @Header       all {string} X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) FDSConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/FDSConfig")
	defer segment.End()

	merchantID := r.PathValue("id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	payload := merchantModel.FDSConfigRequest{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.StructCtx(ctx, payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if resp, err := c.merchantSvc.FDSConfig(ctx, merchantID, payload); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}

// GetFDSConfig godoc
// @Summary      Get merchant FDS configuration
// @Description  Endpoint for get merchant FDS configuration
// @ID			 crm-get-merchant-fds-configs
// @Tags         CRM - Merchant Config
// @Accept       json
// @Produce      json
// @Param        id path string true "Merchant UUID"
// @Success      200 {object} response.Response{data=merchant.GetFDSConfigResponse}
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /crm/v1/merchants/{id}/fds-configs [get]
// @Header       all {string} X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) GetFDSConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/GetFDSConfig")
	defer segment.End()

	merchantID := r.PathValue("id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	config, err := c.merchantSvc.GetFDSConfig(ctx, merchantID)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, config)
}

// GetTransactionConfig	godoc
// @Summary				Endpoint for get setting transaction configurations for merchants.
// @Description			Endpoint for get setting transaction configurations such as max and min disbursement amount, etc.
// @ID					crm-merchant-get-transaction-configs
// @Tags				CRM - Merchant
// @Accept				json
// @Produce				json
// @Param 				id		path		string true "Merchant Id or Sub-Merchant Id"
// @Success				200  	{object}	response.Response{data=merchant.TransactionConfigResp}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/merchants/{id}/transaction-configs [get]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) GetTransactionConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/GetTransactionConfig")
	defer segment.End()

	merchantId := r.PathValue("id")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	if resp, err := c.merchantSvc.GetTransactionConfig(ctx, merchantId); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}

// UpdateSettlementConfig		godoc
// @Summary				Endpoint for setting settlement configurations for merchants fee.
// @Description			Endpoint for setting settlement configurations for merchants fee.
// @ID					crm-merchant-fee-settlement-configs
// @Tags				CRM - Merchant
// @Accept				json
// @Produce				json
// @Param				Request	body		merchant.SettlementConfig true "JSON Body for settlement configs per merchant fee"
// @Success				200  	{object}	response.Response{data=merchant.SettlementConfig}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/merchants/fee/{id}/settlement-configs [patch]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) UpdateSettlementConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/UpdateSettlementConfig")
	defer segment.End()

	id := r.PathValue("id")
	if err := uuid.Validate(id); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantFeeIDNotValid))
		return
	}

	var request merchantModel.SettlementConfig
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("parser: %v", err)))
		return
	}

	if err := c.validate.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	// Validate input
	if err := request.ValidateRequest(); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.merchantSvc.UpdateSettlementConfig(ctx, id, &request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, request)
	}
}

// PaymentInvestigationConfig godoc
// @Summary      Update payment investigation config for merchant
// @Description  API endpoint for updating merchant payment investigation configuration.
// @ID			 crm-merchant-payment-investigation-config
// @Tags         CRM - Merchant Config
// @Accept       json
// @Produce      json
// @Param        id path string true "Merchant UUID"
// @Param        request body merchant.PaymentInvestigationConfigRequest true "Payment investigation configuration"
// @Success      200 {object} response.Response{data=merchant.PaymentInvestigationConfigResponse}
// @Failure      400 {object} response.Response
// @Failure      422 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /crm/v1/merchants/{id}/payment-investigation-configs [patch]
// @Header       all {string} X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) PaymentInvestigationConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/PaymentInvestigationConfig")
	defer segment.End()

	merchantID := r.PathValue("id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	payload := merchantModel.PaymentInvestigationConfigRequest{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.StructCtx(ctx, payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := c.merchantSvc.PaymentInvestigationConfig(ctx, merchantID, payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, result)
}
