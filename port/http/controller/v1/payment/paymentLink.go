package payment

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *PaymentController) CreatePaymentLink(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payment/CreatePaymentLink")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload unifiedPaymentModel.DashboardPaymentLinkCreateRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	payload.MerchantID = user.MerchantId
	payload.UserID = user.UUID

	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = payload.Validate(); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	respData, err := c.unifiedPaymentService.CreateDashboardPaymentLink(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, respData.ToDashboardPaymentLinkResponse())
}
