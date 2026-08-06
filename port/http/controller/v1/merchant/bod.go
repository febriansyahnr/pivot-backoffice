package merchant

import (
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// CreateMerchantBOD godoc
// @Summary			 Endpoint for create data Board Of Director (BOD)
// @Description		 Endpoint for create data Board Of Director (BOD)
// @ID				 merchant-create-bod
// @Tags			 API - Merchant Board Of Directors
// @Accept			 multipart/form-data
// @Produce			 multipart/form-data
// @Param			 id			path		string								true	"Merchant Id"
// @Param			 Request	body		merchant.UpsertBoardOfDirectorReq 	true 	"JSON Body for create board of director data"
// @Success			 200  		{object}	response.Response{data=merchant.UpsertBoardOfDirectorResp}
// @Failure			 500  		{object}	response.Response
// @Router			 /crm/v1/merchants/{id}/bods [post]
// @Header       	 all     {string}  X-CRM-Key "{"key": "value"}"
func (h *MerchantController) CreateMerchantBOD(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/CreateMerchantBOD")
	defer segment.End()

	file, fileHeader, err := r.FormFile("identityFile")
	if err != nil && err != http.ErrMissingFile {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if fileHeader != nil {
		defer file.Close()
	}

	request := &merchant.UpsertBoardOfDirectorReq{
		Method:         r.Method,
		MerchantId:     r.PathValue("id"),
		Position:       r.PostFormValue("position"),
		Name:           r.PostFormValue("name"),
		IdentityNumber: r.PostFormValue("identityNumber"),
		IdentityFile:   fileHeader,
		Hash:           util.IOToSHA256(file),
		PositionLong:   r.PostFormValue("positionLong"),
		CreatedBy:      r.PostFormValue("createdBy"),
		Shares:         r.PostFormValue("shares"),
	}
	if err := h.validate.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := request.ValidateRequest(); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if id, err := h.merchantSvc.UpsertMerchantBOD(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, merchant.UpsertBoardOfDirectorResp{Id: id})
	}
}

// GetListMerchantBOD godoc
// @Summary			  Endpoint for get list data Board Of Director (BOD)
// @Description		  Endpoint for get list data Board Of Director (BOD)
// @ID				  merchant-get-list-bod
// @Tags			  API - Merchant Board Of Directors
// @Accept			  json
// @Produce			  json
// @Param			  id	path		string	true	"Merchant Id"
// @Success			  200  	{object}	response.Response{data=merchant.ListBoardOfDirectorResp}
// @Failure			  500  	{object}	response.Response
// @Router			  /crm/v1/merchants/{id}/bods [get]
// @Header       	  all     {string}  X-CRM-Key "{"key": "value"}"
func (h *MerchantController) GetListMerchantBOD(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/GetListMerchantBOD")
	defer segment.End()

	merchantId := r.PathValue("id")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid merchant id format")))
		return
	}

	resp, err := h.merchantSvc.GetListMerchantBODs(ctx, merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, resp)
	}
}

// UpdateMerchantBOD godoc
// @Summary			 Endpoint for update data Board Of Director (BOD)
// @Description		 Endpoint for update data Board Of Director (BOD)
// @ID				 merchant-update-bod
// @Tags			 API - Merchant Board Of Directors
// @Accept			 multipart/form-data
// @Produce			 multipart/form-data
// @Param			 id			path		string								true	"Merchant Id"
// @Param			 bod_id		path		string								true	"Unique Id for Board Of Director (BOD)"
// @Param			 Request	body		merchant.UpsertBoardOfDirectorReq 	true 	"JSON Body for update board of director data"
// @Success			 200  		{object}	response.Response{data=merchant.UpsertBoardOfDirectorResp}
// @Failure			 500  		{object}	response.Response
// @Router			 /crm/v1/merchants/{id}/bods/{bod_id} [put]
// @Header       	 all     {string}  X-CRM-Key "{"key": "value"}"
func (h *MerchantController) UpdateMerchantBOD(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/UpdateMerchantBOD")
	defer segment.End()

	file, fileHeader, err := r.FormFile("identityFile")
	if err != nil && err != http.ErrMissingFile {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if fileHeader != nil {
		defer file.Close()
	}

	request := &merchant.UpsertBoardOfDirectorReq{
		Id:             r.PathValue("bod_id"),
		Method:         r.Method,
		MerchantId:     r.PathValue("id"),
		Position:       r.PostFormValue("position"),
		Name:           r.PostFormValue("name"),
		IdentityNumber: r.PostFormValue("identityNumber"),
		IdentityFile:   fileHeader,
		Hash:           util.IOToSHA256(file),
		PositionLong:   r.PostFormValue("positionLong"),
		Shares:         r.PostFormValue("shares"),
	}
	if err := h.validate.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if err := request.ValidateRequest(); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if id, err := h.merchantSvc.UpsertMerchantBOD(ctx, request); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, merchant.UpsertBoardOfDirectorResp{Id: id})
	}
}
