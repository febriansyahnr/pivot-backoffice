package reconciliation

import (
	"net/http"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// UploadFile implements controller.V1ReconciliationController.
func (c *ReconciliationController) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/reconciliation/UploadFile")
	defer segment.End()
	identifier := r.Header.Get(constant.XIdentifierKey)
	if identifier == "" {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, constant.ErrMissingXIdentifier))
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, err))
		return
	}
	defer file.Close()

	transactionType := r.FormValue("transactionType")
	if transactionType == "" {
		transactionType = constant.TypePayment
	}
	if !slices.Contains([]string{
		constant.TypeDisbursement,
		constant.TypePayment,
		constant.TypeWithdrawal,
		constant.TypeRefund,
	}, transactionType) {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, constant.ErrReconTransactionTypeInvalid))
		return
	}

	uuid, err := c.reconSvc.UploadFile(ctx, transactionType, identifier, file, fileHeader)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, map[string]string{"message": "success upload file", "uuid": *uuid})
}
