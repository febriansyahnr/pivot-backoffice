package withdrawalCrmController

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

func (h *handler) ChangeStatusWithdrawal(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/withdrawal/ChangeStatusWithdrawal")
	defer segment.End()

	withdrawalID := r.PathValue("id")
	if err := uuid.Validate(withdrawalID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, fmt.Errorf("id is required")))
		return
	}

	request := &withdrawal.WithdrawalChangeStatusRequest{
		WithdrawalID: withdrawalID,
	}

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validator.StructCtx(ctx, request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	trx, err := h.service.ChangeStatusWithdrawal(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, trx)
}
