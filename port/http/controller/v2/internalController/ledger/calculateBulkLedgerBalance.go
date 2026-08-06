package ledgerController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *LedgerController) CalculateBulkLedgerBalance(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/ledger/CalculateBulkLedgerBalance")
	defer segment.End()

	var request account_model.CalculateBulkLedgerBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	request.MerchantID = r.Header.Get(constant.HeaderXMerchantId)

	balance, err := c.ledgerSvc.CalculateBulkLedgerBalance(ctx, &request)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, balance)
}
