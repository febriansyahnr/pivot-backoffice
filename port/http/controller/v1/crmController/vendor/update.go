package vendor

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"
)

func (c *CRMVendorController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/vendor/Update")
	defer segment.End()

	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidId))
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	payload := vendorModel.UpdateVendorRequest{
		UUID: id,
	}

	if name := r.FormValue("name"); name != "" {
		payload.Name = util.ValueToPtr(name)
	}
	if beneficialOwner := r.FormValue("beneficialOwner"); beneficialOwner != "" {
		payload.BeneficialOwner = util.ValueToPtr(beneficialOwner)
	}
	if businessCategory := r.FormValue("businessCategory"); businessCategory != "" {
		payload.BusinessCategory = util.ValueToPtr(businessCategory)
	}
	if avgMonthlyTpvAmountStr := r.FormValue("avgMonthlyTpvAmount"); avgMonthlyTpvAmountStr != "" {
		amount, err := decimal.NewFromString(avgMonthlyTpvAmountStr)
		if err != nil {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidAvgMonthlyTpvAmount))
			return
		}
		payload.AvgMonthlyTpvAmount = &amount
	}
	if bankName := r.FormValue("bankName"); bankName != "" {
		payload.BankName = util.ValueToPtr(bankName)
	}
	if bankCode := r.FormValue("bankCode"); bankCode != "" {
		payload.BankCode = util.ValueToPtr(bankCode)
	}
	if accountNumber := r.FormValue("accountNumber"); accountNumber != "" {
		payload.AccountNumber = util.ValueToPtr(accountNumber)
	}
	if accountName := r.FormValue("accountName"); accountName != "" {
		payload.AccountName = util.ValueToPtr(accountName)
	}

	// Get document files
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		payload.DocumentFiles = r.MultipartForm.File["documents"]
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	vendor, err := c.vendorService.Update(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, vendor)
}
