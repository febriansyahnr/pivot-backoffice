package merchant

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UploadReservedShortName	godoc
// @Summary		Upload reserved merchant short names.
// @Description	Allows uploading a file containing reserved merchant short names.
// @ID			crm-merchant-upload-reserved-short-name
// @Tags		CRM - Merchant
// @Accept		multipart/form-data
// @Produce		json
// @Param		file	formData	file	true	"File containing reserved short names"
// @Success		200	{object}	response.Response{data=bool}
// @Failure		500	{object}	response.Response
// @Router		/crm/v1/merchants/reserved-short-names/upload [post]
// @Header      all     {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) UploadReservedShortName(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/UploadReservedShortName")
	defer segment.End()

	file, fileHeader, err := r.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if file != nil {
		defer file.Close()
	}

	err = c.merchantSvc.UploadReservedShortName(ctx, &merchant.ReservedMerchantShortNameRequest{
		File: fileHeader,
	})
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, true)
}
