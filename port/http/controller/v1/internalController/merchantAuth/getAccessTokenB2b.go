package internalMerchantAuthController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *InternalMerchantAuthController) GetAccessTokenB2b(w http.ResponseWriter, r *http.Request) {
	_, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/merchantAuth/GetAccessTokenB2b")
	defer segment.End()

	var (
		err error
	)

	var payload merchantModel.AccessTokenB2bRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Get ClientID and ClientSecret from header
	payload.ClientID = r.Header.Get(constant.ClientIdKey) // clientId = merchantId
	payload.ClientSecret = r.Header.Get(constant.ClientSecretKey)

	if err = c.validate.Struct(payload); err != nil {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	accessToken, err := c.merchantSvc.GetAccessTokenB2b(r.Context(), payload.ClientID, payload.ClientSecret)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	resp := merchantModel.AccessTokenB2bResponse{
		AccessToken: *accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   strconv.FormatFloat(constant.MerchantAuthExpirationDuration.Seconds(), 'f', -1, 64),
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (h *InternalMerchantAuthController) GetSNAPAccessTokenB2B(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	var err error

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/merchantAuth/GetSNAPAccessTokenB2B")
	defer segment.End()

	request := &merchantModel.SNAPAccessTokenB2BReq{
		ClientId:  strings.TrimSpace(r.Header.Get(constant.HeaderXClientKey)),
		Timestamp: strings.TrimSpace(r.Header.Get(constant.HeaderXTimestamp)),
		Signature: strings.TrimSpace(r.Header.Get(constant.HeaderXSignature)),
	}
	if err = json.NewDecoder(r.Body).Decode(request); err != nil {
		if strings.Contains(err.Error(), "SNAPAccessTokenB2BReq.grantType") {
			response.SendOpenApiSnapResponseError(ctx, w, errors.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, "grantType")))

		} else {
			response.SendOpenApiSnapResponseError(ctx, w, errors.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, "(JSON Format)")))
		}
		h.logger.Warn(ctx, "Open Api Snap | Unmarshal request", logger.String("information", err.Error()))
		return
	}

	ww := monitor.WrapResponse(w, r)
	defer func() {
		var respBody response.OpenApiSnapResp

		wr, ok := w.(*middleware.ResponseWriter)
		if ok {
			_ = json.Unmarshal(wr.BodyBytes(), &respBody)
		}
		monitor.WriteAndSend(
			ctx, "api-v1.0-snap-generate-access-token", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", request.ClientId),
					fmt.Sprintf("response_code:%s", respBody.ResponseCode),
					fmt.Sprintf("response_message:%s", respBody.ResponseMessage),
				}
			},
		)
	}()

	if request.Timestamp == "" {
		err = errors.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "Header "+constant.HeaderXTimestamp))

	} else if request.ClientId == "" {
		err = errors.New(response.HttpErrUnauthorized, fmt.Errorf(constant.UnauthorizedSnapFmt, constant.HeaderXClientKey))

	} else if request.Signature == "" {
		err = errors.New(response.HttpErrUnauthorized, fmt.Errorf(constant.UnauthorizedSnapFmt, constant.HeaderXSignature))

	} else if request.GrantType == "" {
		err = errors.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "grantType"))
	}
	if err != nil {
		h.logger.Warn(ctx, "Open Api Snap | Validate required fields", logger.String("information", err.Error()))
		response.SendOpenApiSnapResponseError(ctx, w, err)
		return
	}

	if resp, err := h.merchantSvc.GetSNAPAccessTokenB2B(ctx, request); err != nil {
		response.SendOpenApiSnapResponseError(ctx, w, err)

	} else {
		response.SendOpenApiSnapResponseOK(ctx, w, resp)
	}
}

func (c *InternalMerchantAuthController) GenerateB2BTokenSNAPSignature(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		req merchantModel.GenerateSnapB2BTokenSignatureRequest
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/GenerateB2BTokenSNAPSignature")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&req); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if signature, err := c.merchantSvc.GenOpenAPISignature(ctx, &merchantModel.GenSignatureReq{
		MerchantId: req.ClientID,
		Timestamp:  req.Timestamp,
		PrivateKey: req.PrivateKey,
		GrantType:  req.GrantType,
	}); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, merchantModel.GenSignatureResp{Signature: signature})
	}

}

func (c *InternalMerchantAuthController) ValidateB2B2CTokenSNAPSignature(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		req merchantModel.SNAPValidateB2b2cTokenSignatureRequest
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/ValidateB2B2CTokenSNAPSignature")
	defer segment.End()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(&req); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.merchantSvc.ValidateSNAPAccessTokenRequestSignature(ctx, &req); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, "ok")
	}

}
