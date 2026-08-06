package crmXbController

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetListConfigSpread	godoc
// @Summary				Get List Config Spread of XB Payout
// @Description			Get List Config Spread of XB Payout
// @ID					get-list-config-spread-of-xb-payout
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Success				200  	{object}	response.Response
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/xb/config/spread/list [get]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetListConfigSpread(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/xb/GetListConfigSpread")
	defer segment.End()

	var err error
	requestPayload := xbModel.GetListConfigSpreadRequest{
		Page:    constant.DefaultPage,
		PerPage: constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	merchantId := query.Get("merchantId")
	if merchantId != "" {
		requestPayload.MerchantID, err = uuid.Parse(merchantId)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidMerchantId))
			return
		}
	}

	resp, err := h.XbPayoutSvc.GetListConfigSpread(ctx, &requestPayload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

// GetConfigSpreadDetailByID	godoc
// @Summary				Get Detail Config Spread of XB Payout
// @Description			Get Detail Config Spread of XB Payout
// @ID					get-detail-config-spread-of-xb-payout
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id		path		string true "XB Config Spread ID"
// @Success				200  	{object}	response.Response{data=xbModel.GetConfigSpreadDetailResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/xb/config/spread/{id} [get]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) GetConfigSpreadDetailByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/xb/GetConfigSpreadDetailByID")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	resp, err := h.XbPayoutSvc.GetConfigSpreadDetailByID(ctx, id)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, resp)
}

// CreateConfigSpread	godoc
// @Summary				Create Config Spread of XB Payout
// @Description			Create Config Spread of XB Payout
// @ID					create-config-spread-of-xb-payout
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				Request	body		xbModel.CreateConfigSpreadRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=xbModel.CreateConfigSpreadResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/xb/config/spread [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) CreateConfigSpread(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/xb/CreateConfigSpread")
	defer segment.End()

	requestPayload := xbModel.CreateConfigSpreadRequest{}
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(requestPayload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := h.XbPayoutSvc.CreateConfigSpread(ctx, &requestPayload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, resp)
}

// UpdateConfigSpread	godoc
// @Summary				Update Config Spread of XB Payout
// @Description			Update Config Spread of XB Payout
// @ID					update-config-spread-of-xb-payout
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				id		path		string true "XB Config Spread ID"
// @Param 				Request	body		xbModel.UpdateConfigSpreadRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=xbModel.UpdateConfigSpreadResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/xb/config/spread/{id} [put]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) UpdateConfigSpread(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/xb/UpdateConfigSpread")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	configSpreadUUID, _ := uuid.Parse(id)
	requestPayload := xbModel.UpdateConfigSpreadRequest{
		UUID: configSpreadUUID,
	}
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(requestPayload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := h.XbPayoutSvc.UpdateConfigSpread(ctx, &requestPayload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, resp)
}
