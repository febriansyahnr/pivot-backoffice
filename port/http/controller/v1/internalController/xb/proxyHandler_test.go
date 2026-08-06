package internalXbController

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/stretchr/testify/assert"
)

func TestProxyHandler(t *testing.T) {
	controller := &InternalXbController{
		config: &config.Config{
			XbCoreProcessorConfig: config.XbCoreProcessorConfig{
				BaseUrl: "http://localhost:8080",
			},
		},
		secret: &config.Secret{
			XbCoreProcessorSecret: config.XbCoreProcessorSecret{
				InternalServiceKey: "secret",
			},
		},
	}

	tests := []struct {
		name       string
		path       string
		headers    map[string]string
		statusCode int
	}{
		{"valid path and headers", "/test", map[string]string{"key": "value"}, http.StatusBadGateway},
		{"invalid path", "", nil, http.StatusBadGateway},
		{"nil headers", "/test", nil, http.StatusBadGateway},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", test.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			for key, val := range test.headers {
				req.Header.Set(key, val)
			}

			w := httptest.NewRecorder()
			handler := controller.ProxyHandler(test.path, test.headers)
			handler(w, req)

			assert.Equal(t, test.statusCode, w.Code)
		})
	}
}

func TestProxyHandlerError(t *testing.T) {
	controller := &InternalXbController{
		config: &config.Config{
			XbCoreProcessorConfig: config.XbCoreProcessorConfig{
				BaseUrl: "http://localhost:8080",
			},
		},
		secret: &config.Secret{
			XbCoreProcessorSecret: config.XbCoreProcessorSecret{
				InternalServiceKey: "secret",
			},
		},
	}

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := controller.ProxyHandler("/test", nil)
	handler(w, req)

	// Simulate an error
	controller.config.XbCoreProcessorConfig.BaseUrl = "invalid-url"

	w = httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "An error occurred in our service", response["message"])
}

func TestProxyHandlerWithQueryConversion(t *testing.T) {
	controller := &InternalXbController{
		config: &config.Config{
			XbCoreProcessorConfig: config.XbCoreProcessorConfig{
				BaseUrl: "http://localhost:8080",
			},
		},
		secret: &config.Secret{
			XbCoreProcessorSecret: config.XbCoreProcessorSecret{
				InternalServiceKey: "secret",
			},
		},
	}

	tests := []struct {
		name             string
		path             string
		queryParams      string
		expectedQuery    string
		headers          map[string]string
		statusCode       int
	}{
		{
			name:          "camelCase to snake_case conversion",
			path:          "/api/v1/master/bank/list",
			queryParams:   "countryCode=JPN&bankCode=ABC123",
			expectedQuery: "bank_code=ABC123&country_code=JPN", // URL encoding will sort alphabetically
			headers:       map[string]string{"Content-Type": "application/json"},
			statusCode:    http.StatusBadGateway, // Since we're not running actual server
		},
		{
			name:          "mixed case parameters",
			path:          "/api/v1/master/bank/list",
			queryParams:   "countryCode=USA&transferMethod=SWIFT&accountType=CHECKING",
			expectedQuery: "account_type=CHECKING&country_code=USA&transfer_method=SWIFT",
			headers:       nil,
			statusCode:    http.StatusBadGateway,
		},
		{
			name:          "already snake_case parameters",
			path:          "/api/v1/master/bank/list",
			queryParams:   "country_code=SGP&bank_code=DBS",
			expectedQuery: "bank_code=DBS&country_code=SGP",
			headers:       nil,
			statusCode:    http.StatusBadGateway,
		},
		{
			name:          "single parameter conversion",
			path:          "/api/v1/master/bank/list",
			queryParams:   "destinationCurrency=USD",
			expectedQuery: "destination_currency=USD",
			headers:       nil,
			statusCode:    http.StatusBadGateway,
		},
		{
			name:          "no query parameters",
			path:          "/api/v1/master/bank/list",
			queryParams:   "",
			expectedQuery: "",
			headers:       nil,
			statusCode:    http.StatusBadGateway,
		},
		{
			name:          "complex camelCase conversion",
			path:          "/api/v1/master/bank/list",
			queryParams:   "identificationNumber=123456&sourceOfIncome=salary&XMLHttpRequest=true",
			expectedQuery: "identification_number=123456&source_of_income=salary&x_m_l_http_request=true",
			headers:       nil,
			statusCode:    http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create request with query parameters
			reqURL := test.path
			if test.queryParams != "" {
				reqURL += "?" + test.queryParams
			}

			req, err := http.NewRequest("GET", reqURL, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Add headers if provided
			for key, val := range test.headers {
				req.Header.Set(key, val)
			}

			// Create mock server to capture the forwarded request
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the query parameters have been converted correctly
				if test.expectedQuery != "" {
					assert.Equal(t, test.expectedQuery, r.URL.RawQuery, "Query parameters should be converted to snake_case")
				} else {
					assert.Empty(t, r.URL.RawQuery, "Query should be empty when no parameters provided")
				}

				// Verify internal service key header is added
				assert.Equal(t, "secret", r.Header.Get("X-Internal-Service-Key"), "Internal service key header should be set")

				// Return a test response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"test_field": "test_value"}`))
			}))
			defer mockServer.Close()

			// Update controller to use mock server
			controller.config.XbCoreProcessorConfig.BaseUrl = mockServer.URL

			w := httptest.NewRecorder()
			handler := controller.ProxyHandlerWithQueryConversion(test.path, test.headers)
			handler(w, req)

			// For successful proxy requests to mock server, we expect 200
			// For tests with invalid URLs, we expect BadGateway
			expectedStatus := http.StatusOK
			if test.statusCode == http.StatusBadGateway {
				// Reset to invalid URL to test error case
				controller.config.XbCoreProcessorConfig.BaseUrl = "http://localhost:8080" // This will fail
				w = httptest.NewRecorder()
				handler(w, req)
				expectedStatus = http.StatusBadGateway
			}

			assert.Equal(t, expectedStatus, w.Code)
		})
	}
}

