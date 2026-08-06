package internalAccountInquiry

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (ctrl *AccountInquiryController) CheckStatusRequestInquiry(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/accountInquiry/CheckStatusRequestInquiry")
	defer span.End()

	inquiryID := chi.URLParam(r, "inquiryId")
	_, err := uuid.Parse(inquiryID)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInquiryIdNotFound))
		return
	}

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	merchantID := merchant.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.MerchantId)
	}

	request, err := ctrl.service.CheckStatusRequestInquiry(ctx, merchantID, inquiryID)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, request)
}
