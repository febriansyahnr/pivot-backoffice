package vendor

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"
)

func (c *CRMVendorController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/vendor/Create")
	defer segment.End()

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Parse avgMonthlyTpvAmount
	avgMonthlyTpvAmount, err := decimal.NewFromString(r.FormValue("avgMonthlyTpvAmount"))
	if err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidAvgMonthlyTpvAmount))
		return
	}

	payload := vendorModel.CreateVendorRequest{
		MerchantID:          r.FormValue("merchantId"),
		Name:                r.FormValue("name"),
		BeneficialOwner:     r.FormValue("beneficialOwner"),
		BusinessCategory:    r.FormValue("businessCategory"),
		AvgMonthlyTpvAmount: avgMonthlyTpvAmount,
		BankName:            r.FormValue("bankName"),
		BankCode:            r.FormValue("bankCode"),
		AccountNumber:       r.FormValue("accountNumber"),
		AccountName:         r.FormValue("accountName"),
	}

	// Get document files
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		payload.DocumentFiles = r.MultipartForm.File["documents"]
	}

	if err := c.validate.Struct(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	vendor, err := c.vendorService.Create(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, vendor)
}
