package merchant

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *MerchantController) CloseMerchant(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchant/CloseMerchant")
	defer segment.End()

	var (
		err    error
		userID = "retool"
	)

	id := chi.URLParam(r, "id")
	if err = uuid.Validate(id); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, fmt.Errorf("merchant id is required")))
		return
	}

	var payload merchantModel.UpdateMerchantStatus
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	payload.ID = id
	payload.Status = constant.MerchantStatusClosed
	if err = c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.merchantSvc.CloseMerchant(ctx, &payload); err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		nil,
		&userID,
		constant.TagMerchant,
		constant.ActivityUserUpdateMerchantFee,
		payload,
	)

	response.SendGeneralResponseOK(w, "Update merchant status success")
}
