package fraudnetrepository

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	type Mocker struct {
		cfg         *config.Config
		secret      *config.Secret
		httpRequest *httpMocks.IHTTPRequest
		rabbitMq    *rabbitMqMocks.RabbitMQExt
	}

	mockSecret := &config.Secret{
		FraudNetSecret: config.FraudNetSecret{
			AccessKey:    "username",
			AccessSecret: "password",
		},
	}

	mockConfig := &config.Config{
		FraudNetConfig: config.FraudNetConfig{
			BaseURL: "http://localhost",
		},
	}

	validResponse := fdscommon.UpdateResponse{
		Success: true,
		Data:    fdscommon.UpdateData{},
	}

	validRequest := fdscommon.UpdateRequest{}

	testCases := []struct {
		desc         string
		wantErr      bool
		mocker       func(m *Mocker)
		request      *fdscommon.UpdateRequest
		expcResponse *fdscommon.UpdateResponse
	}{
		{
			desc:    "error request check",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"PATCH",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil, 0, assert.AnError)
			},
			request:      &validRequest,
			expcResponse: nil,
		},
		{
			desc:    "error validate response",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"PATCH",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`ERROR`), 200, nil)
			},
			request:      &validRequest,
			expcResponse: nil,
		},
		{
			desc:    "HTTP error with valid response body",
			wantErr: false, // We expect no error because we can still process the response
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"PATCH",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"success": true,
					"data": {}
				}`), 400, assert.AnError) // Return an error with a valid response body
			},
			request: &validRequest,
			expcResponse: &fdscommon.UpdateResponse{
				Success: true,
				Data:    fdscommon.UpdateData{},
			},
		},
		{
			desc:    "HTTP error with invalid response body",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"PATCH",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`INVALID JSON`), 400, assert.AnError) // Return an error with an invalid response body
			},
			request:      &validRequest,
			expcResponse: nil,
		},
		{
			desc:    "success when response is 200",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"PATCH",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"success": true
				}`), 200, nil)
			},
			request:      &validRequest,
			expcResponse: &validResponse,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			m := &Mocker{
				cfg:    &config.Config{},
				secret: &config.Secret{},

				httpRequest: httpMocks.NewIHTTPRequest(t),
				rabbitMq:    rabbitMqMocks.NewRabbitMQExt(t),
			}

			tc.mocker(m)

			r := New(m.cfg, m.secret, log, m.httpRequest)

			resp, err := r.Update(context.Background(), tc.request)
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
