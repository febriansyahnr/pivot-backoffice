package internalFeeController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalFeeController) CalculateWhitelabelMerchantFee(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/fee/CalculateWhitelabelMerchantFee")
	defer segment.End()

	var request feeModel.CalculateWhitelabelMerchantFeeRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	request.MerchantID = r.Header.Get(constant.HeaderXMerchantId)

	_, detail, err := c.svc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
		MerchantID:      request.RequesterMerchantID,
		Reference:       constant.ReferenceWallet,
		ReferenceType:   request.ReferenceType,
		ReferenceAmount: request.Amount,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, detail)
}
