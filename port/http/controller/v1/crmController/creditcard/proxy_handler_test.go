package crmCreditcardController

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyHandlerWithQueryConversion_QueryParams(t *testing.T) {
	// Create a test server that will receive the proxied request
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters are converted to snake_case
		assert.Equal(t, "merchant-123", r.URL.Query().Get("merchant_id"))
		assert.Equal(t, "444000", r.URL.Query().Get("bin_number"))
		assert.Equal(t, "TEST001", r.URL.Query().Get("mid"))
		assert.Equal(t, "3DS", r.URL.Query().Get("use_case"))

		// Return a response with snake_case keys
		response := map[string]interface{}{
			"code": "200",
			"data": map[string]interface{}{
				"merchant_id": "merchant-123",
				"bin_number":  "444000",
				"mid":         "TEST001",
				"use_case":    "3DS",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer backendServer.Close()

	// Create handler with test config
	h := &handler{
		config: &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: backendServer.URL,
			},
		},
		secret: &config.Secret{
			CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
				InternalServiceKey: "test-key",
			},
		},
	}

	// Create test request with camelCase query parameters
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bin-mappings?merchantId=merchant-123&binNumber=444000&mid=TEST001&useCase=3DS", nil)
	w := httptest.NewRecorder()

	// Call the proxy handler
	handler := h.ProxyHandlerWithQueryConversion("/api/v1/bin-mappings", nil)
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response body has camelCase keys
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "200", response["code"])

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "merchant-123", data["merchantId"])
	assert.Equal(t, "444000", data["binNumber"])
	assert.Equal(t, "TEST001", data["mid"])
	assert.Equal(t, "3DS", data["useCase"])
}

func TestProxyHandlerWithQueryConversion_RequestBodyConversion(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    map[string]interface{}
		shouldConvert  bool
		expectedBody   map[string]interface{}
	}{
		{
			name:   "POST with camelCase body",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"merchantId": "merchant-123",
				"binNumber":  "444000",
				"mid":        "TEST001",
				"useCase":    "3DS",
				"createdBy":  "admin@example.com",
			},
			shouldConvert: true,
			expectedBody: map[string]interface{}{
				"merchant_id": "merchant-123",
				"bin_number":  "444000",
				"mid":         "TEST001",
				"use_case":    "3DS",
				"created_by":  "admin@example.com",
			},
		},
		{
			name:   "PUT with camelCase body",
			method: http.MethodPut,
			requestBody: map[string]interface{}{
				"mid":       "TEST002",
				"useCase":   "NON_3DS",
				"updatedBy": "admin@example.com",
			},
			shouldConvert: true,
			expectedBody: map[string]interface{}{
				"mid":        "TEST002",
				"use_case":   "NON_3DS",
				"updated_by": "admin@example.com",
			},
		},
		{
			name:   "PATCH with camelCase body",
			method: http.MethodPatch,
			requestBody: map[string]interface{}{
				"useCase": "3DS",
			},
			shouldConvert: true,
			expectedBody: map[string]interface{}{
				"use_case": "3DS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that will receive the proxied request
			backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Read the request body
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var receivedBody map[string]interface{}
				err = json.Unmarshal(body, &receivedBody)
				require.NoError(t, err)

				if tt.shouldConvert {
					// Verify body was converted to snake_case
					assert.Equal(t, tt.expectedBody, receivedBody)
				}

				// Return a response
				response := map[string]interface{}{
					"code": "200",
					"data": map[string]interface{}{
						"created": true,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer backendServer.Close()

			// Create handler with test config
			h := &handler{
				config: &config.Config{
					CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
						BaseUrl: backendServer.URL,
					},
				},
				secret: &config.Secret{
					CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
						InternalServiceKey: "test-key",
					},
				},
			}

			// Create test request
			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(tt.method, "/api/v1/bin-mappings", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call the proxy handler
			handler := h.ProxyHandlerWithQueryConversion("/api/v1/bin-mappings", nil)
			handler(w, req)

			// Verify response
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestProxyHandlerWithQueryConversion_ResponseConversion(t *testing.T) {
	// Create a test server that returns snake_case response
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"code":    "200",
			"message": "success",
			"data": map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"uuid":        "123e4567-e89b-12d3-a456-426614174000",
						"merchant_id": "merchant-123",
						"bin_number":  "444000",
						"mid":         "TEST001",
						"use_case":    "3DS",
						"created_at":  "2024-01-01T00:00:00Z",
						"updated_at":  "2024-01-01T00:00:00Z",
					},
				},
				"pagination": map[string]interface{}{
					"page_number":  1,
					"page_limit":   10,
					"total_record": 1,
					"total_page":   1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer backendServer.Close()

	// Create handler with test config
	h := &handler{
		config: &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: backendServer.URL,
			},
		},
		secret: &config.Secret{
			CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
				InternalServiceKey: "test-key",
			},
		},
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bin-mappings", nil)
	w := httptest.NewRecorder()

	// Call the proxy handler
	handler := h.ProxyHandlerWithQueryConversion("/api/v1/bin-mappings", nil)
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response body has camelCase keys
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "200", response["code"])
	assert.Equal(t, "success", response["message"])

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)

	results, ok := data["results"].([]interface{})
	require.True(t, ok)
	require.Len(t, results, 1)

	firstResult, ok := results[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", firstResult["uuid"])
	assert.Equal(t, "merchant-123", firstResult["merchantId"])
	assert.Equal(t, "444000", firstResult["binNumber"])
	assert.Equal(t, "TEST001", firstResult["mid"])
	assert.Equal(t, "3DS", firstResult["useCase"])
	assert.Equal(t, "2024-01-01T00:00:00Z", firstResult["createdAt"])
	assert.Equal(t, "2024-01-01T00:00:00Z", firstResult["updatedAt"])

	pagination, ok := data["pagination"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), pagination["pageNumber"])
	assert.Equal(t, float64(10), pagination["pageLimit"])
	assert.Equal(t, float64(1), pagination["totalRecord"])
	assert.Equal(t, float64(1), pagination["totalPage"])
}

