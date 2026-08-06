package merchant

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UploadDocument	godoc
// @Summary			Endpoint for uploading merchant document
// @Description		Endpoint for uploading merchant document
// @ID				merchant-upload-document
// @Tags			API - Merchant Document
// @Accept			multipart/form-data
// @Produce			multipart/form-data
// @Param			id		path		string						true	"merchant id"
// @Param			Request	body		merchant.UploadDocumentReq 	true 	"JSON Body for Update Merchant fee"
// @Success			200  	{object}	response.Response{data=map[string]string}
// @Failure			500  	{object}	response.Response
// @Router			/crm/v1/merchants/{id}/documents [post]
// @Header       	all     {string}  X-CRM-Key "{"key": "value"}"
func (c *MerchantController) UploadDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/UploadDocument")
	defer segment.End()

	file, fileHeader, err := r.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if file != nil {
		defer file.Close()
	}

	document := &merchant.UploadDocumentReq{
		MerchantId: r.PathValue("id"),
		Type:       r.PostFormValue("type"),
		Identifier: r.PostFormValue("identifier"),
		CreatedBy:  r.PostFormValue("createdBy"),
		File:       fileHeader,
		Hash:       util.IOToSHA256(file),
		Notes:      r.PostFormValue("notes"),
	}

	// handle previous payload request
	if document.Identifier == "" {
		document.Identifier = r.PostFormValue("number")
	}

	if err = c.validate.Struct(document); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if id, err := c.merchantSvc.UploadDocument(ctx, document); err != nil {
		response.SendGeneralResponseError(w, err)

	} else {
		response.SendGeneralResponseOK(w, &merchant.UploadDocumentResp{Id: id})
	}
}

// GetDocuments handles the HTTP request to retrieve a list of documents for a specific merchant.
// It parses the request parameters, validates the MerchantID, and interacts with the merchant service
// to fetch the documents. The response is sent back to the client with pagination metadata.
func (c *MerchantController) GetDocuments(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchant/GetDocuments")
	defer segment.End()

	request, err := c.ParseMerchantDocumentFilterParam(r)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	request.MerchantID = r.PathValue("id")

	if err := c.validate.Var(request.MerchantID, "required,uuid"); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	result, err := c.merchantSvc.GetDocuments(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (c *MerchantController) ParseMerchantDocumentFilterParam(r *http.Request) (merchant.MerchantDocumentFilterRequest, error) {
	var (
		opt merchant.MerchantDocumentFilterRequest
		err error
	)
	opt.Page = 1
	opt.PerPage = 10
	opt.Sort = "ASC"
	opt.SortBy = "createdAt"

	if r.URL.Query().Get("page") != "" {
		opt.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return opt, pkgErrs.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			return opt, pkgErrs.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("startCreatedAt") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startCreatedAt"))
		if err != nil {
			return opt, pkgErrs.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartCreatedAt = &d
	}

	if r.URL.Query().Get("endCreatedAt") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endCreatedAt"))
		if err != nil {
			return opt, pkgErrs.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndCreatedAt = &d
	}

	if r.URL.Query().Get("sort") != "" {
		opt.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		opt.SortBy = r.URL.Query().Get("sortBy")
	}

	opt.DocumentType = r.URL.Query().Get("documentType")
	opt.Identifier = r.URL.Query().Get("keyword")
	opt.DocumentID = r.URL.Query().Get("documentID")

	return opt, nil
}
