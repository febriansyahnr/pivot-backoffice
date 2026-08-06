package ledgerController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *LedgerController) Create(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		req ledger_model.CreateNewLedgerEntryRequest
		err error
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v2/ledger/Create")
	defer segment.End()

	merchantId := r.Header.Get(constant.HeaderXMerchantId)
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	if err = c.validator.Struct(req); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	err = c.ledgerSvc.RecordTransaction(ctx, merchantId, &req)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, nil)
}