func TestProxyHandlerWithQueryConversionError(t *testing.T) {
	controller := &InternalXbController{
		config: &config.Config{
			XbCoreProcessorConfig: config.XbCoreProcessorConfig{
				BaseUrl: "http://invalid-url-that-does-not-exist:9999",
			},
		},
		secret: &config.Secret{
			XbCoreProcessorSecret: config.XbCoreProcessorSecret{
				InternalServiceKey: "secret",
			},
		},
	}

	req, err := http.NewRequest("GET", "/api/v1/master/bank/list?countryCode=JPN", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := controller.ProxyHandlerWithQueryConversion("/api/v1/master/bank/list", nil)
	handler(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "An error occurred in our service", response["message"])
}

func TestCamelToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"countryCode", "country_code"},
		{"bankCode", "bank_code"},
		{"transferMethod", "transfer_method"},
		{"accountType", "account_type"},
		{"sourceOfIncome", "source_of_income"},
		{"identificationNumber", "identification_number"},
		{"destinationCurrency", "destination_currency"},
		{"FxRate", "fx_rate"},
		{"XMLHttpRequest", "x_m_l_http_request"},
		{"ID", "i_d"},
		{"simple", "simple"},
		{"alreadySnake_case", "already_snake_case"},
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"ABC", "a_b_c"},
		{"testHTMLParser", "test_h_t_m_l_parser"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := camelToSnakeCase(test.input)
			assert.Equal(t, test.expected, result, "camelToSnakeCase conversion should be correct")
		})
	}
}

func TestProxyHandlerWithQueryConversionResponseTransformation(t *testing.T) {
	controller := &InternalXbController{
		config: &config.Config{
			XbCoreProcessorConfig: config.XbCoreProcessorConfig{
				BaseUrl: "http://localhost:8080",
			},
		},
		secret: &config.Secret{
			XbCoreProcessorSecret: config.XbCoreProcessorSecret{
				InternalServiceKey: "secret",
			},
		},
	}

	// Mock response from XB Core Processor (snake_case)
	mockResponse := `{
		"message": "success",
		"data": {
			"country_code": "JPN",
			"bank_name": "Bank of Tokyo",
			"bank_code": "BOT001",
			"transfer_methods": [
				{
					"method_name": "SWIFT",
					"processing_time": "1-3 days",
					"minimum_amount": 100.0
				}
			],
			"created_at": "2023-12-05T10:00:00Z",
			"updated_at": "2023-12-05T12:00:00Z"
		}
	}`

	// Create mock server that returns snake_case response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query conversion happened
		assert.Equal(t, "country_code=JPN", r.URL.RawQuery)

		// Return snake_case response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	controller.config.XbCoreProcessorConfig.BaseUrl = mockServer.URL

	// Make request with camelCase query params
	req, err := http.NewRequest("GET", "/api/v1/master/bank/list?countryCode=JPN", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := controller.ProxyHandlerWithQueryConversion("/api/v1/master/bank/list", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	// Verify response has been transformed to camelCase
	assert.Equal(t, "success", response["message"])

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok, "Data field should be present")

	// Verify snake_case fields are converted to camelCase
	assert.Equal(t, "JPN", data["countryCode"])
	assert.Equal(t, "Bank of Tokyo", data["bankName"])
	assert.Equal(t, "BOT001", data["bankCode"])
	assert.Equal(t, "2023-12-05T10:00:00Z", data["createdAt"])
	assert.Equal(t, "2023-12-05T12:00:00Z", data["updatedAt"])

	// Verify nested array transformation
	transferMethods, ok := data["transferMethods"].([]interface{})
	assert.True(t, ok, "Transfer methods should be an array")
	assert.Len(t, transferMethods, 1)

	method, ok := transferMethods[0].(map[string]interface{})
	assert.True(t, ok, "Transfer method should be an object")
	assert.Equal(t, "SWIFT", method["methodName"])
	assert.Equal(t, "1-3 days", method["processingTime"])
	assert.Equal(t, 100.0, method["minimumAmount"])

	// Verify original snake_case fields are not present
	assert.NotContains(t, data, "country_code")
	assert.NotContains(t, data, "bank_name")
	assert.NotContains(t, data, "bank_code")
	assert.NotContains(t, data, "created_at")
	assert.NotContains(t, data, "updated_at")
	assert.NotContains(t, method, "method_name")
	assert.NotContains(t, method, "processing_time")
	assert.NotContains(t, method, "minimum_amount")
}
