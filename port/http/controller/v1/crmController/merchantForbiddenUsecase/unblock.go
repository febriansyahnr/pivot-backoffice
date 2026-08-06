package merchantForbiddenUsecase

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1CRMMerchantForbiddenUseCaseController) Unblock(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchantForbiddenUsecase/Unblock")
	defer segment.End()

	var (
		err error
	)

	var payload merchantForbiddenUsecase.MerchantForbiddenUseCaseRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	payload.Requester = constant.UserSystemType
	if err = c.validate.Struct(payload); err != nil {
		err := errors.New(err.Error())
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	err = c.forbiddenUsecaseSvc.UnblockUseCase(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, "success")
}
