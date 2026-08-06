package openApi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	snapSign "github.com/paper-indonesia/pdk/go/snap/signature"
	"go.uber.org/zap"
)

type SnapTrxMiddleware interface {
	Authorize(next http.Handler) http.Handler
}

type snapTrxMiddleware struct {
	logger          logger.ILogger
	jwt             jwt.IJwt
	merchantService service.IMerchantService
}

var mandatoryHeaders = []string{
	constant.HeaderAuthorization,
	constant.HeaderXTimestamp,
	constant.HeaderXSignature,
	constant.HeaderXPartnerId,
	constant.HeaderXExternalId,
	constant.HeaderXChannelId,
}

var (
	unauthorizedErrFmt       = "Unauthorized %s"
	invalidB2BTokenErrMsg    = "Invalid Token (B2B)"
	invalidMandatoryFieldFmt = "Invalid Mandatory Field %s"
	invalidFieldFormatFmt    = "Invalid Field Format %s"
)

var otelTracer = otel.Tracer("OpenApiMiddleware")

func NewSnapAuthMiddleware(logger logger.ILogger, jwt jwt.IJwt, merchantSvc service.IMerchantService) SnapTrxMiddleware {
	return &snapTrxMiddleware{logger, jwt, merchantSvc}
}

func (md *snapTrxMiddleware) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, segment := otelTracer.Start(r.Context(), "http/middleware/openApi/Authorize")
		defer segment.End()

		if err := md.ValidateMandatoryHeaders(ctx, r.Header); err != nil {
			response.SendOpenApiSnapResponseError(ctx, w, err)
			return
		}

		if channelId := r.Header.Get(constant.HeaderXChannelId); channelId != "HARSYA" {
			md.logger.Warn(ctx, "Open Api Snap | Channel-Id is not registered", zap.String(constant.HeaderXChannelId, channelId))
			response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf(unauthorizedErrFmt, constant.HeaderXChannelId)))
			return
		}

		tokenHeader := r.Header.Get(constant.HeaderAuthorization)

		merchantAuthRespFunc := func(e error) {
			if strings.Contains(e.Error(), response.HttpErrInternal+" | ") {
				md.logger.Error(ctx, "Open Api Snap | Internal server error", zap.Error(e))
				response.SendOpenApiSnapResponseError(ctx, w, e)

			} else {
				md.logger.Warn(ctx, "Open Api Snap | "+e.Error(), zap.String(constant.HeaderAuthorization, tokenHeader))
				response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.SnapErrInvalidTokenB2B, errors.New(invalidB2BTokenErrMsg)))
			}
		}

		merchant, err := md.MerchantAuthorization(ctx, tokenHeader)
		if err != nil {
			merchantAuthRespFunc(err)
			return
		}
		if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
			subMerchant, err := md.MerchantValidation(ctx, subMerchantId)
			if err != nil {
				merchantAuthRespFunc(err)
				return

			} else if subMerchant.ParentID.String != merchant.UUID {
				md.logger.Warn(
					ctx, "Open Api Snap | Invalid sub merchant with parent id",
					zap.Any("data", map[string]string{
						"merchant":               merchant.UUID,
						"sub_merchant_id":        subMerchant.UUID,
						"sub_merchant_parent_id": subMerchant.ParentID.String},
					))
				response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.SnapErrInvalidTokenB2B, errors.New(invalidB2BTokenErrMsg)))
				return
			}

			merchant = subMerchant
		}

		rawBody, _ := io.ReadAll(r.Body)

		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

		rawJsonBody := json.RawMessage{}
		if err := json.Unmarshal(rawBody, &rawJsonBody); err != nil {
			md.logger.Warn(ctx, "Open Api Snap | Unmarshal JSON message", zap.ByteString("raw_body_request", rawBody))
			response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(invalidFieldFormatFmt, "(Request Body Format)")))
			return
		}

		secret, err := md.merchantService.GetPKCS8SecretKey(ctx, merchant.UUID)
		if err != nil {
			md.logger.Error(ctx, "Open Api Snap | Get PKCS8 secret", zap.Error(err))
			response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.SnapErrInvalidTokenB2B, errors.New(invalidB2BTokenErrMsg)))
			return
		}
		rawPKCS8Secret, _ := base64.StdEncoding.DecodeString(secret.Data)

		pkcs8Secret := merchantModel.PKCS8SecretKeyResponse{}
		_ = json.Unmarshal(rawPKCS8Secret, &pkcs8Secret)

		snapPath := r.RequestURI
		if path := r.Header.Get(constant.HeaderXSnapPath); path != "" {
			snapPath = path
		}
		snapSignPayload := snapSign.TrxSignature{
			HttpMethod:   r.Method,
			URL:          snapPath,
			AccessToken:  r.Header.Get(constant.HeaderAuthorization),
			ClientSecret: pkcs8Secret.MerchantSecret,
			Timestamp:    r.Header.Get(constant.HeaderXTimestamp),
			BodyPayload:  rawJsonBody,
		}
		if err := md.SnapTrxVerify(ctx, snapSignPayload, r.Header.Get(constant.HeaderXSignature)); err != nil {
			md.logger.Warn(ctx, "Open Api Snap | Invalid signature", zap.String("detail", err.Error()))
			response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf(unauthorizedErrFmt, constant.HeaderXSignature)))
			return
		}

		ctx = context.WithValue(ctx, constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: merchant.UUID,
			ClientId:   r.Header.Get(constant.HeaderXPartnerId),
		})
		r = r.WithContext(ctx)

		// Add merchant id to context
		ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, merchant.UUID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (md *snapTrxMiddleware) ValidateMandatoryHeaders(ctx context.Context, reqHeader http.Header) error {
	for _, key := range mandatoryHeaders {

		if val := reqHeader.Get(key); val == "" {

			md.logger.Warn(ctx, "Open API Snap | Header "+key+" is required", zap.Any("headers", reqHeader))

			switch key {
			default:
				return pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf(unauthorizedErrFmt, key))

			case constant.HeaderAuthorization:
				return pkgErrs.New(response.SnapErrInvalidTokenB2B, errors.New(invalidB2BTokenErrMsg))

			case constant.HeaderXTimestamp, constant.HeaderXExternalId:
				return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(invalidMandatoryFieldFmt, "Header "+key))
			}
		}
	}
	return nil
}

