package inboundService

import (
	"context"
	"testing"

	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	inboundPdk "github.com/paper-indonesia/pdk/v2/chiExt/inbound"
	"github.com/paper-indonesia/pdk/v2/logger"
	pdkEncoder "github.com/paper-indonesia/pdk/v2/logger/encoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRecord(t *testing.T) {
	testCases := []struct {
		name    string
		input   inboundPdk.HttpInbound
		setup   func(repo *mocksRepo.IInboundRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: record inbound with client",
			input: inboundPdk.HttpInbound{
				IP:             "127.0.0.1",
				Method:         "POST",
				URL:            "/open-api/v2/payments",
				Headers:        map[string][]string{"Content-Type": {"application/json"}},
				Body:           map[string]any{"key": "value"},
				StatusCode:     200,
				ResponseTimeMs: 100.5,
				Client: &inboundPdk.HttpClientDetails{
					Feature:     "PAYMENT",
					OriginId:    "origin-123",
					ReferenceId: "ref-123",
				},
				TraceId: "trace-123",
				ResponseBody: map[string]any{
					"data": map[string]any{
						"id": "payment-id-123",
					},
				},
			},
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("Insert", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: skip recording when client is nil",
			input: inboundPdk.HttpInbound{
				IP:         "127.0.0.1",
				Method:     "POST",
				URL:        "/open-api/v2/payments",
				StatusCode: 200,
				Client:     nil,
			},
			setup:   func(repo *mocksRepo.IInboundRepository) {},
			wantErr: false,
		},
		{
			name: "ERROR: repository insert fails",
			input: inboundPdk.HttpInbound{
				IP:             "127.0.0.1",
				Method:         "POST",
				URL:            "/open-api/v2/payments",
				Headers:        map[string][]string{"Content-Type": {"application/json"}},
				Body:           map[string]any{"key": "value"},
				StatusCode:     200,
				ResponseTimeMs: 100.5,
				Client: &inboundPdk.HttpClientDetails{
					Feature:     "PAYMENT",
					OriginId:    "origin-123",
					ReferenceId: "ref-123",
				},
				TraceId: "trace-123",
				ResponseBody: map[string]any{
					"data": map[string]any{},
				},
			},
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("Insert", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: record with response body containing uuid instead of id",
			input: inboundPdk.HttpInbound{
				IP:             "127.0.0.1",
				Method:         "GET",
				URL:            "/open-api/v2/payments",
				Headers:        map[string][]string{},
				StatusCode:     200,
				ResponseTimeMs: 50.0,
				Client: &inboundPdk.HttpClientDetails{
					Feature:     "PAYMENT",
					OriginId:    "",
					ReferenceId: "ref-456",
				},
				TraceId: "trace-456",
				ResponseBody: map[string]any{
					"data": map[string]any{
						"uuid": "uuid-from-response",
					},
				},
			},
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("Insert", mock.Anything, mock.Anything).Return(nil).Run(
					func(args mock.Arguments) {
						req := args.Get(1).(*inboundModel.InboundRequest)
						require.Equal(t, "uuid-from-response", req.Client.OriginId)
					},
				)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIInboundRepository(t)
			log, _ := logger.NewZapLogger(logger.Config{})
			inspector := pdkEncoder.NewInspector(nil)

			tc.setup(repo)

			svc := New(nil, log, repo, WithMaskingSensitiveData(nil))
			svc.(*InboundService).inspector = inspector

			err := svc.Record(context.Background(), tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetOriginIDFromResponse(t *testing.T) {
	testCases := []struct {
		name         string
		responseBody []byte
		expectedID   string
	}{
		{
			name:         "returns id when data.id is present",
			responseBody: []byte(`{"data":{"id":"payment-123","uuid":"uuid-456"}}`),
			expectedID:   "payment-123",
		},
		{
			name:         "returns uuid when data.id is empty but data.uuid is present",
			responseBody: []byte(`{"data":{"id":"","uuid":"uuid-456"}}`),
			expectedID:   "uuid-456",
		},
		{
			name:         "returns empty when both id and uuid are empty",
			responseBody: []byte(`{"data":{"id":"","uuid":""}}`),
			expectedID:   "",
		},
		{
			name:         "returns empty when data is missing",
			responseBody: []byte(`{"other":"value"}`),
			expectedID:   "",
		},
		{
			name:         "returns empty for invalid json",
			responseBody: []byte(`invalid json`),
			expectedID:   "",
		},
		{
			name:         "returns empty for nil response body",
			responseBody: nil,
			expectedID:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getOriginIDFromResponse(tc.responseBody)
			require.Equal(t, tc.expectedID, result)
		})
	}
}

func TestRecordInsertRequestFields(t *testing.T) {
	t.Run("verify inserted request fields are mapped correctly from input", func(t *testing.T) {
		repo := mocksRepo.NewIInboundRepository(t)
		log, _ := logger.NewZapLogger(logger.Config{})

		input := inboundPdk.HttpInbound{
			IP:             "10.0.0.1",
			Method:         "POST",
			URL:            "/open-api/v2/payments",
			Headers:        map[string][]string{"Authorization": {"Bearer token"}},
			Body:           `{"amount":10000}`,
			StatusCode:     201,
			ResponseTimeMs: 200.0,
			Client: &inboundPdk.HttpClientDetails{
				Feature:     "PAYMENT",
				OriginId:    "original-origin",
				ReferenceId: "ref-789",
			},
			TraceId: "trace-789",
			ResponseBody: map[string]any{
				"data": map[string]any{
					"id": "overridden-origin-id",
				},
			},
		}

		repo.On("Insert", mock.Anything, mock.Anything).Return(nil).Run(
			func(args mock.Arguments) {
				req := args.Get(1).(*inboundModel.InboundRequest)
				require.NotEmpty(t, req.ID, "ID should be generated")
				require.Equal(t, input.IP, req.IP)
				require.Equal(t, input.Method, req.Method)
				require.Equal(t, input.URL, req.URL)
				require.Equal(t, input.StatusCode, req.StatusCode)
				require.Equal(t, input.ResponseTimeMs, req.ResponseTimeMs)
				require.Equal(t, input.Client.Feature, req.Client.Feature)
				require.Equal(t, input.Client.ReferenceId, req.Client.ReferenceId)
				require.Equal(t, "overridden-origin-id", req.Client.OriginId, "OriginId should be overridden from response")
			},
		)

		svc := New(nil, log, repo, WithMaskingSensitiveData(nil))
		err := svc.Record(context.Background(), input)
		require.NoError(t, err)
	})
}
