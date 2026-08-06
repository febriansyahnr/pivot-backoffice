package cardFundedPayoutController

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

type payoutRequester interface {
	SetUserID(id string)
	SetUserName(name string)
	SetMerchantID(id string)
}

func (h *handler) preparePayoutActionRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, request payoutRequester) bool {
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return false
	}
	request.SetUserID(user.UUID)
	request.SetUserName(user.Name)
	request.SetMerchantID(user.MerchantId)

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return false
	}

	if err := h.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return false
	}
	return true
}

func (h *handler) CreatePayout(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/CreatePayout")
	defer span.End()

	request := model.CreatePayoutRequest{}

	if valid := h.preparePayoutActionRequest(ctx, w, r, &request); !valid {
		return
	}

	result, err := h.cardFundedPayoutService.CreatePayout(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}

func (h *handler) ApprovePayout(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/ApprovePayout")
	defer span.End()

	request := model.ApprovePayoutRequest{
		ID: r.PathValue("payoutId"),
	}
	if valid := h.preparePayoutActionRequest(ctx, w, r, &request); !valid {
		return
	}

	result, err := h.cardFundedPayoutService.ApprovePayout(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}

func (h *handler) RejectPayout(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/cardFundedPayout/RejectPayout")
	defer span.End()

	request := model.RejectPayoutRequest{
		ID: r.PathValue("payoutId"),
	}
	if valid := h.preparePayoutActionRequest(ctx, w, r, &request); !valid {
		return
	}

	result, err := h.cardFundedPayoutService.RejectPayout(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponseOK(w, result)
}
