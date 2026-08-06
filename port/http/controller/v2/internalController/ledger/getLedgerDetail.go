package ledgerController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *LedgerController) GetLedgerDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v2/ledger/GetLedgerDetail")
	defer segment.End()

	referenceID := chi.URLParam(r, "referenceId")
	err := c.validator.Var(referenceID, "required,uuid")
	if err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	ledgerDetail, err := c.ledgerSvc.GetLedgerDetail(ctx, referenceID)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	resp := make([]*ledger_model.GetLedgerDetailResponse, 0)
	for _, v := range ledgerDetail {
		resp = append(resp, ledger_model.ToGetLedgerDetailResponse(&v))
	}

	response.SendApiResponseOK(w, resp)
}
