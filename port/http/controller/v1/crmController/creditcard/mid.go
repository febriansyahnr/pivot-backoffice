package crmCreditcardController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetMIDList(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		err     error
		payload = creditcardModel.GetMIDListRequest{}
	)

	ctx, span := otelTracer.Start(ctx, "port/http/controller/v1/crmController/creditcard/GetMIDList")
	defer span.End()

	query := r.URL.Query()
	payload.Mid = query.Get("mid")
	payload.Acquirer = query.Get("acquirer")
	payload.Name = query.Get("name")
	payload.Type = query.Get("type")
	payload.TransactionType = query.Get("transactionType")
	payload.InstallmentType = query.Get("installmentType")
	isActive := query.Get("isActive")
	switch isActive {
	case "true":
		payload.IsActive = util.BoolPtr(true)
	case "false":
		payload.IsActive = util.BoolPtr(false)
	}

	isDefault := query.Get("isDefault")
	switch isDefault {
	case "true":
		payload.IsDefault = util.BoolPtr(true)
	case "false":
		payload.IsDefault = util.BoolPtr(false)
	}

	payload.Page = constant.DefaultPage
	payload.PerPage = constant.DefaultPaginationPageSize

	pageStr := query.Get("page")
	if pageStr != "" {
		payload.Page, err = strconv.Atoi(pageStr)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPageStr := query.Get("perPage")
	if perPageStr != "" {
		payload.PerPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	resp, err := h.creditcardSvc.GetMIDList(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, resp.Data, resp.Meta)
}

func (h *handler) GetMIDMapList(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		err error
	)

	ctx, span := otelTracer.Start(ctx, "port/http/controller/v1/crmController/creditcard/GetMIDMapList")
	defer span.End()

	query := r.URL.Query()
	page := constant.DefaultPage
	perPage := constant.DefaultPaginationPageSize
	merchantId := query.Get("merchantId")

	pageStr := query.Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPageStr := query.Get("perPage")
	if perPageStr != "" {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	resp, err := h.creditcardSvc.GetMIDMapList(ctx, perPage, page, merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, resp.Data, resp.Meta)
}

func (h *handler) CreateMID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/creditcard/CreateMID")
	defer span.End()

	var request creditcardModel.CreateMIDRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	err := h.creditcardSvc.CreateMID(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]bool{"created": true})
}

func (h *handler) UpdateMID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/creditcard/UpdateMID")
	defer span.End()

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	request := creditcardModel.UpdateMIDRequest{
		UUID: id,
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	err := h.creditcardSvc.UpdateMID(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]bool{"updated": true})
}
