package crmPaymentMethodController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetByMerchant		godoc
// @Summary				Get payment method by merchant ID
// @Description			Get payment method by merchant ID
// @ID					crm-get-payment-method-by-merchant-id
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id			path		string true "ID merchant"
// @Param       		category	query     	string  false  "filter category"
// @Param       		type		query     	string  false  "filter type"
// @Param       		acquirer	query     	string  false  "filter acquirer"
// @Success				200  		{object}	response.Response{data=[]paymentModel.PaymentMethodWithPivot}
// @Failure				500  		{object}	response.Response
// @Router				/crm/v1/merchants/{id}/payment-methods [get]
// @Header       		all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetByMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/paymentMethod/GetByMerchant")
	defer segment.End()

	merchantID := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	category := r.URL.Query().Get("category")
	paymentMethodType := r.URL.Query().Get("type")
	paymentMethodSubtype := r.URL.Query().Get("subtype")
	acquirer := r.URL.Query().Get("acquirer")

	// call service
	resp, err := h.paymentMethodSvc.GetPaymentMethodByMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantID,
		Category:   category,
		Type:       paymentMethodType,
		Subtype:    paymentMethodSubtype,
		Acquirer:   acquirer,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}

// GetStaticVAByMerchant		godoc
// @Summary				Get static va payment method by merchant ID
// @Description			Get static va  payment method by merchant ID
// @ID					crm-get-static-va-payment-method-by-merchant-id
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id			path		string true "ID merchant"
// @Param       		category	query     	string  false  "filter category"
// @Param       		type		query     	string  false  "filter type"
// @Param       		acquirer	query     	string  false  "filter acquirer"
// @Success				200  		{object}	response.Response{data=[]paymentModel.PaymentMethodWithPivot}
// @Failure				500  		{object}	response.Response
// @Router				/crm/v1/merchants/{id}/payment-methods/static-va [get]
// @Header       		all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetStaticVAByMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/paymentMethod/GetStaticVAByMerchant")
	defer segment.End()

	merchantID := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}
	acquirer := r.URL.Query().Get("acquirer")

	// call service
	resp, err := h.paymentMethodSvc.GetStaticVAPaymentMethodByMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantID,
		Acquirer:   acquirer,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}

// GetStaticQRByMerchant		godoc
// @Summary				Get static qr payment method by merchant ID
// @Description			Get static qr  payment method by merchant ID
// @ID					crm-get-static-qr-payment-method-by-merchant-id
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id			path		string true "ID merchant"
// @Param       		category	query     	string  false  "filter category"
// @Param       		type		query     	string  false  "filter type"
// @Param       		acquirer	query     	string  false  "filter acquirer"
// @Success				200  		{object}	response.Response{data=[]paymentModel.PaymentMethodWithPivot}
// @Failure				500  		{object}	response.Response
// @Router				/crm/v1/merchants/{id}/payment-methods/static-qris [get]
// @Header       		all     	{string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetStaticQRByMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/paymentMethod/GetStaticQRByMerchant")
	defer segment.End()

	merchantID := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	// call service
	resp, err := h.paymentMethodSvc.GetStaticQRPaymentMethodByMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: merchantID,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}
