package crmCreditcardController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// BlockCard		godoc
// @Summary		Block creditcard
// @Description	Block creditcard
// @ID			block-creditcard
// @Tags		API - CRM
// @Accept		json
// @Produce		json
// @Param 		Request	body		card.BlockCardRequest true "JSON Body for Block Card"
// @Success		200  	{object}	response.Response{data=card.BlockCardResponse}
// @Failure		400  	{object}	response.Response
// @Failure		500  	{object}	response.Response
// @Router		/crm/v1/card/block [put]
// @Header       all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) BlockCard(w http.ResponseWriter, r *http.Request) {

	var (
		ctx = r.Context()
		req creditcardModel.BlockCardRequest
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/creditcard/BlockCard")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidValidation))
		return
	}

	err := h.creditcardSvc.BlockCard(ctx, &req)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	resp := &creditcardModel.BlockCardResponse{
		Success: true,
		Message: "Card blocked successfully",
	}

	response.SendGeneralResponseOK(w, resp)
}