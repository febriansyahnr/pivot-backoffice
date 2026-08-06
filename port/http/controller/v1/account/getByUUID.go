package accountController

import (
	errs "errors"
	"net/http"

	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (b *account) GetByUUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/accounts/GetByUUID")
	defer segment.End()

	// Get Account By UUID
	accountId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	account, err := b.accountSvc.GetAccount(ctx, accountId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if account == nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, errs.New("data not found")))
		return
	}

	response.SendApiResponseOK(w, account.ToResponse())
}
