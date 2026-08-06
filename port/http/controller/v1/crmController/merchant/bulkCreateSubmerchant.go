package merchant

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMMerchantController) BulkCreateSubmerchant(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/BulkCreateSubmerchant")
	defer segment.End()

	var payload merchantModel.BulkCreateSubMerchantRequest

	payload.MerchantId = r.FormValue("merchantId")
	if _, err := uuid.Parse(payload.MerchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}
	payload.KYCType = strings.ToUpper(r.FormValue("kycType"))
	if payload.KYCType != constant.MerchantKYCTypeKYC && payload.KYCType != constant.MerchantKYCTypeNonKYC {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrIncorrectKYCType))
		return
	}

	payload.IsInvitePIC = r.FormValue("isInvitePIC") == "true"

	file, fileHeader, err := util.ValidateFile(r, util.ValidateFileFormParams{
		FieldName: "file",
		FileSize:  constant.FileSize5MB,
		Extension: constant.FileExtensionCsv,
	})
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	payload.FileName = fileHeader.Filename
	defer file.Close()

	payload.SubmerchantDetails, err = util.ReadCsvFile(file, util.ReadCsvFileParams{
		IgnoreFirstRow:   true,
		Delimiter:        ',',
		LazyQuotes:       true,
		TrimLeadingSpace: true,
	})
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	if len(payload.SubmerchantDetails) == 0 {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrBulkCreateSubMerchantNoData))
		return
	}

	result, err := c.merchantSvc.BulkCreateSubMerchant(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, result)
}

func (c *CRMMerchantController) GetBulkCreateSubmerchantSummary(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/GetBulkCreateSubmerchantSummary")
	defer segment.End()

	var request merchantModel.GetBulkCreateSubMerchantSummaryRequest

	merchantId := r.PathValue("merchantId")
	if _, err := uuid.Parse(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}
	request.MerchantId = merchantId

	sessionId := r.PathValue("sessionId")
	if _, err := uuid.Parse(sessionId); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("invalid id %s", sessionId)))
		return
	}
	request.ID = sessionId

	result, err := c.merchantSvc.GetBulkCreateSubMerchantSummary(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, result)
}
