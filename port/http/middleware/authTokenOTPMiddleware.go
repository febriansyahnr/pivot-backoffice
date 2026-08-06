package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"

	"github.com/go-chi/chi/v5/middleware"
)

type tokenOTPFunc func(context.Context, string) (*otpModel.TokenOTPClaims, error)

func AuthTokenFromOTP(jwtCore jwt.IJwt, redis redisExt.IRedisExt) MiddlewareFunc {
	return authTokenFromOTP(constant.TokenOTPNamespace, jwtCore.ValidateTokenFromOTP, redis)
}

func AuthTokenFromFeature2FA(jwtCore jwt.IJwt, redis redisExt.IRedisExt) MiddlewareFunc {
	return authTokenFromOTP(constant.TokenFeature2FANamespace, jwtCore.ValidateTokenFromFeature2FA, redis)
}

func authTokenFromOTP(key string, validate tokenOTPFunc, redis redisExt.IRedisExt) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				claims  *otpModel.TokenOTPClaims
				authKey string
			)

			err := func() (err error) {
				ctx, segment := otelTracer.Start(r.Context(), "http/middleware/authTokenFromOTP")
				defer segment.End()

				token := r.Header.Get("Authorization")
				if !strings.Contains(token, "Bearer") {
					return pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf("invalid token format"))
				}
				token = strings.Replace(token, "Bearer ", "", 1)

				if claims, err = validate(ctx, token); err != nil {
					return pkgErrs.New(response.HttpErrUnauthorized, err)
				}

				if key == constant.TokenFeature2FANamespace && !claims.Identifier.FeatureValidation(r.URL.Path) {
					return pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf("invalid feature token"))
				}

				var cacheValue any = &claims.Email

				if key == constant.TokenOTPNamespace {
					authKey = fmt.Sprintf(
						constant.OTPKeyFormatting+"%s", claims.UUID, claims.Identifier.FeatureName(), ":"+constant.TokenOTPNamespace,
					)
					cacheValue = &claims.VerifyOTP

				} else {
					authKey = fmt.Sprintf(
						constant.OTPKeyFormatting+"%s", claims.UUID, claims.Identifier.FeatureName(), ":"+key+":"+token,
					)
				}

				if err = redis.Get(ctx, authKey).Scan(cacheValue); err != nil {
					return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrOTPTokenNotRegistered)
				}

				if key == constant.TokenOTPNamespace && claims.VerifyOTP.Token != token {
					return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrOTPTokenNotRegistered)
				}
				return nil
			}()

			if err != nil {
				response.SendApiResponseError(r.Context(), w, err)
				return
			}

			ctx := context.WithValue(r.Context(), pdkConstant.CtxMerchantIDKey, claims.UUID)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r.WithContext(context.WithValue(ctx, constant.CtxTokenOTPKey, claims)))

			if ww.Status() != http.StatusOK {
				return
			}
			_ = redis.Del(r.Context(), authKey)
		})
	}
}

func SpecialCaseRequireAuthForSendOTP(jwtCore jwt.IJwt) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, segment := otelTracer.Start(r.Context(), "http/middleware/SpecialCaseRequireAuthForSendOTP")
			defer segment.End()

			buf, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(buf))

			payload := otpModel.SendOTPReq{}
			if err := json.Unmarshal(buf, &payload); err != nil {
				response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
				return
			}

			if payload.Event != constant.OTPIdentifierResetPIN && payload.Event != constant.OTPIdentifierChangePassword {
				next.ServeHTTP(w, r)
				return
			}

			token := r.Header.Get("X-Access-Token")
			if token == "" {
				response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, errors.New("token required")))
				return
			}

			claims, err := jwtCore.Verify(ctx, token)
			if err != nil {
				response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, err))
				return

			} else if claims.Email != payload.Email {
				response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("this email is not associated with your account")))
				return
			}
			if _, err := jwtCore.GetTokenLoggedInDevices(ctx, claims.Email, claims.DeviceIdentifier); err != nil {
				response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, errors.New("token not registered")))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
