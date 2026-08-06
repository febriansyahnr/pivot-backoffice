package merchantRcn

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (m *MerchantRcnController) InquiryTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/internalController/v1/merchantRCN/InquiryTransactions")
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)
	defer segment.End()

	var request vccSettlement.VccTransactionInquiryRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	merchantId := r.Header.Get(constant.HeaderXMerchantId)
	if merchantId == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrIncorrectMerchantID))
		return
	}
	request.MerchantId = merchantId

	err = m.validate.Struct(request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	data, err := m.vccSettlementSvc.RcnTransactionInquiry(ctx, &request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, data)
}
