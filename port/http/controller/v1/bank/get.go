package bankController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// List				godoc
// @Summary			List of all banks
// @Description		List of all banks
// @ID				api-bank-list
// @Tags			API - Bank
// @Accept			json
// @Produce			json
// @Success			200  	{object}	response.ApiResponse{data=[]bank.BankListResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/banks/ [get]
// @Security		Bearer
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "http/controller/v1/bank/List")
	defer segment.End()

	// response payload
	bankDB := bankTransfer.NewBankDB()

	response.SendApiResponseOK(w, bankDB.List())
}
