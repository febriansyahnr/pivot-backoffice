package internalFeeController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalFeeController) CalculateWalletTransactionFee(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/fee/CalculateWalletTransactionFee")
	defer segment.End()

	var request feeModel.GetTransactionFeeRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	request.MerchantID = r.Header.Get(constant.HeaderXMerchantId)

	onbehalfDetail, err := c.svc.GetTransactionFeeOnBehalf(ctx, &feeModel.GetTrxFeeOnBehalfRequest{
		MerchantId:        request.MerchantID,
		SubMerchantId:     request.MerchantID,
		Reference:         constant.ReferenceWallet,
		ReferenceType:     request.ReferenceType,
		TransactionAmount: request.Amount,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, onbehalfDetail)
}
