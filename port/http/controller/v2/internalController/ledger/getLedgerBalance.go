package ledgerController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *LedgerController) GetLedgerBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v2/ledger/GetLedgerBalance")
	defer segment.End()

	accountID := chi.URLParam(r, "accountId")
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	balance, err := c.ledgerSvc.GetLedgerBalance(ctx, accountUUID)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, balance)
}
