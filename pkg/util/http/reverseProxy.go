package httputil

import (
	"errors"
	"net/http"
	"net/http/httputil"

	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ReverseProxyConfiguration struct {
	PrepareFunc []func(*http.Request) error
	Logger      logger.ILogger
}

func ReverseProxy(config *ReverseProxyConfiguration) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		proxy := httputil.ReverseProxy{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Director: func(req *http.Request) {
				for _, prepareFunc := range config.PrepareFunc {
					err := prepareFunc(req)
					if err != nil {
						response.SendApiResponseError(r.Context(), w, err)
						return
					}
					config.Logger.Info(ctx, "send http request", logger.Any("url", r.URL.String()), logger.Any("headers", r.Header))
				}
			},
		}

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
			config.Logger.Warn(ctx, "error when do reverse proxy", logger.Error(err), logger.Any("url", r.URL.String()))

			errMsg := "error occurred. request id: " + requestId
			response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrBadGateway, errors.New(errMsg)))
		}

		proxy.ServeHTTP(w, r)
	})
}
