package tnc

import (
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Sign records the authenticated user's acceptance of the active TNC version
// on behalf of their merchant. The actor (user id, email, merchant id) is
// resolved from the JWT claims placed on the request context by AuthMiddleware.
func (c *TNCSigningController) Sign(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/tnc/Sign")
	defer segment.End()

	claims, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok || claims == nil || claims.UUID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserUnauthorized))
		return
	}

	merchantID := claims.MerchantId
	if merchantID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
		return
	}

	req := &tncModel.SignTNCRequest{
		MerchantID: merchantID,
		SignedBy:   claims.UUID,
		Email:      claims.Email,
		IPAddress:  readUserIP(r),
		UserAgent:  r.UserAgent(),
	}

	history, err := c.service.SignTNC(ctx, req)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, history)
}

func readUserIP(r *http.Request) string {
	// Best-effort: prefer X-Forwarded-For from the edge, fall back to RemoteAddr.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return r.RemoteAddr
}
