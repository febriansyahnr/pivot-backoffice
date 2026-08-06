package adjustment

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// CreateManualTopup	godoc
// @Summary				Top up merchant balance manually (via bank transfer manually)
// @Description			Top up merchant balance manually (via bank transfer manually)
// @ID					crm-balance-topup-manual
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				Request	body		adjustment.ManualTopupRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=adjustment.ManualTopupResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/balances/topup/manual [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) CreateManualTopup(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/balance/CreateManualTopup")
	defer segment.End()

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB

	var (
		err error
		f   multipart.File
	)

	request := adjustModel.ManualTopupRequest{
		MerchantID:  r.PostFormValue("merchant_id"),
		BankRefID:   r.PostFormValue("bank_reference_id"),
		BankName:    r.PostFormValue("bank_name"),
		BankAccount: r.PostFormValue("bank_account"),
		Currency:    r.PostFormValue("currency"),
		CreatedBy:   r.PostFormValue("created_by"),
		Notes:       r.PostFormValue("notes"),
	}
	if request.Amount, err = strconv.ParseFloat(r.PostFormValue("amount"), 64); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf(constant.ErrDetailMsgRequestFormatField, "amount")))
		return
	}
	if request.SendCallback, err = util.PostFormBoolValue(r, "send_callback"); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf(constant.ErrDetailMsgRequestFormatField, "send_callback")))
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if f, request.File, err = r.FormFile("proof_of_transfer"); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}
	defer f.Close()

	if !strings.Contains(".jpg .jpeg .png", filepath.Ext(request.File.Filename)) {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, errors.New("transfer proof format is not supported")))
		return
	}

	id, err := h.service.CreateManualTopup(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}
	response.SendGeneralResponseOK(w, adjustModel.ManualTopupResponse{ID: id})
}
