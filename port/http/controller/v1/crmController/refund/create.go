package v1CRMRefundController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *refundController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/refund/Create")
	defer span.End()

	var (
		resp *refundModel.RefundResponse
		err  error
	)

	payload := refundModel.NewCreatRefundThroughCRMRequest()
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err = c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	payload.CreateRefundRequest.MerchantID = payload.MerchantID
	payload.CreateRefundRequest.Method = payload.Method
	payload.CreateRefundRequest.TransferDestination = nil
	if resp, err = c.refundSvc.Create(ctx, &payload.CreateRefundRequest); err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}
