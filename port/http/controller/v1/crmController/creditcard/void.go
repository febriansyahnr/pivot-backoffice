package crmCreditcardController

import (
	"encoding/json"
	"net/http"

	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Void		godoc
// @Summary				Void creditcard transaction
// @Description			Void creditcard transaction
// @ID					void-creditcard-transaction
// @Tags				API - CRM
// @Accept				mpfd
// @Produce				mpfd
// @Param 				Request	body		card.VoidRequest true "Form Body for Send"
// @Success				200  	{object}	response.Response{data=card.VoidRequest}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/creditcard/void [post]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (h *handler) Void(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/crmController/creditcard/Void")
	defer segment.End()

	var request creditcardModel.VoidRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	void, err := h.creditcardSvc.Void(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, void)
}
