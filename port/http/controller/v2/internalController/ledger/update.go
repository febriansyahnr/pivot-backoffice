package ledgerController

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *LedgerController) Update(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		req ledger_model.BulkUpdateLedgerEntryRequest
		err error
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v2/ledger/Create")
	defer segment.End()

	referenceId := chi.URLParam(r, "referenceId")
	refUuid, err := uuid.Parse(referenceId)
	if err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	req.ReferenceID = refUuid

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err = c.validator.Struct(req); err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	err = c.ledgerSvc.BulkUpdateLedgerEntry(ctx, &req)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, nil)
}
