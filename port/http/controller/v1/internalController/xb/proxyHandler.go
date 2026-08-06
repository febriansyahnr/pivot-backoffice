package internalXbController

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ProxyHandlerWithQueryConversion creates a proxy handler that converts camelCase query params to snake_case
func (c *InternalXbController) ProxyHandlerWithQueryConversion(path string, headers map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otelTracer.Start(r.Context(), "xb-http-proxy-with-conversion:"+path)
		defer span.End()

		r = r.WithContext(ctx)

		// Convert camelCase query parameters to snake_case
		convertedQuery := url.Values{}
		for key, values := range r.URL.Query() {
			snakeKey := camelToSnakeCase(key)
			convertedQuery[snakeKey] = values
		}

		// Build the target URL with converted query parameters
		targetURL, _ := url.Parse(c.config.XbCoreProcessorConfig.BaseUrl + path)
		if len(convertedQuery) > 0 {
			targetURL.RawQuery = convertedQuery.Encode()
		}

		for key, val := range headers {
			r.Header.Set(key, val)
		}

		r.Header.Set(constant.HeaderXInternalServiceKey, c.secret.XbCoreProcessorSecret.InternalServiceKey)

		proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
			r.URL.Scheme = targetURL.Scheme
			r.URL.Host = targetURL.Host
			r.URL.Path = targetURL.Path
			r.URL.RawQuery = targetURL.RawQuery
			r.Host = targetURL.Host
		}}

		proxy.Transport = otelhttp.NewTransport(http.DefaultTransport)

		proxy.ErrorHandler = func(wr http.ResponseWriter, req *http.Request, err error) {
			requestId := req.Header.Get("X-Request-Id")

			wr.WriteHeader(http.StatusBadGateway)
			wr.Header().Set("Content-Type", "application/json")
			wr.Write([]byte(fmt.Sprintf(`{"message": "An error occurred in our service", "requestId":"%s"}`, requestId)))
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			var response any
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				return err
			}
			r.Body.Close()

			raw, err := json.Marshal(util.MapSnakeToCamel(response))
			if err != nil {
				return err
			}

			r.Body = io.NopCloser(bytes.NewBuffer(raw))
			r.ContentLength = int64(len(raw))
			r.Header.Set("Content-Length", strconv.Itoa(len(raw)))

			return nil
		}

		proxy.ServeHTTP(w, r)
	}
}

func (c *InternalXbController) ProxyHandler(path string, headers map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otelTracer.Start(r.Context(), "xb-http-proxy:"+path)
		defer span.End()

		r = r.WithContext(ctx)

		url, _ := url.Parse(c.config.XbCoreProcessorConfig.BaseUrl + path)

		for key, val := range headers {
			r.Header.Set(key, val)
		}

		r.Header.Set(constant.HeaderXInternalServiceKey, c.secret.XbCoreProcessorSecret.InternalServiceKey)

		proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
			r.URL.Scheme = url.Scheme
			r.URL.Host = url.Host
			r.URL.Path = url.Path
			r.Host = url.Host
		}}

		proxy.Transport = otelhttp.NewTransport(http.DefaultTransport)

		proxy.ErrorHandler = func(wr http.ResponseWriter, req *http.Request, err error) {
			requestId := req.Header.Get("X-Request-Id")

			wr.WriteHeader(http.StatusBadGateway)
			wr.Header().Set("Content-Type", "application/json")
			wr.Write([]byte(fmt.Sprintf(`{"message": "An error occurred in our service", "requestId":"%s"}`, requestId)))
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			var response any
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				return err
			}
			r.Body.Close()

			raw, err := json.Marshal(util.MapSnakeToCamel(response))
			if err != nil {
				return err
			}

			r.Body = io.NopCloser(bytes.NewBuffer(raw))
			r.ContentLength = int64(len(raw))
			r.Header.Set("Content-Length", strconv.Itoa(len(raw)))

			return nil
		}

		proxy.ServeHTTP(w, r)
	}
}

// camelToSnakeCase converts camelCase string to snake_case
func camelToSnakeCase(str string) string {
	if str == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}