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

func TestCheck(t *testing.T) {
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

	validRequest := &amlcommon.CheckRequest{
		Name:        "John Doe",
		ReferenceID: "ref-123",
	}

	validResponse := &amlcommon.CheckResponse{
		Code:            "SUCCESS",
		Message:         "Request successful",
		TransactionID:   "txn-123",
		PricingStrategy: "standard",
		Data: amlcommon.CheckResponseData{
			TransID: "trans-123",
			Status:  "approved",
		},
	}

	testCases := []struct {
		desc         string
		wantErr      bool
		mocker       func(m *Mocker)
		request      *amlcommon.CheckRequest
		expcResponse *amlcommon.CheckResponse
	}{
		{
			desc:    "error request check",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil, 0, assert.AnError)
			},
			request:      validRequest,
			expcResponse: nil,
		},
		{
			desc:    "error validate response",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`INVALID JSON`), 200, nil)
			},
			request:      validRequest,
			expcResponse: nil,
		},
		{
			desc:    "HTTP error with valid response body",
			wantErr: false, // We expect no error because we can still process the response
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "Request successful",
					"transactionId": "txn-123",
					"pricingStrategy": "standard",
					"data": {
						"transId": "trans-123",
						"status": "approved"
					}
				}`), 400, assert.AnError) // Return an error with a valid response body
			},
			request:      validRequest,
			expcResponse: validResponse,
		},
		{
			desc:    "HTTP error with invalid response body",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`INVALID JSON`), 400, assert.AnError) // Return an error with an invalid response body
			},
			request:      validRequest,
			expcResponse: nil,
		},
		{
			desc:    "success when response is 200",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "Request successful",
					"transactionId": "txn-123",
					"pricingStrategy": "standard",
					"data": {
						"transId": "trans-123",
						"status": "approved"
					}
				}`), 200, nil)
			},
			request:      validRequest,
			expcResponse: validResponse,
		},
		{
			desc:    "success with review status",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "Request processed",
					"transactionId": "txn-456",
					"pricingStrategy": "premium",
					"data": {
						"transId": "trans-456",
						"status": "review"
					}
				}`), 200, nil)
			},
			request: validRequest,
			expcResponse: &amlcommon.CheckResponse{
				Code:            "SUCCESS",
				Message:         "Request processed",
				TransactionID:   "txn-456",
				PricingStrategy: "premium",
				Data: amlcommon.CheckResponseData{
					TransID: "trans-456",
					Status:  "review",
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

			resp, err := r.Check(context.Background(), tc.request)
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