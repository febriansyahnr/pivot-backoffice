package internalAccountInquiry

import (
	"context"
	"encoding/json"
	"net/http"

	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (ctrl *AccountInquiryController) RequestAccountInquiry(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/accountInquiry/RequestAccountInquiry")
	defer segment.End()

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	var req requestAccountInquiries.RequestAccountInquiriesHttpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	req.MerchantID = merchant.MerchantId
	req.ParentMerchantID = merchant.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		req.MerchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.MerchantId)
	}

	req.ChannelInformation.AccountNumber = strings.ReplaceAll(req.ChannelInformation.AccountNumber, " ", "")

	if err := ctrl.validator.Var(req.ChannelInformation.AccountNumber, "required,numeric"); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := ctrl.validator.Struct(req); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	request, err := ctrl.service.RequestAccountInquiry(ctx, req)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, request)
}