func (md *snapTrxMiddleware) MerchantAuthorization(ctx context.Context, token string) (*merchantModel.Merchant, error) {
	if !strings.Contains(token, "Bearer ") {
		return nil, errors.New("invalid token format")
	}
	token = strings.Replace(token, "Bearer ", "", 1)

	claims, err := md.jwt.VerifyMerchantToken(ctx, token)
	if err != nil {
		return nil, err

	} else if time.Now().UTC().Unix() > claims.ExpiresAt.Unix() {
		return nil, errors.New("token has expired")
	}
	return md.MerchantValidation(ctx, claims.MerchantId)
}

func (md *snapTrxMiddleware) MerchantValidation(ctx context.Context, id string) (*merchantModel.Merchant, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, err
	}
	merchantInfo, err := md.merchantService.FindMerchantByID(ctx, id)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrInternal, err)

	} else if merchantInfo == nil {
		return nil, constant.ErrMerchantNotFound
	}
	return merchantInfo, nil
}

func (md *snapTrxMiddleware) SnapTrxVerify(ctx context.Context, sign snapSign.TrxSignature, headerSign string) error {

	trxSign, err := snapSign.NewTrxSignature(sign)
	if err != nil {
		return err
	}

	if verify, err := trxSign.Verify(headerSign); err != nil || !verify {
		if err == nil {
			err = errors.New("invalid signature")
		}
		return err
	}
	return nil
}
