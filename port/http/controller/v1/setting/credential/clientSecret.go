package credential

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetClientSecretById		godoc
// @Summary					To display the client secret by ID (PIN authentication required)
// @Description				To display the client secret by ID (PIN authentication required)
// @ID						settings-credentials-view-client-secret
// @Tags					Settings - Credentials
// @Accept					json
// @Produce					json
// @Param 					id		path		string true "Credential ID"
// @Success					200  	{object}	response.ApiResponse{data=credential.ClientSecretResp}
// @Failure					500  	{object}	response.ApiResponse
// @Router					/api/v1/settings/credentials/client-secrets/{id} [get]
// @Security				Bearer
func (h *handler) GetClientSecretById(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/credential/GetClientSecretById")
	defer segment.End()

	h.ClientSecretById(w, r)
}

// GenerateClientSecretById	godoc
// @Summary					To regenerate client secret by ID (PIN authentication required)
// @Description				To regenerate client secret by ID (PIN authentication required)
// @ID						settings-credentials-generate-client-secret
// @Tags					Settings - Credentials
// @Accept					json
// @Produce					json
// @Param 					id		path		string true "Credential ID"
// @Success					200  	{object}	response.ApiResponse{data=credential.ClientSecretResp}
// @Failure					500  	{object}	response.ApiResponse
// @Router					/api/v1/settings/credentials/client-secrets/{id} [post]
// @Security				Bearer
func (h *handler) GenerateClientSecretById(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/setting/credential/GenerateClientSecretById")
	defer segment.End()

	h.ClientSecretById(w, r)
}

// Note: General function
func (h *handler) ClientSecretById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	reqPIN := []byte{}
	if decodedPIN := r.Header.Get(constant.HeaderXRequestPIN); decodedPIN != "" {
		reqPIN, _ = base64.StdEncoding.DecodeString(decodedPIN)
	}
	request := &credential.ClientSecretReq{
		Info:       r,
		UserID:     user.UUID,
		MerchantID: user.MerchantId,
		SecretID:   r.PathValue("id"),
		PIN:        string(reqPIN),
		Action:     r.Method,
	}
	if err := h.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	resp, err := h.service.ClientSecretById(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	buff, _ := json.Marshal(
		response.ApiResponse{
			Data:    resp,
			Code:    response.HttpStatusOK,
			Message: "OK",
		})
	w.Header().Set(constant.HeaderXResponseSignature, encryption.GenerateHMAC(h.securitySecret.RespEncryptKey, string(buff)))

	response.SendApiResponseOK(w, resp)
}
