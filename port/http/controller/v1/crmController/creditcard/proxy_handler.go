package crmCreditcardController

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ProxyHandlerWithQueryConversion creates a proxy handler that converts camelCase query params and request body to snake_case
func (h *handler) ProxyHandlerWithQueryConversion(path string, headers map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otelTracer.Start(r.Context(), "creditcard-http-proxy-with-conversion:"+path)
		defer span.End()

		r = r.WithContext(ctx)

		// Convert camelCase query parameters to snake_case
		convertedQuery := url.Values{}
		for key, values := range r.URL.Query() {
			snakeKey := util.CamelToSnake(key)
			convertedQuery[snakeKey] = values
		}

		// Convert request body from camelCase to snake_case for POST/PUT/PATCH
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			if r.Body != nil && r.ContentLength > 0 {
				var requestBody any
				if err := json.NewDecoder(r.Body).Decode(&requestBody); err == nil {
					r.Body.Close()

					// Convert camelCase to snake_case
					converted := util.MapCamelToSnake(requestBody)
					raw, _ := json.Marshal(converted)

					r.Body = io.NopCloser(bytes.NewBuffer(raw))
					r.ContentLength = int64(len(raw))
					r.Header.Set("Content-Length", strconv.Itoa(len(raw)))
				}
			}
		}

		// Build the target URL with converted query parameters
		targetURL, _ := url.Parse(h.config.CreditcardCoreProcessorConfig.BaseUrl + path)
		if len(convertedQuery) > 0 {
			targetURL.RawQuery = convertedQuery.Encode()
		}

		for key, val := range headers {
			r.Header.Set(key, val)
		}

		r.Header.Set(constant.HeaderXInternalServiceKey, h.secret.CreditcardCoreProcessorSecret.InternalServiceKey)

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

			wr.Header().Set("Content-Type", "application/json")
			wr.WriteHeader(http.StatusBadGateway)
			wr.Write([]byte(fmt.Sprintf(`{"message": "An error occurred in our service", "requestId":"%s"}`, requestId)))
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			// Only process JSON responses
			contentType := r.Header.Get("Content-Type")
			if !bytes.Contains([]byte(contentType), []byte("application/json")) {
				// Pass through non-JSON responses unchanged
				return nil
			}

			// Read the body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return nil // Pass through on read error
			}
			r.Body.Close()

			// Try to decode as JSON
			var response any
			if err := json.Unmarshal(bodyBytes, &response); err != nil {
				// If JSON decode fails, restore original body and pass through
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				return nil
			}

			// Convert snake_case to camelCase
			raw, err := json.Marshal(util.MapSnakeToCamel(response))
			if err != nil {
				// If marshal fails, restore original body and pass through
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				return nil
			}

			r.Body = io.NopCloser(bytes.NewBuffer(raw))
			r.ContentLength = int64(len(raw))
			r.Header.Set("Content-Length", strconv.Itoa(len(raw)))

			return nil
		}

		proxy.ServeHTTP(w, r)
	}
}

// ProxyHandler creates a simple proxy handler without query parameter conversion
func (h *handler) ProxyHandler(path string, headers map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otelTracer.Start(r.Context(), "creditcard-http-proxy:"+path)
		defer span.End()

		r = r.WithContext(ctx)

		url, _ := url.Parse(h.config.CreditcardCoreProcessorConfig.BaseUrl + path)

		for key, val := range headers {
			r.Header.Set(key, val)
		}

		r.Header.Set(constant.HeaderXInternalServiceKey, h.secret.CreditcardCoreProcessorSecret.InternalServiceKey)

		proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
			r.URL.Scheme = url.Scheme
			r.URL.Host = url.Host
			r.URL.Path = url.Path
			r.Host = url.Host
		}}

		proxy.Transport = otelhttp.NewTransport(http.DefaultTransport)

		proxy.ErrorHandler = func(wr http.ResponseWriter, req *http.Request, err error) {
			requestId := req.Header.Get("X-Request-Id")

			wr.Header().Set("Content-Type", "application/json")
			wr.WriteHeader(http.StatusBadGateway)
			wr.Write([]byte(fmt.Sprintf(`{"message": "An error occurred in our service", "requestId":"%s"}`, requestId)))
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			// Only process JSON responses
			contentType := r.Header.Get("Content-Type")
			if !bytes.Contains([]byte(contentType), []byte("application/json")) {
				// Pass through non-JSON responses unchanged
				return nil
			}

			// Read the body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return nil // Pass through on read error
			}
			r.Body.Close()

			// Try to decode as JSON
			var response any
			if err := json.Unmarshal(bodyBytes, &response); err != nil {
				// If JSON decode fails, restore original body and pass through
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				return nil
			}

			// Convert snake_case to camelCase
			raw, err := json.Marshal(util.MapSnakeToCamel(response))
			if err != nil {
				// If marshal fails, restore original body and pass through
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				return nil
			}

			r.Body = io.NopCloser(bytes.NewBuffer(raw))
			r.ContentLength = int64(len(raw))
			r.Header.Set("Content-Length", strconv.Itoa(len(raw)))

			return nil
		}

		proxy.ServeHTTP(w, r)
	}
}
