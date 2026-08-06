package advanceairepository

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInquiry(t *testing.T) {
	type Mocker struct {
		cfg         *config.Config
		secret      *config.Secret
		httpRequest *httpMocks.IHTTPRequest
	}

	mockSecret := &config.Secret{
		AdvanceAISecret: config.AdvanceAISecret{
			ApiKey: "test-api-key",
		},
	}

	mockConfig := &config.Config{
		AdvanceAIConfig: config.AdvanceAIConfig{
			BaseURL:   "https://api.advance.ai",
			JourneyID: "journey-123",
		},
	}

	validTransactionID := "txn-123"

	validResponse := &amlcommon.InquiryResponse{
		Code:            "SUCCESS",
		Message:         "Inquiry successful",
		TransactionID:   "txn-123",
		PricingStrategy: "standard",
		Datetime:        1640995200.0,
		Timestamp:       1640995200,
		Data: amlcommon.InquiryResponseData{
			ID: "txn-123",
			Journey: amlcommon.Journey{
				ID:   1,
				Name: "AML Screening",
			},
			CustomerProfile: amlcommon.CustomerProfile{
				ID:   "customer-123",
				Name: "John Doe",
			},
			Nodes: []amlcommon.Node{
				{
					Type:               "screening",
					Name:               "PEP Screening",
					ID:                 1,
					Code:               stringPtr("SUCCESS"),
					Message:            stringPtr("Screening completed"),
					StartedAt:          "2023-01-01T00:00:00Z",
					CompletedAt:        "2023-01-01T00:01:00Z",
					Attributes:         map[string]interface{}{"score": float64(85)},
					Result:             &amlcommon.NodeResult{Detail: []amlcommon.NodeDetail{{Name: "Test", ProfileID: "test-profile"}}},
					VerificationResult: "approved",
				},
			},
		},
	}

	testCases := []struct {
		desc           string
		wantErr        bool
		mocker         func(m *Mocker)
		transactionID  string
		expcResponse   *amlcommon.InquiryResponse
	}{
		{
			desc:    "error request inquiry",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil, 0, assert.AnError)
			},
			transactionID: validTransactionID,
			expcResponse:  nil,
		},
		{
			desc:    "error validate response",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`INVALID JSON`), 200, nil)
			},
			transactionID: validTransactionID,
			expcResponse:  nil,
		},
		{
			desc:    "HTTP error with valid response body",
			wantErr: false, // We expect no error because we can still process the response
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "Inquiry successful",
					"transactionId": "txn-123",
					"pricingStrategy": "standard",
					"datetime": 1640995200.0,
					"timestamp": 1640995200,
					"data": {
						"id": "txn-123",
						"journey": {
							"id": 1,
							"name": "AML Screening"
						},
						"customerProfile": {
							"id": "customer-123",
							"name": "John Doe"
						},
						"nodes": [
							{
								"type": "screening",
								"name": "PEP Screening",
								"id": 1,
								"code": "SUCCESS",
								"message": "Screening completed",
								"startedAt": "2023-01-01T00:00:00Z",
								"completedAt": "2023-01-01T00:01:00Z",
								"attributes": {"score": 85},
								"result": {"detail": [{"name": "Test", "profileId": "test-profile"}]},
								"verificationResult": "approved"
							}
						]
					}
				}`), 404, assert.AnError) // Return an error with a valid response body
			},
			transactionID: validTransactionID,
			expcResponse:  validResponse,
		},
		{
			desc:    "HTTP error with invalid response body",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`INVALID JSON`), 404, assert.AnError) // Return an error with an invalid response body
			},
			transactionID: validTransactionID,
			expcResponse:  nil,
		},
		{
			desc:    "success when response is 200",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "Inquiry successful",
					"transactionId": "txn-123",
					"pricingStrategy": "standard",
					"datetime": 1640995200.0,
					"timestamp": 1640995200,
					"data": {
						"id": "txn-123",
						"journey": {
							"id": 1,
							"name": "AML Screening"
						},
						"customerProfile": {
							"id": "customer-123",
							"name": "John Doe"
						},
						"nodes": [
							{
								"type": "screening",
								"name": "PEP Screening",
								"id": 1,
								"code": "SUCCESS",
								"message": "Screening completed",
								"startedAt": "2023-01-01T00:00:00Z",
								"completedAt": "2023-01-01T00:01:00Z",
								"attributes": {"score": 85},
								"result": {"detail": [{"name": "Test", "profileId": "test-profile"}]},
								"verificationResult": "approved"
							}
						]
					}
				}`), 200, nil)
			},
			transactionID: validTransactionID,
			expcResponse:  validResponse,
		},
		{
			desc:    "success with complex data",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "Inquiry completed",
					"transactionId": "txn-789",
					"pricingStrategy": "premium",
					"datetime": 1640995300.5,
					"timestamp": 1640995300,
					"data": {
						"id": "txn-789",
						"journey": {
							"id": 2,
							"name": "Enhanced AML Screening"
						},
						"customerProfile": {
							"id": "customer-789",
							"name": "Jane Smith"
						},
						"nodes": [
							{
								"type": "screening",
								"name": "Sanctions Screening",
								"id": 1,
								"code": "SUCCESS",
								"message": "No matches found",
								"startedAt": "2023-01-01T10:00:00Z",
								"completedAt": "2023-01-01T10:05:00Z",
								"attributes": {"confidence": 95.5},
								"result": {"detail": []},
								"verificationResult": "clear"
							},
							{
								"type": "pep",
								"name": "PEP Screening",
								"id": 2,
								"startedAt": "2023-01-01T10:05:00Z",
								"completedAt": "2023-01-01T10:07:00Z",
								"attributes": {"database_version": "2023.1"},
								"result": {"detail": [{"name": "PEP Test", "profileId": "pep-profile"}]},
								"verificationResult": "review"
							}
						]
					}
				}`), 200, nil)
			},
			transactionID: "txn-789",
			expcResponse: &amlcommon.InquiryResponse{
				Code:            "SUCCESS",
				Message:         "Inquiry completed",
				TransactionID:   "txn-789",
				PricingStrategy: "premium",
				Datetime:        1640995300.5,
				Timestamp:       1640995300,
				Data: amlcommon.InquiryResponseData{
					ID: "txn-789",
					Journey: amlcommon.Journey{
						ID:   2,
						Name: "Enhanced AML Screening",
					},
					CustomerProfile: amlcommon.CustomerProfile{
						ID:   "customer-789",
						Name: "Jane Smith",
					},
					Nodes: []amlcommon.Node{
						{
							Type:               "screening",
							Name:               "Sanctions Screening",
							ID:                 1,
							Code:               stringPtr("SUCCESS"),
							Message:            stringPtr("No matches found"),
							StartedAt:          "2023-01-01T10:00:00Z",
							CompletedAt:        "2023-01-01T10:05:00Z",
							Attributes:         map[string]interface{}{"confidence": float64(95.5)},
							Result:             &amlcommon.NodeResult{Detail: []amlcommon.NodeDetail{}},
							VerificationResult: "clear",
						},
						{
							Type:               "pep",
							Name:               "PEP Screening",
							ID:                 2,
							StartedAt:          "2023-01-01T10:05:00Z",
							CompletedAt:        "2023-01-01T10:07:00Z",
							Attributes:         map[string]interface{}{"database_version": "2023.1"},
							Result:             &amlcommon.NodeResult{Detail: []amlcommon.NodeDetail{{Name: "PEP Test", ProfileID: "pep-profile"}}},
							VerificationResult: "review",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			m := &Mocker{
				cfg:    &config.Config{},
				secret: &config.Secret{},

				httpRequest: httpMocks.NewIHTTPRequest(t),
			}

			tc.mocker(m)

			r := New(m.cfg, m.secret, log, m.httpRequest)

			resp, err := r.Inquiry(context.Background(), tc.transactionID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expcResponse != nil {
				assert.Equal(t, tc.expcResponse, resp)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}