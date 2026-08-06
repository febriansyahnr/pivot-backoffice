package payment

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *PaymentController) VCCTerminalBatchCharge(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/payments/VCCTerminalBatchCharge")
	defer segment.End()

	userInfo, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request := paymentModel.VCCTerminalChargeRequest{
		MerchantID:       userInfo.MerchantId,
		UserID:           userInfo.UUID,
		EncryptedRequest: &encryption.DataEncryption{},
	}
	if err := json.NewDecoder(r.Body).Decode(request.EncryptedRequest); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.StructCtx(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	result, err := h.paymentService.VCCTerminalBatchCharge(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}
