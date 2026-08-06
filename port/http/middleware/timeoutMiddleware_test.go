package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicTimeout(t *testing.T) {
	_, pdkLogger, err := test.SetupLogger()
	require.NoError(t, err)

	tests := []struct {
		name           string
		path           string
		defaultTimeout int
		handler        http.Handler
		expectedStatus int
	}{
		{
			name:           "when path matches snap core event and request completes in time",
			path:           "/api/v1/disbursements/approval-actions",
			defaultTimeout: 5,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
				deadline, ok := r.Context().Deadline()
				assert.True(t, ok)
				assert.WithinDuration(t, time.Now().Add(600*time.Second), deadline, 5*time.Second)
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "when path matches snap core event but request times out",
			path:           "/api/v1/disbursements/single/retry",
			defaultTimeout: 5,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "when path does not match any special route",
			path:           "/api/v1/unknown-route",
			defaultTimeout: 2,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
				deadline, ok := r.Context().Deadline()
				assert.True(t, ok)
				assert.WithinDuration(t, time.Now(), deadline, 1*time.Second)
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "when path does not match any special route and times out",
			path:           "/api/v1/unknown-route",
			defaultTimeout: 2,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(3 * time.Second)
				r.Context().Done()
			}),
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "when path matches open API endpoint",
			path:           "/open-api/v1/inquiry-account",
			defaultTimeout: 5,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
				deadline, ok := r.Context().Deadline()
				assert.True(t, ok)
				assert.WithinDuration(t, time.Now().Add(600*time.Second), deadline, 5*time.Second)
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "when path matches internal API endpoint",
			path:           "/internal/v1/payments",
			defaultTimeout: 5,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
				deadline, ok := r.Context().Deadline()
				assert.True(t, ok)
				assert.WithinDuration(t, time.Now().Add(600*time.Second), deadline, 5*time.Second)
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply middleware
			middleware := DynamicTimeout(pdkLogger, tt.defaultTimeout)
			wrappedHandler := middleware(tt.handler)

			// Create test request
			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			assert.NoError(t, err)

			// Execute request
			recorder := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(recorder, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}