func TestProxyHandlerWithQueryConversion_CustomHeaders(t *testing.T) {
	// Create a test server that verifies headers
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify internal service key header
		assert.Equal(t, "test-key", r.Header.Get("X-Internal-Service-Key"))

		// Verify custom headers
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))

		response := map[string]interface{}{
			"code": "200",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer backendServer.Close()

	// Create handler with test config
	h := &handler{
		config: &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: backendServer.URL,
			},
		},
		secret: &config.Secret{
			CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
				InternalServiceKey: "test-key",
			},
		},
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bin-mappings", nil)
	w := httptest.NewRecorder()

	// Call the proxy handler with custom headers
	customHeaders := map[string]string{
		"X-Custom-Header": "custom-value",
	}
	handler := h.ProxyHandlerWithQueryConversion("/api/v1/bin-mappings", customHeaders)
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProxyHandler_SimpleProxy(t *testing.T) {
	// Create a test server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify internal service key header
		assert.Equal(t, "test-key", r.Header.Get("X-Internal-Service-Key"))

		response := map[string]interface{}{
			"code": "200",
			"data": map[string]interface{}{
				"deleted": true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer backendServer.Close()

	// Create handler with test config
	h := &handler{
		config: &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: backendServer.URL,
			},
		},
		secret: &config.Secret{
			CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
				InternalServiceKey: "test-key",
			},
		},
	}

	// Create test request
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bin-mappings/123", nil)
	w := httptest.NewRecorder()

	// Call the proxy handler
	handler := h.ProxyHandler("/api/v1/bin-mappings/123", nil)
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response still converts snake_case to camelCase
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "200", response["code"])
}

func TestProxyHandler_ErrorHandling(t *testing.T) {
	// Create a test server that returns an error
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"code":    "400",
			"message": "invalid_request",
			"error":   "Invalid bin number",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer backendServer.Close()

	// Create handler with test config
	h := &handler{
		config: &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: backendServer.URL,
			},
		},
		secret: &config.Secret{
			CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
				InternalServiceKey: "test-key",
			},
		},
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bin-mappings", nil)
	w := httptest.NewRecorder()

	// Call the proxy handler
	handler := h.ProxyHandlerWithQueryConversion("/api/v1/bin-mappings", nil)
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error response is still converted to camelCase
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "400", response["code"])
	assert.Equal(t, "invalid_request", response["message"])
	assert.Equal(t, "Invalid bin number", response["error"])
}
