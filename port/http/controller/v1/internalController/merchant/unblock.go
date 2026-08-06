package internal_merchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1InternalMerchantController) Unblock(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantForbiddenUsecase/Unblock")
	defer segment.End()

	var (
		err error
	)

	var payload merchantForbiddenUsecase.MerchantForbiddenUseCaseRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	if payload.MerchantID == "" || payload.UseCase == "" {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	merchantInfo := r.Context().Value(constant.CtxMerchantInfo)
	merchant, ok := merchantInfo.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	payload.Requester = merchant.MerchantId
	if err = c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	err = c.forbiddenUsecaseSvc.UnblockUseCase(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, "success")
}
