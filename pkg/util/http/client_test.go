package httputil_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get-test":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"method":"%s","status":"ok"}`, r.Method)

		case "/post-test":
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"method":"%s","status":"ok"}`, r.Method)

		case "/put-test":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"method":"%s","status":"ok"}`, r.Method)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{})

	tests := []struct {
		name           string
		method         string
		path           string
		body           any
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "Method GET",
			method:         http.MethodGet,
			path:           "/get-test",
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"method":"%s","status":"ok"}`, http.MethodGet),
		},
		{
			name:           "Method POST",
			method:         http.MethodPost,
			path:           "/post-test",
			body:           `{"message":"hello"}`,
			wantStatusCode: http.StatusCreated,
			wantRespBody:   fmt.Sprintf(`{"method":"%s","status":"ok"}`, http.MethodPost),
		},
		{
			name:           "Method PUT: Bytes body",
			method:         http.MethodPut,
			path:           "/put-test",
			body:           []byte(`{"message":"hello"}`),
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"method":"%s","status":"ok"}`, http.MethodPut),
		},
		{
			name:           "Method PUT: Map body",
			method:         http.MethodPut,
			path:           "/put-test",
			body:           map[string]string{"message": "hello"},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"method":"%s","status":"ok"}`, http.MethodPut),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, code, err := client.Request(t.Context(), tt.method, server.URL+tt.path, tt.body, nil)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, code)
			assert.JSONEq(t, tt.wantRespBody, string(body))
		})
	}
}

func TestHTTPClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{})
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	_, _, err := client.Request(ctx, http.MethodGet, server.URL+"/test", nil, nil)

	if assert.Error(t, err) {
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
	}
}

func TestHTTPClientReusedConnection(t *testing.T) {
	var count atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count.Load() > 2 {
			time.Sleep(200 * time.Millisecond)
		}
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"method":"%s","status":"ok"}`, r.Method)
	}))
	defer server.Close()

	buf := new(bytes.Buffer)
	defer func() { buf = nil }()

	logger := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	client := NewHTTPClient(HTTPClientConfig{}, WithLogger(logger))

	for range 4 {
		buf.Reset()
		count.Add(1)

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		_, _, err := client.Request(ctx, http.MethodGet, server.URL+"/test", nil, nil)

		switch count.Load() {
		case 1, 2:
			assert.NoError(t, err)

		case 3:
			assert.Error(t, err)
			assert.ErrorContains(t, err, "context deadline exceeded")
			assert.Contains(t, buf.String(), fmt.Sprintf(`"msg":"Failed to perform HTTP request","hostPort":"%s","reused":true,"wasIdle":true`, server.Listener.Addr()))

		case 4:
			assert.Error(t, err)
			assert.ErrorContains(t, err, "context deadline exceeded")
			assert.Contains(t, buf.String(), fmt.Sprintf(`"msg":"Failed to perform HTTP request","hostPort":"%s","reused":false,"wasIdle":false`, server.Listener.Addr()))
		}
	}
}
