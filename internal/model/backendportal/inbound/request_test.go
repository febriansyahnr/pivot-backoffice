package inboundModel_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/inbound"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/unifiedPayment"

	"github.com/stretchr/testify/assert"
)

func TestToInbound(t *testing.T) {
	client := &inboundModel.Client{
		Feature:     "payment",
		TraceId:     "trace-123",
		OriginId:    "origin-456",
		ReferenceId: "ref-789",
	}

	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}

	body := map[string]interface{}{
		"amount":   1000,
		"currency": "IDR",
	}

	metadata := map[string]interface{}{
		"platform": "web",
	}

	responseBody := []byte(`{"status":"success"}`)

	request := &inboundModel.InboundRequest{
		ID:                "req-123",
		Client:            client,
		IP:                "192.168.1.1",
		Method:            "POST",
		URL:               "https://api.example.com/pay",
		Headers:           headers,
		Body:              body,
		StatusCode:        200,
		ResponseTimeMs:    123.45,
		ResponseBody:      responseBody,
		SnapCompatibility: true,
		Metadata:          metadata,
	}

	inbound := request.ToInbound()

	assert.Equal(t, request.ID, inbound.ID)
	assert.Equal(t, request.IP, inbound.IP)
	assert.Equal(t, request.Method, inbound.Method)
	assert.Equal(t, request.URL, inbound.URL)
	assert.Equal(t, request.StatusCode, inbound.StatusCode)
	assert.Equal(t, request.ResponseTimeMs, inbound.ResponseTimeMs)
	assert.Equal(t, request.SnapCompatibility, inbound.SnapCompatibility)
	assert.WithinDuration(t, time.Now().UTC(), inbound.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), inbound.UpdatedAt, time.Second)

	var unmarshalledClient inboundModel.Client
	_ = json.Unmarshal(inbound.Client, &unmarshalledClient)
	assert.Equal(t, *client, unmarshalledClient)

	var unmarshalledHeaders map[string][]string
	_ = json.Unmarshal(inbound.Headers, &unmarshalledHeaders)
	assert.Equal(t, headers, unmarshalledHeaders)

	assert.True(t, inbound.Metadata.Valid)
	var unmarshalledMetadata map[string]interface{}
	_ = json.Unmarshal(inbound.Metadata.JSONText, &unmarshalledMetadata)
	assert.Equal(t, metadata, unmarshalledMetadata)

	assert.True(t, inbound.ResponseBody.Valid)
	assert.Equal(t, responseBody, []byte(inbound.ResponseBody.JSONText.String()))
}

func TestInboundRequest_SetSnapCompatible(t *testing.T) {
	// Create valid response body for payment tests
	chargeDetails := []*unifiedPaymentModel.ChargeResponse{
		{
			ID:     "charge-123",
			Status: "SUCCESS",
		},
	}
	unifiedResponse := unifiedPaymentModel.UnifiedPaymentSessionResponse{
		ChargeDetails: chargeDetails,
	}
	data, _ := json.Marshal(unifiedResponse)
	responseData := map[string]json.RawMessage{
		"data": data,
	}
	validResponseBody, _ := json.Marshal(responseData)

	// Create invalid response bodies
	invalidJSONResponseBody := []byte(`{"data": invalid json}`)
	emptyChargeDetailsBody, _ := json.Marshal(map[string]json.RawMessage{
		"data": json.RawMessage(`{"chargeDetails": []}`),
	})

	tests := []struct {
		name           string
		request        inboundModel.InboundRequest
		expectedResult bool
	}{
		{
			name: "Should set SnapCompatibility to true when URL is /internal/v1/access-token/b2b",
			request: inboundModel.InboundRequest{
				URL: "/internal/v1/access-token/b2b",
			},
			expectedResult: true,
		},
		{
			name: "Should not set SnapCompatibility to true for other URLs",
			request: inboundModel.InboundRequest{
				URL: "/some-other-endpoint",
			},
			expectedResult: false,
		},
		{
			name: "Should set SnapCompatibility for POST payment request with valid response",
			request: inboundModel.InboundRequest{
				Method:       http.MethodPost,
				URL:          "/open-api/v2/payments",
				ResponseBody: validResponseBody,
			},
			expectedResult: true,
		},
		{
			name: "Should set SnapCompatibility for payment confirm with valid response",
			request: inboundModel.InboundRequest{
				Method:       http.MethodGet,
				URL:          "/open-api/v2/payments/a1b2c3d4-e5f6-7890-abcd-ef1234567890/confirm",
				ResponseBody: validResponseBody,
			},
			expectedResult: true,
		},
		{
			name: "Should not set SnapCompatibility when not payment request or confirm",
			request: inboundModel.InboundRequest{
				Method: http.MethodGet,
				URL:    "/open-api/v2/some-other-endpoint",
			},
			expectedResult: false,
		},
		{
			name: "Should not set SnapCompatibility when response body is invalid JSON",
			request: inboundModel.InboundRequest{
				Method:       http.MethodPost,
				URL:          "/open-api/v2/payments",
				ResponseBody: invalidJSONResponseBody,
			},
			expectedResult: false,
		},
		{
			name: "Should not set SnapCompatibility when charge details are empty",
			request: inboundModel.InboundRequest{
				Method:       http.MethodPost,
				URL:          "/open-api/v2/payments",
				ResponseBody: emptyChargeDetailsBody,
			},
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a copy of the request to test
			req := tc.request

			// Call the method
			req.SetSnapCompatible()

			// Check the result
			assert.Equal(t, tc.expectedResult, req.SnapCompatibility)
		})
	}
}
